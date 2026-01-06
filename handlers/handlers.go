package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/uberswe/awesometarkov.com/models"
	"github.com/uberswe/awesometarkov.com/parser"
)

// BaseURL is the canonical base URL for the site
const BaseURL = "https://www.awesometarkov.com"

// Meta contains SEO metadata for templates
type Meta struct {
	Title        string
	Description  string
	CanonicalURL string
	OGType       string
	OGImage      string
	PageType     string
}

// Breadcrumb represents a single breadcrumb item for navigation and JSON-LD
type Breadcrumb struct {
	Name string
	URL  string
}

// buildMeta creates a Meta struct with the given parameters
func buildMeta(title, description, path, pageType, ogType string) Meta {
	return Meta{
		Title:        title,
		Description:  description,
		CanonicalURL: BaseURL + path,
		OGType:       ogType,
		OGImage:      BaseURL + "/og" + path + ".png",
		PageType:     pageType,
	}
}

// buildHomeOGImage returns the OG image URL for the home page
func buildHomeOGImage() string {
	return BaseURL + "/og/home.png"
}

// buildSearchOGImage returns the OG image URL for the search page
func buildSearchOGImage() string {
	return BaseURL + "/og/search.png"
}

// Handler holds the application dependencies
type Handler struct {
	Data        *models.SiteData
	Templates   *template.Template
	TemplateMap map[string]*template.Template
}

// NewHandler creates a new handler with the given data and templates
func NewHandler(data *models.SiteData, templates *template.Template) *Handler {
	return &Handler{
		Data:      data,
		Templates: templates,
	}
}

// NewHandlerWithTemplateMap creates a new handler with a template map for base template inheritance
func NewHandlerWithTemplateMap(data *models.SiteData, templateMap map[string]*template.Template) *Handler {
	return &Handler{
		Data:        data,
		TemplateMap: templateMap,
	}
}

// executeTemplate executes the appropriate template based on handler configuration
func (h *Handler) executeTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	if h.TemplateMap != nil {
		tmpl, ok := h.TemplateMap[templateName]
		if !ok {
			return fmt.Errorf("template %s not found", templateName)
		}
		return tmpl.ExecuteTemplate(w, templateName, data)
	}
	return h.Templates.ExecuteTemplate(w, templateName, data)
}

// HomeData contains data for the home page template
type HomeData struct {
	Meta           Meta
	Query          string // For search form in header
	Categories     []models.Category
	TotalResources int
}

// CategoryData contains data for the category page template
type CategoryData struct {
	Meta           Meta
	Query          string // For search form in header
	Breadcrumbs    []Breadcrumb
	Category       *models.Category
	Categories     []models.Category
	TotalResources int
}

// SearchData contains data for the search results page template
type SearchData struct {
	Meta           Meta
	Query          string
	Results        []models.SearchResult
	ResultCount    int
	Categories     []models.Category
	TotalResources int
}

// ResourceData contains data for the resource page template
type ResourceData struct {
	Meta           Meta
	Query          string // For search form in header
	Breadcrumbs    []Breadcrumb
	Resource       *models.Resource
	Categories     []models.Category
	TotalResources int
}

// RedirectData contains data for the external link redirect page
type RedirectData struct {
	URL          string
	ResourceName string
}

// Home handles the home page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := HomeData{
		Meta: Meta{
			Title:        "Awesome Tarkov - Curated Escape From Tarkov Resources",
			Description:  fmt.Sprintf("A curated collection of %d+ Escape From Tarkov resources including maps, ammo charts, quest trackers, and community tools.", h.Data.TotalResources),
			CanonicalURL: BaseURL + "/",
			OGType:       "website",
			OGImage:      buildHomeOGImage(),
			PageType:     "home",
		},
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Category handles category pages
func (h *Handler) Category(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path /category/{slug}
	path := strings.TrimPrefix(r.URL.Path, "/category/")
	slug := strings.TrimSuffix(path, "/")

	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	category := parser.GetCategoryBySlug(h.Data, slug)
	if category == nil {
		http.NotFound(w, r)
		return
	}

	// Count resources in category
	resourceCount := 0
	for _, sub := range category.Subcategories {
		resourceCount += len(sub.Resources)
	}

	// Build subcategory preview for description
	subcatPreview := ""
	for i, sub := range category.Subcategories {
		if i > 0 {
			subcatPreview += ", "
		}
		if i >= 3 {
			subcatPreview += "and more"
			break
		}
		subcatPreview += sub.Name
	}

	description := category.Description
	if description == "" {
		description = fmt.Sprintf("Browse %d %s resources for Escape From Tarkov including %s.", resourceCount, category.Name, subcatPreview)
	}

	data := CategoryData{
		Meta: Meta{
			Title:        fmt.Sprintf("%s - Awesome Tarkov", category.Name),
			Description:  description,
			CanonicalURL: BaseURL + "/category/" + slug,
			OGType:       "website",
			OGImage:      BaseURL + "/og/category/" + slug + ".png",
			PageType:     "category",
		},
		Breadcrumbs: []Breadcrumb{
			{Name: "Home", URL: BaseURL + "/"},
			{Name: category.Name, URL: BaseURL + "/category/" + slug},
		},
		Category:       category,
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "category.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Search handles search requests
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var results []models.SearchResult
	if query != "" {
		results = parser.Search(h.Data, query)
	}

	title := "Search - Awesome Tarkov"
	description := "Search Escape From Tarkov resources including maps, ammo charts, quest trackers, and more."
	if query != "" {
		title = fmt.Sprintf("Search: %s - Awesome Tarkov", query)
		description = fmt.Sprintf("Found %d Tarkov resources matching '%s'.", len(results), query)
	}

	data := SearchData{
		Meta: Meta{
			Title:        title,
			Description:  description,
			CanonicalURL: BaseURL + "/search",
			OGType:       "website",
			OGImage:      buildSearchOGImage(),
			PageType:     "search",
		},
		Query:          query,
		Results:        results,
		ResultCount:    len(results),
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "search.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Resource handles individual resource pages
func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	// Extract category and resource slugs from path /resource/{category}/{resource}
	path := strings.TrimPrefix(r.URL.Path, "/resource/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	categorySlug := parts[0]
	resourceSlug := parts[1]

	resource := parser.GetResourceBySlug(h.Data, categorySlug, resourceSlug)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	// Truncate description for meta if too long
	metaDesc := resource.Description
	if len(metaDesc) > 160 {
		metaDesc = metaDesc[:157] + "..."
	}

	resourcePath := "/resource/" + categorySlug + "/" + resourceSlug

	data := ResourceData{
		Meta: Meta{
			Title:        fmt.Sprintf("%s - %s | Awesome Tarkov", resource.Name, resource.CategoryName),
			Description:  metaDesc,
			CanonicalURL: BaseURL + resourcePath,
			OGType:       "article",
			OGImage:      BaseURL + "/og" + resourcePath + ".png",
			PageType:     "resource",
		},
		Breadcrumbs: []Breadcrumb{
			{Name: "Home", URL: BaseURL + "/"},
			{Name: resource.CategoryName, URL: BaseURL + "/category/" + categorySlug},
			{Name: resource.Name, URL: BaseURL + resourcePath},
		},
		Resource:       resource,
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "resource.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Redirect handles the external link redirect page
// Uses resource ID (category-slug/resource-slug) instead of arbitrary URLs for security
// Supports optional index for multiple links: /go/{category}/{resource}/{index}
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Extract category, resource slugs, and optional index from path
	path := strings.TrimPrefix(r.URL.Path, "/go/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || len(parts) > 3 {
		http.NotFound(w, r)
		return
	}

	categorySlug := parts[0]
	resourceSlug := parts[1]
	linkIndex := 0

	// Parse optional index
	if len(parts) == 3 {
		if idx, err := fmt.Sscanf(parts[2], "%d", &linkIndex); err != nil || idx != 1 {
			http.NotFound(w, r)
			return
		}
	}

	// Look up resource by ID - this prevents open redirect attacks
	resource := parser.GetResourceBySlug(h.Data, categorySlug, resourceSlug)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	// Determine which URL to use
	var targetURL string
	if len(resource.URLs) > 0 {
		// Use URLs slice if available
		if linkIndex < 0 || linkIndex >= len(resource.URLs) {
			http.NotFound(w, r)
			return
		}
		targetURL = resource.URLs[linkIndex].URL
	} else if resource.URL != "" {
		// Fall back to primary URL for backward compatibility
		if linkIndex != 0 {
			http.NotFound(w, r)
			return
		}
		targetURL = resource.URL
	} else {
		http.NotFound(w, r)
		return
	}

	data := RedirectData{
		URL:          targetURL,
		ResourceName: resource.Name,
	}

	err := h.executeTemplate(w, "redirect.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// LegalData contains data for legal pages (privacy, terms)
type LegalData struct {
	Meta           Meta
	Query          string // For search form in header
	Categories     []models.Category
	TotalResources int
}

// Privacy handles the privacy policy page
func (h *Handler) Privacy(w http.ResponseWriter, r *http.Request) {
	data := LegalData{
		Meta: Meta{
			Title:        "Privacy Policy - Awesome Tarkov",
			Description:  "Privacy Policy for Awesome Tarkov. Learn how we collect, use, and protect your information.",
			CanonicalURL: BaseURL + "/privacy",
			OGType:       "website",
			OGImage:      BaseURL + "/og/home.png",
			PageType:     "legal",
		},
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "privacy.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Terms handles the terms and conditions page
func (h *Handler) Terms(w http.ResponseWriter, r *http.Request) {
	data := LegalData{
		Meta: Meta{
			Title:        "Terms and Conditions - Awesome Tarkov",
			Description:  "Terms and Conditions for using Awesome Tarkov. Please read carefully before using our site.",
			CanonicalURL: BaseURL + "/terms",
			OGType:       "website",
			OGImage:      BaseURL + "/og/home.png",
			PageType:     "legal",
		},
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.executeTemplate(w, "terms.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
