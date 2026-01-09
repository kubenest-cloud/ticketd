package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/default_form.css static/admin/*
var staticFS embed.FS

type templateCache struct {
	pages map[string]*template.Template
}

func parseTemplates() (*templateCache, error) {
	funcs := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04")
		},
	}

	files, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	layoutPath := ""
	pages := make(map[string]*template.Template)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == "layout.html" {
			layoutPath = filepath.ToSlash("templates/" + file.Name())
			break
		}
	}
	if layoutPath == "" {
		return nil, fmt.Errorf("layout template not found")
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "layout.html" {
			continue
		}
		pagePath := filepath.ToSlash("templates/" + file.Name())
		tmpl, err := template.New("layout").Funcs(funcs).ParseFS(templateFS, layoutPath, pagePath)
		if err != nil {
			return nil, err
		}
		pages[file.Name()] = tmpl
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no page templates found")
	}
	return &templateCache{pages: pages}, nil
}

func defaultCSS() ([]byte, error) {
	return staticFS.ReadFile("static/default_form.css")
}

func adminAssets() (fs.FS, error) {
	return fs.Sub(staticFS, "static/admin")
}
