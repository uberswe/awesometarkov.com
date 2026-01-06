package handlers

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// URLSet represents the root sitemap element
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL represents a single URL entry in the sitemap
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// getBaseURL returns the base URL from environment or default
func getBaseURL() string {
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}
	return BaseURL
}

// Sitemap generates and serves the sitemap.xml
func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL()
	urls := []URL{}
	today := time.Now().Format("2006-01-02")

	// Add homepage
	urls = append(urls, URL{
		Loc:        baseURL + "/",
		LastMod:    today,
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// Add category pages
	for _, category := range h.Data.Categories {
		urls = append(urls, URL{
			Loc:        fmt.Sprintf("%s/category/%s", baseURL, category.Slug),
			LastMod:    today,
			ChangeFreq: "weekly",
			Priority:   0.8,
		})
	}

	// Add resource pages
	for _, category := range h.Data.Categories {
		for _, subcategory := range category.Subcategories {
			for _, resource := range subcategory.Resources {
				urls = append(urls, URL{
					Loc:        fmt.Sprintf("%s/resource/%s/%s", baseURL, category.Slug, resource.Slug),
					LastMod:    today,
					ChangeFreq: "monthly",
					Priority:   0.6,
				})
			}
		}
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlset); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Robots serves the robots.txt file
func (h *Handler) Robots(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL()
	robotsTxt := fmt.Sprintf(`# robots.txt for %s
User-agent: *
Allow: /
Disallow: /search
Disallow: /static/

# Sitemap location
Sitemap: %s/sitemap.xml
`, baseURL, baseURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(robotsTxt))
}

// SearchEnginePinger handles pinging search engines
type SearchEnginePinger struct {
	SitemapURL string
	Client     *http.Client
}

// NewSearchEnginePinger creates a new pinger instance
func NewSearchEnginePinger(sitemapURL string) *SearchEnginePinger {
	return &SearchEnginePinger{
		SitemapURL: sitemapURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PingAll pings all configured search engines
func (p *SearchEnginePinger) PingAll() {
	engines := []struct {
		Name    string
		PingURL string
	}{
		{"Google", "http://www.google.com/ping"},
		{"Bing", "http://www.bing.com/ping"},
	}

	for _, engine := range engines {
		p.ping(engine.Name, engine.PingURL)
	}
}

// ping sends a sitemap notification to a search engine
func (p *SearchEnginePinger) ping(name, pingURL string) {
	fullURL := fmt.Sprintf("%s?sitemap=%s", pingURL, url.QueryEscape(p.SitemapURL))

	resp, err := p.Client.Get(fullURL)
	if err != nil {
		log.Printf("Failed to ping %s: %v", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully pinged %s with sitemap", name)
	} else {
		log.Printf("Ping to %s returned status: %d", name, resp.StatusCode)
	}
}
