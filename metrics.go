package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	metricsNamespace = "goprove"
	pushInterval     = 15 * time.Second
	jobName          = "goprove_site"
)

// Metrics holds all Prometheus collectors and the optional GeoIP reader.
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	visitorsGeo     *prometheus.CounterVec
	buildInfo       *prometheus.GaugeVec
	registry        *prometheus.Registry
	geoDB           *geoip2.Reader
}

// newMetrics creates and registers all Prometheus metrics.
// geoDBPath may be empty if no GeoIP database is available.
func newMetrics(version string, geoDBPath string) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry: reg,
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests.",
		}, []string{"method", "path", "status"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),
		responseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes.",
			Buckets:   prometheus.ExponentialBuckets(256, 2, 10), // 256B .. 128KB
		}, []string{"path"}),
		visitorsGeo: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "http_requests_geo_total",
			Help:      "HTTP requests by visitor geolocation.",
		}, []string{"country", "city", "latitude", "longitude"}),
		buildInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "build_info",
			Help:      "Build information. Always 1.",
		}, []string{"version"}),
	}

	reg.MustRegister(m.requestsTotal, m.requestDuration, m.responseSize, m.visitorsGeo, m.buildInfo)

	m.buildInfo.WithLabelValues(version).Set(1)

	// Open GeoIP database if available
	if geoDBPath != "" {
		db, err := geoip2.Open(geoDBPath)
		if err != nil {
			log.Printf("warning: could not open GeoIP database %s: %v (geo tracking disabled)", geoDBPath, err)
		} else {
			m.geoDB = db
			log.Printf("GeoIP database loaded from %s", geoDBPath)
		}
	}

	return m
}

// Close releases the GeoIP database.
func (m *Metrics) Close() {
	if m.geoDB != nil {
		m.geoDB.Close()
	}
}

// Middleware returns an http.Handler that records metrics for each request.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rw.status)

		m.requestsTotal.WithLabelValues(r.Method, path, status).Inc()
		m.requestDuration.WithLabelValues(r.Method, path).Observe(duration)
		m.responseSize.WithLabelValues(path).Observe(float64(rw.size))

		m.recordGeo(r)
	})
}

// recordGeo looks up the client IP in the GeoIP database and increments the geo counter.
func (m *Metrics) recordGeo(r *http.Request) {
	if m.geoDB == nil {
		return
	}

	ip := clientIP(r)
	if ip == nil {
		return
	}

	city, err := m.geoDB.City(ip)
	if err != nil {
		return
	}

	country := city.Country.IsoCode
	cityName := city.City.Names["en"]
	lat := strconv.FormatFloat(city.Location.Latitude, 'f', 2, 64)
	lon := strconv.FormatFloat(city.Location.Longitude, 'f', 2, 64)

	if country == "" {
		return
	}

	m.visitorsGeo.WithLabelValues(country, cityName, lat, lon).Inc()
}

// StartPusher pushes metrics to the Pushgateway every pushInterval.
// It blocks forever; run it in a goroutine.
func (m *Metrics) StartPusher(gatewayURL string) {
	pusher := push.New(gatewayURL, jobName).Gatherer(m.registry)

	ticker := time.NewTicker(pushInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := pusher.Add(); err != nil {
			log.Printf("error pushing metrics: %v", err)
		}
	}
}

// clientIP extracts the real client IP from headers or RemoteAddr.
func clientIP(r *http.Request) net.IP {
	// X-Real-IP (set by nginx/reverse proxy)
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return net.ParseIP(strings.TrimSpace(ip))
	}

	// X-Forwarded-For (first entry is the client)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return net.ParseIP(strings.TrimSpace(parts[0]))
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}

// normalizePath collapses dynamic-looking paths to reduce label cardinality.
// For this site, all paths are static, so we just clean up the path.
func normalizePath(path string) string {
	if strings.HasPrefix(path, "/static/") {
		return "/static/*"
	}
	return path
}

// responseWriter wraps http.ResponseWriter to capture status code and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	// Ignore 1xx informational (e.g. 103 Early Hints) — only capture final status
	if code >= 200 {
		rw.status = code
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
