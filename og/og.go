package og

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
)

const (
	Width  = 1200
	Height = 630
)

// Font sizes in points
const (
	FontHuge   = 72 // Main titles
	FontLarge  = 48 // Subtitles
	FontMedium = 36 // Body text
	FontSmall  = 28 // Labels
	FontTiny   = 22 // Branding
)

// Font path - use system Arial or fallback
var fontPath = "/Library/Fonts/Arial Unicode.ttf"

// Tarkov-themed color scheme - military greens and amber accents
var (
	ColorBgStart     = hexToRGBA("#0a0a0a")
	ColorBgEnd       = hexToRGBA("#1a1a1a")
	ColorAccent      = hexToRGBA("#c49a3c") // Tarkov amber/gold
	ColorTextPrimary = hexToRGBA("#e5e5e5")
	ColorTextMuted   = hexToRGBA("#737373")
	ColorCardBg      = hexToRGBA("#1e1e1e")
	ColorMilitary    = hexToRGBA("#4a5d23") // Military green for decorative elements
)

// Cache stores generated images
type Cache struct {
	mu     sync.RWMutex
	images map[string][]byte
}

// NewCache creates a new image cache
func NewCache() *Cache {
	return &Cache{
		images: make(map[string][]byte),
	}
}

// Get retrieves a cached image
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.images[key]
	return data, ok
}

// Set stores an image in the cache
func (c *Cache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.images[key] = data
}

// Generator creates OG images
type Generator struct {
	Cache *Cache
}

// NewGenerator creates a new OG image generator
func NewGenerator() *Generator {
	return &Generator{
		Cache: NewCache(),
	}
}

// hexToRGBA converts a hex color string to color.RGBA
func hexToRGBA(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.RGBA{0, 0, 0, 255}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// drawGradientBackground draws a vertical gradient background
func drawGradientBackground(dc *gg.Context) {
	for y := 0; y < Height; y++ {
		t := float64(y) / float64(Height)
		r := lerp(float64(ColorBgStart.R), float64(ColorBgEnd.R), t)
		g := lerp(float64(ColorBgStart.G), float64(ColorBgEnd.G), t)
		b := lerp(float64(ColorBgStart.B), float64(ColorBgEnd.B), t)
		dc.SetRGB(r/255, g/255, b/255)
		dc.DrawLine(0, float64(y), float64(Width), float64(y))
		dc.Stroke()
	}
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// drawBottomAccent draws the amber accent line at the bottom
func drawBottomAccent(dc *gg.Context) {
	dc.SetColor(ColorAccent)
	dc.DrawRectangle(0, float64(Height-6), float64(Width), 6)
	dc.Fill()
}

// drawMilitaryStripes draws decorative military-style stripes in corners
func drawMilitaryStripes(dc *gg.Context) {
	dc.SetColor(ColorMilitary)
	dc.SetLineWidth(3)

	// Top left corner stripes
	for i := 0; i < 4; i++ {
		offset := float64(i * 15)
		dc.DrawLine(0, 30+offset, 30+offset, 0)
	}
	dc.Stroke()

	// Top right corner stripes
	for i := 0; i < 4; i++ {
		offset := float64(i * 15)
		dc.DrawLine(float64(Width)-30-offset, 0, float64(Width), 30+offset)
	}
	dc.Stroke()
}

// loadFont loads the font at the specified size
func loadFont(dc *gg.Context, size float64) {
	if err := dc.LoadFontFace(fontPath, size); err != nil {
		// Fallback - the text just won't render well
		return
	}
}

// drawTextCentered draws centered text with color and font size
func drawTextCentered(dc *gg.Context, text string, y float64, col color.RGBA, fontSize float64) {
	dc.SetColor(col)
	loadFont(dc, fontSize)
	dc.DrawStringAnchored(text, float64(Width)/2, y, 0.5, 0.5)
}

// drawTextLeft draws left-aligned text with color and font size
func drawTextLeft(dc *gg.Context, text string, x, y float64, col color.RGBA, fontSize float64) {
	dc.SetColor(col)
	loadFont(dc, fontSize)
	dc.DrawString(text, x, y)
}

// drawStatsBox draws a statistics box with text
func drawStatsBox(dc *gg.Context, x, y, width, height float64, value, label string) {
	// Box background
	dc.SetColor(ColorCardBg)
	dc.DrawRoundedRectangle(x, y, width, height, 12)
	dc.Fill()

	// Border
	dc.SetColor(ColorAccent)
	dc.SetLineWidth(3)
	dc.DrawRoundedRectangle(x, y, width, height, 12)
	dc.Stroke()

	// Value - large text
	dc.SetColor(ColorTextPrimary)
	loadFont(dc, FontMedium)
	dc.DrawStringAnchored(value, x+width/2, y+height*0.4, 0.5, 0.5)

	// Label - smaller text
	dc.SetColor(ColorTextMuted)
	loadFont(dc, FontSmall)
	dc.DrawStringAnchored(label, x+width/2, y+height*0.75, 0.5, 0.5)
}

// drawStatsBoxLarge draws a larger statistics box with bigger text
func drawStatsBoxLarge(dc *gg.Context, x, y, width, height float64, value, label string) {
	// Box background
	dc.SetColor(ColorCardBg)
	dc.DrawRoundedRectangle(x, y, width, height, 12)
	dc.Fill()

	// Border
	dc.SetColor(ColorAccent)
	dc.SetLineWidth(4)
	dc.DrawRoundedRectangle(x, y, width, height, 12)
	dc.Stroke()

	// Value - large text
	dc.SetColor(ColorTextPrimary)
	loadFont(dc, FontLarge)
	dc.DrawStringAnchored(value, x+width/2, y+height*0.4, 0.5, 0.5)

	// Label - medium text
	dc.SetColor(ColorTextMuted)
	loadFont(dc, FontMedium)
	dc.DrawStringAnchored(label, x+width/2, y+height*0.75, 0.5, 0.5)
}

// GenerateHome creates the home page OG image
func (g *Generator) GenerateHome(totalResources, categoryCount int) image.Image {
	dc := gg.NewContext(Width, Height)

	// Background
	drawGradientBackground(dc)

	// Draw decorative military stripes
	drawMilitaryStripes(dc)

	// Draw a decorative grid pattern
	dc.SetRGBA(1, 1, 1, 0.03)
	for i := 0; i < Width; i += 40 {
		dc.DrawLine(float64(i), 0, float64(i), float64(Height))
	}
	for i := 0; i < Height; i += 40 {
		dc.DrawLine(0, float64(i), float64(Width), float64(i))
	}
	dc.SetLineWidth(1)
	dc.Stroke()

	// Title - AWESOME TARKOV
	drawTextCentered(dc, "AWESOME TARKOV", 140, ColorAccent, FontHuge)

	// Subtitle
	drawTextCentered(dc, "Escape From Tarkov Resources", 220, ColorTextPrimary, FontLarge)

	// Stats boxes
	boxWidth := 300.0
	boxHeight := 100.0
	boxY := 300.0
	gap := 40.0
	totalWidth := 3*boxWidth + 2*gap
	startX := (float64(Width) - totalWidth) / 2

	drawStatsBox(dc, startX, boxY, boxWidth, boxHeight, strconv.Itoa(totalResources)+"+", "Resources")
	drawStatsBox(dc, startX+boxWidth+gap, boxY, boxWidth, boxHeight, strconv.Itoa(categoryCount), "Categories")
	drawStatsBox(dc, startX+2*(boxWidth+gap), boxY, boxWidth, boxHeight, "Community", "Driven")

	// Bottom accent
	drawBottomAccent(dc)

	// Branding
	drawTextCentered(dc, "awesometarkov.com", float64(Height)-35, ColorTextMuted, FontTiny)

	return dc.Image()
}

// GenerateCategory creates a category page OG image
func (g *Generator) GenerateCategory(name string, resourceCount int, subcategories []string) image.Image {
	dc := gg.NewContext(Width, Height)

	// Background
	drawGradientBackground(dc)

	// Draw decorative military stripes
	drawMilitaryStripes(dc)

	// Top branding
	drawTextLeft(dc, "awesometarkov.com", 60, 55, ColorTextMuted, FontSmall)

	// Category name
	drawTextCentered(dc, name, 220, ColorAccent, FontHuge)

	// Resource count
	countText := strconv.Itoa(resourceCount) + " Resources"
	drawTextCentered(dc, countText, 320, ColorTextPrimary, FontLarge)

	// Subcategory preview
	if len(subcategories) > 0 {
		maxShow := 3
		if len(subcategories) < maxShow {
			maxShow = len(subcategories)
		}
		preview := strings.Join(subcategories[:maxShow], "  •  ")
		if len(subcategories) > 3 {
			preview += "  •  ..."
		}
		drawTextCentered(dc, preview, 420, ColorTextMuted, FontMedium)
	}

	// Bottom accent
	drawBottomAccent(dc)

	// Branding
	drawTextCentered(dc, "awesometarkov.com", float64(Height)-40, ColorTextMuted, FontSmall)

	return dc.Image()
}

// GenerateResource creates a resource page OG image
func (g *Generator) GenerateResource(name, description, category, platform, audience, price string) image.Image {
	dc := gg.NewContext(Width, Height)

	// Background
	drawGradientBackground(dc)

	// Draw decorative military stripes
	drawMilitaryStripes(dc)

	// Breadcrumb - category path
	drawTextLeft(dc, category, 60, 55, ColorTextMuted, FontSmall)

	// Resource name - truncate if needed
	if len(name) > 30 {
		name = name[:27] + "..."
	}
	drawTextCentered(dc, name, 150, ColorAccent, FontHuge)

	// Description - multi-line wrapped text
	dc.SetColor(ColorTextPrimary)
	loadFont(dc, FontSmall)
	maxWidth := float64(Width) - 120 // 60px padding on each side
	dc.DrawStringWrapped(description, float64(Width)/2, 240, 0.5, 0, maxWidth, 1.4, gg.AlignCenter)

	// Metadata badges - larger boxes
	boxWidth := 300.0
	boxHeight := 120.0
	boxY := 360.0
	gap := 40.0
	totalWidth := 3*boxWidth + 2*gap
	startX := (float64(Width) - totalWidth) / 2

	drawStatsBoxLarge(dc, startX, boxY, boxWidth, boxHeight, platform, "Platform")
	drawStatsBoxLarge(dc, startX+boxWidth+gap, boxY, boxWidth, boxHeight, audience, "Audience")
	drawStatsBoxLarge(dc, startX+2*(boxWidth+gap), boxY, boxWidth, boxHeight, price, "Price")

	// Bottom accent
	drawBottomAccent(dc)

	// Branding
	drawTextCentered(dc, "awesometarkov.com", float64(Height)-40, ColorTextMuted, FontSmall)

	return dc.Image()
}

// GenerateSearch creates the search page OG image
func (g *Generator) GenerateSearch(totalResources, categoryCount int) image.Image {
	dc := gg.NewContext(Width, Height)

	// Background
	drawGradientBackground(dc)

	// Draw decorative military stripes
	drawMilitaryStripes(dc)

	// Search icon (magnifying glass) - larger
	dc.SetColor(ColorAccent)
	dc.DrawCircle(float64(Width)/2, 130, 60)
	dc.SetLineWidth(8)
	dc.Stroke()
	dc.DrawLine(float64(Width)/2+45, 175, float64(Width)/2+85, 215)
	dc.Stroke()

	// Title
	drawTextCentered(dc, "Search Resources", 300, ColorAccent, FontHuge)

	// Subtitle
	drawTextCentered(dc, "Find maps, ammo charts, trackers, and more", 390, ColorTextPrimary, FontLarge)

	// Stats
	statsText := strconv.Itoa(totalResources) + "+ Resources  •  " + strconv.Itoa(categoryCount) + " Categories"
	drawTextCentered(dc, statsText, 470, ColorTextMuted, FontMedium)

	// Bottom accent
	drawBottomAccent(dc)

	// Branding
	drawTextCentered(dc, "awesometarkov.com", float64(Height)-40, ColorTextMuted, FontSmall)

	return dc.Image()
}

// Cache key helpers
func HomeKey() string {
	return "og:home"
}

func SearchKey() string {
	return "og:search"
}

func CategoryKey(slug string) string {
	return "og:category:" + slug
}

func ResourceKey(categorySlug, resourceSlug string) string {
	return "og:resource:" + categorySlug + "/" + resourceSlug
}
