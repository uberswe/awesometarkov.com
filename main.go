package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/uberswe/awesometarkov.com/handlers"
	"github.com/uberswe/awesometarkov.com/og"
	"github.com/uberswe/awesometarkov.com/parser"
)

func main() {
	// Parse resources from individual markdown files
	data, err := parser.ParseResourcesDir("resources")
	if err != nil {
		log.Fatalf("Failed to parse resources directory: %v", err)
	}
	log.Printf("Loaded %d categories with %d total resources", len(data.Categories), data.TotalResources)

	// Parse templates with custom functions
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse base template and partials first
	baseTemplates := template.New("").Funcs(funcMap)
	baseTemplates, err = baseTemplates.ParseGlob(filepath.Join("templates", "partials", "*.html"))
	if err != nil {
		log.Fatalf("Failed to parse partial templates: %v", err)
	}
	baseTemplates, err = baseTemplates.ParseFiles(filepath.Join("templates", "base.html"))
	if err != nil {
		log.Fatalf("Failed to parse base template: %v", err)
	}

	// Create a template map for each page by cloning base and adding page-specific content
	templateMap := make(map[string]*template.Template)
	pageTemplates := []string{"home.html", "category.html", "resource.html", "search.html", "privacy.html", "terms.html"}

	for _, page := range pageTemplates {
		// Clone the base templates
		pageTemplate, err := baseTemplates.Clone()
		if err != nil {
			log.Fatalf("Failed to clone base templates for %s: %v", page, err)
		}

		// Parse the page-specific template
		pageTemplate, err = pageTemplate.ParseFiles(filepath.Join("templates", page))
		if err != nil {
			log.Fatalf("Failed to parse page template %s: %v", page, err)
		}

		templateMap[page] = pageTemplate
	}

	// Parse standalone templates (don't use base template)
	redirectTmpl, err := template.New("redirect.html").Funcs(funcMap).ParseFiles(filepath.Join("templates", "redirect.html"))
	if err != nil {
		log.Fatalf("Failed to parse redirect template: %v", err)
	}
	templateMap["redirect.html"] = redirectTmpl

	// Create handler with template map
	h := handlers.NewHandlerWithTemplateMap(data, templateMap)

	// Create OG image generator and handler
	ogGenerator := og.NewGenerator()
	ogHandler := handlers.NewOGHandler(ogGenerator, h)

	// Set up routes
	http.HandleFunc("/", h.Home)
	http.HandleFunc("/category/", h.Category)
	http.HandleFunc("/resource/", h.Resource)
	http.HandleFunc("/search", h.Search)
	http.HandleFunc("/go/", h.Redirect)
	http.HandleFunc("/privacy", h.Privacy)
	http.HandleFunc("/terms", h.Terms)

	// SEO routes
	http.HandleFunc("/sitemap.xml", h.Sitemap)
	http.HandleFunc("/robots.txt", h.Robots)

	// OG image routes
	http.HandleFunc("/og/home.png", ogHandler.OGHome)
	http.HandleFunc("/og/search.png", ogHandler.OGSearch)
	http.HandleFunc("/og/category/", ogHandler.OGCategory)
	http.HandleFunc("/og/resource/", ogHandler.OGResource)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Ping search engines on startup (in background)
	go func() {
		pinger := handlers.NewSearchEnginePinger("https://www.awesometarkov.com/sitemap.xml")
		pinger.PingAll()
	}()

	// Start server
	port := ":8082"
	fmt.Printf("Server starting at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
