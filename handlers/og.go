package handlers

import (
	"bytes"
	"image/png"
	"net/http"
	"strings"

	"github.com/uberswe/awesometarkov.com/models"
	"github.com/uberswe/awesometarkov.com/og"
)

// OGHandler handles OG image generation requests
type OGHandler struct {
	Generator *og.Generator
	Handler   *Handler
}

// NewOGHandler creates a new OG image handler
func NewOGHandler(generator *og.Generator, handler *Handler) *OGHandler {
	return &OGHandler{
		Generator: generator,
		Handler:   handler,
	}
}

// OGHome serves the home page OG image
func (o *OGHandler) OGHome(w http.ResponseWriter, r *http.Request) {
	cacheKey := og.HomeKey()

	// Check cache
	if data, ok := o.Generator.Cache.Get(cacheKey); ok {
		o.serveImage(w, data)
		return
	}

	// Count total resources
	totalResources := 0
	for _, cat := range o.Handler.Data.Categories {
		for _, sub := range cat.Subcategories {
			totalResources += len(sub.Resources)
		}
	}

	// Generate image
	img := o.Generator.GenerateHome(totalResources, len(o.Handler.Data.Categories))

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	data := buf.Bytes()
	o.Generator.Cache.Set(cacheKey, data)
	o.serveImage(w, data)
}

// OGSearch serves the search page OG image
func (o *OGHandler) OGSearch(w http.ResponseWriter, r *http.Request) {
	cacheKey := og.SearchKey()

	// Check cache
	if data, ok := o.Generator.Cache.Get(cacheKey); ok {
		o.serveImage(w, data)
		return
	}

	// Count total resources
	totalResources := 0
	for _, cat := range o.Handler.Data.Categories {
		for _, sub := range cat.Subcategories {
			totalResources += len(sub.Resources)
		}
	}

	// Generate image
	img := o.Generator.GenerateSearch(totalResources, len(o.Handler.Data.Categories))

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	data := buf.Bytes()
	o.Generator.Cache.Set(cacheKey, data)
	o.serveImage(w, data)
}

// OGCategory serves category page OG images
func (o *OGHandler) OGCategory(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path: /og/category/{slug}.png
	path := strings.TrimPrefix(r.URL.Path, "/og/category/")
	slug := strings.TrimSuffix(path, ".png")

	if slug == "" {
		http.NotFound(w, r)
		return
	}

	cacheKey := og.CategoryKey(slug)

	// Check cache
	if data, ok := o.Generator.Cache.Get(cacheKey); ok {
		o.serveImage(w, data)
		return
	}

	// Find category
	var category *models.Category
	for i := range o.Handler.Data.Categories {
		if o.Handler.Data.Categories[i].Slug == slug {
			category = &o.Handler.Data.Categories[i]
			break
		}
	}

	if category == nil {
		http.NotFound(w, r)
		return
	}

	// Count resources
	resourceCount := 0
	var subcategoryNames []string
	for _, sub := range category.Subcategories {
		resourceCount += len(sub.Resources)
		subcategoryNames = append(subcategoryNames, sub.Name)
	}

	// Generate image
	img := o.Generator.GenerateCategory(category.Name, resourceCount, subcategoryNames)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	data := buf.Bytes()
	o.Generator.Cache.Set(cacheKey, data)
	o.serveImage(w, data)
}

// OGResource serves resource page OG images
func (o *OGHandler) OGResource(w http.ResponseWriter, r *http.Request) {
	// Extract path: /og/resource/{category}/{resource}.png
	path := strings.TrimPrefix(r.URL.Path, "/og/resource/")
	path = strings.TrimSuffix(path, ".png")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	categorySlug := parts[0]
	resourceSlug := parts[1]

	cacheKey := og.ResourceKey(categorySlug, resourceSlug)

	// Check cache
	if data, ok := o.Generator.Cache.Get(cacheKey); ok {
		o.serveImage(w, data)
		return
	}

	// Find category and resource
	var category *models.Category
	var resource *models.Resource

	for i := range o.Handler.Data.Categories {
		if o.Handler.Data.Categories[i].Slug == categorySlug {
			category = &o.Handler.Data.Categories[i]
			for j := range category.Subcategories {
				for k := range category.Subcategories[j].Resources {
					if category.Subcategories[j].Resources[k].Slug == resourceSlug {
						resource = &category.Subcategories[j].Resources[k]
						break
					}
				}
				if resource != nil {
					break
				}
			}
			break
		}
	}

	if category == nil || resource == nil {
		http.NotFound(w, r)
		return
	}

	// Generate image
	img := o.Generator.GenerateResource(
		resource.Name,
		resource.Description,
		category.Name,
		resource.Platform,
		resource.Audience,
		resource.Price,
	)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	data := buf.Bytes()
	o.Generator.Cache.Set(cacheKey, data)
	o.serveImage(w, data)
}

// serveImage writes the image data with appropriate headers
func (o *OGHandler) serveImage(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}
