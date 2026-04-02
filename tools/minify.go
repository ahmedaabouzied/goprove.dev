//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

func main() {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	if err := minifyFile(m, "static/style.css", "static/style.min.css", "text/css"); err != nil {
		fmt.Fprintf(os.Stderr, "minify css: %v\n", err)
		os.Exit(1)
	}

	if err := minifyFile(m, "static/app.js", "static/app.min.js", "application/javascript"); err != nil {
		fmt.Fprintf(os.Stderr, "minify js: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("minified style.css → style.min.css")
	fmt.Println("minified app.js    → app.min.js")
}

func minifyFile(m *minify.M, src, dst, mediaType string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	out, err := m.String(mediaType, string(data))
	if err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(out), 0644)
}
