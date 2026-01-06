package models

// ResourceLink represents a single link with optional label
type ResourceLink struct {
	URL   string
	Label string
}

// Resource represents a single Tarkov resource entry
type Resource struct {
	Name            string
	Slug            string
	URL             string         // Primary URL (first link) for backward compatibility
	URLs            []ResourceLink // All links for this resource
	Description     string
	Platform        string
	Audience        string
	Price           string
	CategorySlug    string
	SubcategorySlug string
	CategoryName    string
	SubcategoryName string
}

// Subcategory represents a subcategory within a main category
type Subcategory struct {
	Name      string
	Slug      string
	Resources []Resource
}

// Category represents a main category of resources
type Category struct {
	Name          string
	Slug          string
	Description   string
	Subcategories []Subcategory
}

// SearchResult represents a resource with its category context for search results
type SearchResult struct {
	Resource    Resource
	Category    string
	Subcategory string
}

// SiteData holds all parsed data from the markdown file
type SiteData struct {
	Categories     []Category
	TotalResources int
}
