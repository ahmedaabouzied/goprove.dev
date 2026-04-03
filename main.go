package main

import (
	"bytes"
	"embed"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

// Version is set at build time via -ldflags "-X main.Version=<commit>".
var Version = "dev"

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed content/*
var contentFS embed.FS

// Page represents a documentation page parsed from markdown.
type Page struct {
	Slug        string
	Title       string
	Description string
	Section     string
	Order       int
	Content     template.HTML
	RawContent  string
	Headings    []Heading
}

// Heading represents a heading extracted from the rendered HTML for TOC.
type Heading struct {
	Level int
	ID    string
	Text  string
}

// Frontmatter is the YAML frontmatter in each markdown file.
type Frontmatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Section     string `yaml:"section"`
	Order       int    `yaml:"order"`
}

// SitemapURL is a single URL entry in sitemap.xml.
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap is the root sitemap.xml structure.
type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// Site holds all loaded pages and templates.
type Site struct {
	pages     []*Page
	pageMap   map[string]*Page
	homeTmpl  *template.Template
	docTmpl   *template.Template
	md        goldmark.Markdown
	baseURL   string
	buildTime string
	version   string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(Version)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	site, err := newSite("https://goprove.dev")
	if err != nil {
		log.Fatal(err)
	}

	// Metrics setup — configure via PUSHGATEWAY and GEOIP_DB env vars
	pushGW := os.Getenv("PUSHGATEWAY")
	geoDBPath := getEnvOrDefault("GEOIP_DB", "/opt/goprove.dev/GeoLite2-City.mmdb")
	metrics := newMetrics(Version, geoDBPath)
	defer metrics.Close()

	if pushGW != "" {
		go metrics.StartPusher(pushGW)
		log.Printf("Metrics pushing to %s every 15s", pushGW)
	} else {
		log.Println("PUSHGATEWAY not set — metrics disabled")
	}

	mux := http.NewServeMux()

	// Static files — long cache lifetime is safe because URLs are cache-busted with ?v=<version>
	mux.Handle("/static/", staticHandler(http.FileServer(http.FS(staticFS))))

	// Auto-generated SEO/LLM files
	mux.HandleFunc("/sitemap.xml", site.handleSitemap)
	mux.HandleFunc("/robots.txt", site.handleRobots)
	mux.HandleFunc("/llms.txt", site.handleLLMsTxt)
	mux.HandleFunc("/llms-full.txt", site.handleLLMsFullTxt)

	// Doc pages
	for _, p := range site.pages {
		page := p
		mux.HandleFunc("/"+page.Slug, site.withEarlyHints(site.handleDoc(page)))
	}

	// Home
	mux.HandleFunc("/", site.withEarlyHints(site.handleHome))

	log.Printf("goprove.dev version=%s listening on :%s", Version, port)
	log.Fatal(http.ListenAndServe(":"+port, metrics.Middleware(mux)))
}

func newSite(baseURL string) (*Site, error) {
	s := &Site{
		pageMap:   make(map[string]*Page),
		baseURL:   baseURL,
		buildTime: time.Now().Format("2006-01-02"),
		version:   Version,
	}

	// Set up goldmark with syntax highlighting and frontmatter
	s.md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// Parse templates — separate sets so define blocks don't collide
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"lower":    strings.ToLower,
		"hasPrefix": strings.HasPrefix,
	}

	homeTmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/home.html")
	if err != nil {
		return nil, fmt.Errorf("parsing home templates: %w", err)
	}
	s.homeTmpl = homeTmpl

	docTmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/doc.html")
	if err != nil {
		return nil, fmt.Errorf("parsing doc templates: %w", err)
	}
	s.docTmpl = docTmpl

	// Load content
	entries, err := contentFS.ReadDir("content")
	if err != nil {
		return nil, fmt.Errorf("reading content dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := contentFS.ReadFile(filepath.Join("content", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		page, err := s.parsePage(entry.Name(), data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		s.pages = append(s.pages, page)
		s.pageMap[page.Slug] = page
	}

	// Sort by order
	sort.Slice(s.pages, func(i, j int) bool {
		return s.pages[i].Order < s.pages[j].Order
	})

	log.Printf("Loaded %d pages", len(s.pages))
	return s, nil
}

func (s *Site) parsePage(filename string, data []byte) (*Page, error) {
	// Parse markdown with frontmatter
	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := s.md.Convert(data, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}

	// Extract frontmatter
	fm := &Frontmatter{}
	d := frontmatter.Get(ctx)
	if d != nil {
		if err := d.Decode(fm); err != nil {
			return nil, fmt.Errorf("decoding frontmatter: %w", err)
		}
	}

	slug := strings.TrimSuffix(filename, ".md")
	rendered := buf.String()

	return &Page{
		Slug:        slug,
		Title:       fm.Title,
		Description: fm.Description,
		Section:     fm.Section,
		Order:       fm.Order,
		Content:     template.HTML(rendered),
		RawContent:  stripHTML(rendered),
		Headings:    extractHeadings(rendered),
	}, nil
}

// extractHeadings parses rendered HTML for h2/h3 tags to build a TOC.
func extractHeadings(html string) []Heading {
	var headings []Heading
	for _, level := range []int{2, 3} {
		tag := fmt.Sprintf("h%d", level)
		rest := html
		for {
			openTag := "<" + tag
			idx := strings.Index(rest, openTag)
			if idx == -1 {
				break
			}
			rest = rest[idx:]

			// Find id attribute
			idIdx := strings.Index(rest, `id="`)
			closeIdx := strings.Index(rest, ">")
			if idIdx == -1 || idIdx > closeIdx {
				rest = rest[closeIdx+1:]
				continue
			}

			idStart := idIdx + 4
			idEnd := strings.Index(rest[idStart:], `"`)
			id := rest[idStart : idStart+idEnd]

			// Find text content
			textStart := closeIdx + 1
			textEnd := strings.Index(rest[textStart:], "</"+tag+">")
			text := stripHTML(rest[textStart : textStart+textEnd])

			headings = append(headings, Heading{Level: level, ID: id, Text: text})
			rest = rest[textStart+textEnd:]
		}
	}

	// Re-sort by position (since we searched by level)
	// Simple approach: search again in order
	var ordered []Heading
	rest := html
	for {
		bestIdx := -1
		var bestHeading Heading
		for _, h := range headings {
			tag := fmt.Sprintf(`<h%d id="%s"`, h.Level, h.ID)
			idx := strings.Index(rest, tag)
			if idx != -1 && (bestIdx == -1 || idx < bestIdx) {
				bestIdx = idx
				bestHeading = h
			}
		}
		if bestIdx == -1 {
			break
		}
		ordered = append(ordered, bestHeading)
		rest = rest[bestIdx+1:]
	}

	return ordered
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var buf strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

type homeData struct {
	Pages   []*Page
	Version string
}

type docData struct {
	Page     *Page
	Pages    []*Page
	NextPage *Page
	Version  string
}

func (s *Site) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var buf bytes.Buffer
	err := s.homeTmpl.ExecuteTemplate(&buf, "base", homeData{Pages: s.pages, Version: s.version})
	if err != nil {
		log.Printf("error rendering home: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (s *Site) handleDoc(page *Page) http.HandlerFunc {
	// Find next page in order
	var nextPage *Page
	for i, p := range s.pages {
		if p.Slug == page.Slug && i+1 < len(s.pages) {
			nextPage = s.pages[i+1]
			break
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		err := s.docTmpl.ExecuteTemplate(&buf, "base", docData{Page: page, Pages: s.pages, NextPage: nextPage, Version: s.version})
		if err != nil {
			log.Printf("error rendering %s: %v", page.Slug, err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		buf.WriteTo(w)
	}
}

func (s *Site) handleSitemap(w http.ResponseWriter, r *http.Request) {
	sm := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []SitemapURL{
			{Loc: s.baseURL + "/", LastMod: s.buildTime, Priority: "1.0", ChangeFreq: "weekly"},
		},
	}
	for _, p := range s.pages {
		sm.URLs = append(sm.URLs, SitemapURL{
			Loc:        s.baseURL + "/" + p.Slug,
			LastMod:    s.buildTime,
			Priority:   "0.8",
			ChangeFreq: "weekly",
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.Encode(sm)
}

func (s *Site) handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nDisallow:\n\nSitemap: %s/sitemap.xml\n", s.baseURL)
}

func staticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

// withEarlyHints sends a 103 Early Hints response with preload links for the
// critical CSS and font before the full 200 response. Nginx 1.29+ will forward
// the 103 to the browser; older proxies drop it but the Link headers still
// appear in the 200 and trigger browser prefetch.
func (s *Site) withEarlyHints(next http.HandlerFunc) http.HandlerFunc {
	cssLink := `</static/style.min.css?v=` + s.version + `>; rel=preload; as=style`
	fontLink := `</static/fonts/jetbrains-mono-normal-latin.woff2>; rel=preload; as=font; crossorigin`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Link", cssLink)
		w.Header().Add("Link", fontLink)
		w.WriteHeader(http.StatusEarlyHints)
		next(w, r)
	}
}

func (s *Site) handleLLMsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, `# GoProve

> A static analysis tool for Go that uses abstract interpretation to mathematically prove properties about your code.

GoProve detects nil pointer dereferences, division by zero, and integer overflow in Go programs. When it says Error, it's mathematically proven. When it produces no output, safety is guaranteed. When it's unsure, it tells you honestly (Warning).

## Pages

`)
	for _, p := range s.pages {
		fmt.Fprintf(w, "- [%s](%s/%s): %s\n", p.Title, s.baseURL, p.Slug, p.Description)
	}
	fmt.Fprintf(w, "\n## Full content\n\nSee %s/llms-full.txt for complete documentation.\n", s.baseURL)
}

func (s *Site) handleLLMsFullTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "# GoProve — Full Documentation\n\n")
	for _, p := range s.pages {
		fmt.Fprintf(w, "---\n\n# %s\n\n%s\n\n", p.Title, p.RawContent)
	}
}
