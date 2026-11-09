package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// lodestoneTheme is Lodestone's mineral survey chrome: graphite surfaces,
// hairlines, and a monochrome accent, so hue stays with the categories and
// the blocks they mark.
func lodestoneTheme(dark bool) *ui.Theme {
	if dark {
		t := ui.DarkTheme()
		t.Background = ui.Hex("#101011")
		t.Surface = ui.Hex("#19191b")
		t.SurfaceHover = ui.Hex("#232325")
		t.SurfacePressed = ui.Hex("#2c2c2e")
		t.Border = ui.RGBA(237, 237, 237, 0.09)
		t.Text = ui.Hex("#ededed")
		t.TextMuted = ui.Hex("#a1a1a6")
		t.Accent = ui.Hex("#ededed")
		t.AccentHover = ui.Hex("#ffffff")
		t.AccentPressed = ui.Hex("#c8c8cc")
		t.AccentText = ui.Hex("#171717")
		t.Danger = ui.Hex("#e5484d")
		t.Warning = ui.Hex("#e8a33d")
		t.Success = ui.Hex("#4cc38a")
		t.Selection = ui.RGBA(237, 237, 237, 0.12)
		t.Focus = ui.RGBA(237, 237, 237, 0.45)
		t.Scrollbar = ui.RGBA(237, 237, 237, 0.28)
		t.Radius = 8
		t.InactiveAccent = t.Accent
		t.TooltipDelay = 400 * time.Millisecond
		return t
	}
	t := ui.LightTheme()
	t.Background = ui.Hex("#f0f0f2")
	t.Surface = ui.Hex("#ffffff")
	t.SurfaceHover = ui.Hex("#f4f4f5")
	t.SurfacePressed = ui.Hex("#e4e4e7")
	t.Border = ui.RGBA(23, 23, 23, 0.09)
	t.Text = ui.Hex("#171717")
	t.TextMuted = ui.Hex("#737373")
	t.Accent = ui.Hex("#26262b")
	t.AccentHover = ui.Hex("#17171a")
	t.AccentPressed = ui.Hex("#3f3f46")
	t.AccentText = ui.Hex("#ffffff")
	t.Danger = ui.Hex("#dc2626")
	t.Warning = ui.Hex("#b45309")
	t.Success = ui.Hex("#16a34a")
	t.Selection = ui.RGBA(23, 23, 23, 0.08)
	t.Focus = ui.RGBA(23, 23, 23, 0.35)
	t.Scrollbar = ui.RGBA(23, 23, 23, 0.28)
	t.Radius = 8
	t.InactiveAccent = t.Accent
	t.TooltipDelay = 400 * time.Millisecond
	return t
}

func categoryColor(cat string) ui.Color {
	a, _ := categoryPair(cat)
	return a
}

// categoryPair is the two hues a category's blocks mix. The first is the
// chip color; the second keeps neighboring blocks from looking identical.
func categoryPair(cat string) (ui.Color, ui.Color) {
	switch cat {
	case "Hub":
		return ui.Hex("#3b6cf6"), ui.Hex("#22d3ee")
	case "Runners":
		return ui.Hex("#0f8199"), ui.Hex("#34d399")
	case "Image & audio":
		return ui.Hex("#a033c4"), ui.Hex("#f472b6")
	case "Training":
		return ui.Hex("#22863a"), ui.Hex("#a3e635")
	case "Engines":
		return ui.Hex("#c27b0f"), ui.Hex("#fb923c")
	case "Runtime":
		return ui.Hex("#5b6fd6"), ui.Hex("#c084fc")
	case "Custom":
		return ui.Hex("#0e7c66"), ui.Hex("#2dd4bf")
	case "Loose weights":
		return ui.Hex("#d93a35"), ui.Hex("#fb7185")
	case "Large files":
		return ui.Hex("#5c6370"), ui.Hex("#94a3b8")
	default:
		return ui.Hex("#6e6e78"), ui.Hex("#a1a1aa")
	}
}

// blockGradient is one treemap or gallery block: a diagonal mix of the
// category's hues, shifted by index so neighbors are not the same fill.
func blockGradient(cat string, other bool, index int) ui.LinearGradient {
	a, b := categoryPair(cat)
	if other {
		a, b = ui.Hex("#8b93a1"), ui.Hex("#64748b")
	}
	slot := index % 4
	t := float32(slot) / 3
	light := a.MixOklab(ui.Hex("#ffffff"), 0.36-0.08*t)
	mid := a.MixOklab(b, 0.28+0.5*t)
	deep := b.MixOklab(ui.Hex("#12141a"), 0.36)
	return ui.LinearGradient{
		Angle: 150 + float32(slot)*12,
		Oklab: true,
		Stops: []ui.GradientStop{
			{Color: light, At: 0},
			{Color: mid, At: 0.48},
			{Color: deep, At: 1},
		},
	}
}

// shadeGradient dims or shines a block. glow below 0 leaves it alone.
// 0 is the dim end of a selection pulse, 1 the shine. dimmed darkens a
// block that is not the one selected.
func shadeGradient(g ui.LinearGradient, glow float32, dimmed bool) ui.LinearGradient {
	src := g.Stops
	if len(src) == 0 {
		src = []ui.GradientStop{{Color: g.From, At: 0}, {Color: g.To, At: 1}}
	}
	g.Stops = make([]ui.GradientStop, len(src))
	copy(g.Stops, src)
	n := len(g.Stops)
	for i := range g.Stops {
		g.Stops[i].Color = shadeStop(g.Stops[i].Color, glow, dimmed, i, n)
	}
	g.From = g.Stops[0].Color
	g.To = g.Stops[n-1].Color
	return g
}

func shadeStop(c ui.Color, glow float32, dimmed bool, i, n int) ui.Color {
	span := float32(1)
	if n > 1 {
		span = float32(i) / float32(n-1)
	}
	night := ui.Hex("#14161c")
	if dimmed {
		return c.MixOklab(night, 0.36+0.16*span)
	}
	if glow < 0 {
		return c
	}
	dim := c.MixOklab(night, 0.5)
	hot := c.MixOklab(ui.Hex("#ffffff"), 0.62-0.28*span)
	return dim.MixOklab(hot, glow)
}

func inkOn(c ui.Color) ui.Color {
	y := 0.2126*float32(c.R) + 0.7152*float32(c.G) + 0.0722*float32(c.B)
	if y > 150 {
		return ui.Hex("#1d1d1f")
	}
	return ui.Hex("#ffffff")
}

const folderSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>`

const copySVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>`

const refreshSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/><path d="M8 16H3v5"/></svg>`

const sunSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="M4.93 4.93l1.41 1.41"/><path d="M17.66 17.66l1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="M6.34 17.66l-1.41 1.41"/><path d="M19.07 4.93l-1.41 1.41"/></svg>`

const moonSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>`

const listSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 6h13"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M3 6h.01"/><path d="M3 12h.01"/><path d="M3 18h.01"/></svg>`

const backSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12 19-7-7 7-7"/><path d="M19 12H5"/></svg>`

const plusSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>`

const checkSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>`

const xSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="M6 6l12 12"/></svg>`

const trashSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><path d="M10 11v6"/><path d="M14 11v6"/></svg>`

const duplicateSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/><path d="M15 11v6"/><path d="M12 14h6"/></svg>`

const placesSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/><path d="M12 10v6"/><path d="M9 13h6"/></svg>`

const gallerySVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="7" height="7" x="3" y="3" rx="1"/><rect width="7" height="7" x="14" y="3" rx="1"/><rect width="7" height="7" x="3" y="14" rx="1"/><rect width="7" height="7" x="14" y="14" rx="1"/></svg>`

const mapSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="11" height="8" x="3" y="3" rx="1"/><rect width="5" height="8" x="16" y="3" rx="1"/><rect width="7" height="8" x="3" y="13" rx="1"/><rect width="9" height="8" x="12" y="13" rx="1"/></svg>`

var (
	iconFolder    = ui.MustParseSVG([]byte(folderSVG))
	iconCopy      = ui.MustParseSVG([]byte(copySVG))
	iconRefresh   = ui.MustParseSVG([]byte(refreshSVG))
	iconSun       = ui.MustParseSVG([]byte(sunSVG))
	iconMoon      = ui.MustParseSVG([]byte(moonSVG))
	iconList      = ui.MustParseSVG([]byte(listSVG))
	iconBack      = ui.MustParseSVG([]byte(backSVG))
	iconPlus      = ui.MustParseSVG([]byte(plusSVG))
	iconCheck     = ui.MustParseSVG([]byte(checkSVG))
	iconX         = ui.MustParseSVG([]byte(xSVG))
	iconTrash     = ui.MustParseSVG([]byte(trashSVG))
	iconDuplicate = ui.MustParseSVG([]byte(duplicateSVG))
	iconPlaces    = ui.MustParseSVG([]byte(placesSVG))
	iconGallery   = ui.MustParseSVG([]byte(gallerySVG))
	iconMap       = ui.MustParseSVG([]byte(mapSVG))
)
