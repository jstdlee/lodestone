package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// keroWashes are the kero themes Lodestone carries, as kero's settings
// name them (theme = light:mist, dark:paper): kero/theme.go's two washes
// with their tone.
type keroWash struct {
	light, dark []string
	tone        keroTone
}

// keroTone is a kero theme's tone: the hue, how white the light paper is
// (tint), how strong the programs' colors stand (vivid) and the hue of
// the accent, the cursor, selection and controls.
type keroTone struct{ hue, tint, vivid, accent float32 }

var keroWashes = map[string]keroWash{
	// Sea above, blue-gray fading to warm paper below.
	"mist": {
		light: []string{"#e3e9f4", "#ebeff6", "#f1f2f3", "#f2f1ec"},
		dark:  []string{"#0c1628", "#0d1827", "#111b23"},
		tone:  keroTone{255, 1, 0.8, 255},
	},
	// White on a desk of near-neutral dark, flat and quiet.
	"paper": {
		light: []string{"#f6f6f7", "#f7f7f8", "#f8f8f8"},
		dark:  []string{"#141416", "#151517", "#161618"},
		tone:  keroTone{260, 0, 0.9, 262},
	},
}

// lodestoneTheme is Lodestone's chrome, kero's as the host terminal means
// it: mist when light, paper when dark. Hue stays with the categories and
// the blocks they mark; the surfaces and controls carry the looks.
func lodestoneTheme(dark bool) *ui.Theme {
	if dark {
		return keroLook("paper", true)
	}
	return keroLook("mist", false)
}

// keroLook colors the chrome from one kero wash, kero's tinted() doing
// the chart: the surfaces are the theme card, the text its foreground,
// the accent its cursor color, and the control faces and hairlines what
// kero's palettes draw.
func keroLook(id string, dark bool) *ui.Theme {
	w, ok := keroWashes[id]
	if !ok {
		w = keroWashes["paper"]
		id = "paper"
	}
	hexes := w.light
	if dark {
		hexes = w.dark
	}
	wash := ui.Hex(hexes[len(hexes)/2])
	card, fg, accent, selection := keroToneColors(w.tone, dark)

	if dark {
		t := ui.DarkTheme()
		t.Background = wash
		t.Surface = card
		t.SurfaceHover = card.Mix(fg, 0.05)
		t.SurfacePressed = card.Mix(fg, 0.11)
		t.Border = ui.RGBA(255, 255, 255, 0.08)
		t.Text = fg
		t.TextMuted = fg.Mix(card, 0.38)
		t.Accent = accent
		t.AccentHover = accent.Mix(fg, 0.12)
		t.AccentPressed = accent.Mix(fg, 0.22)
		t.AccentText = card
		t.Danger = ui.Hex("#ff6961")
		t.Warning = ui.Hex("#f0a33a")
		t.Success = ui.Hex("#4fd17a")
		t.Selection = selection
		t.Focus = accent.Alpha(0.45)
		t.Scrollbar = fg.Alpha(0.28)
		t.Radius = 8
		t.InactiveAccent = t.Accent
		t.TooltipDelay = 400 * time.Millisecond
		return t
	}
	t := ui.LightTheme()
	t.Background = wash
	t.Surface = card
	t.SurfaceHover = card.Mix(fg, 0.04)
	t.SurfacePressed = card.Mix(fg, 0.07)
	t.Border = ui.RGBA(0, 0, 0, 0.07)
	t.Text = fg
	t.TextMuted = fg.Mix(card, 0.38)
	t.Accent = accent
	t.AccentHover = accent.Mix(fg, 0.10)
	t.AccentPressed = accent.Mix(fg, 0.20)
	t.AccentText = ui.Hex("#ffffff")
	t.Danger = ui.Hex("#d93a35")
	t.Warning = ui.Hex("#c27b0f")
	t.Success = ui.Hex("#2f8a3f")
	t.Selection = selection
	t.Focus = accent.Alpha(0.35)
	t.Scrollbar = fg.Alpha(0.28)
	t.Radius = 8
	t.InactiveAccent = t.Accent
	t.TooltipDelay = 400 * time.Millisecond
	return t
}

// keroToneColors is a tone's card, card text, accent and selection, kero's
// toneTheme's beginnings: the card tinted with the hue, in the light the
// whiter the less tinted, in the dark a deep tinted ground.
func keroToneColors(t keroTone, dark bool) (card, fg, accent, selection ui.Color) {
	tint := func(l, c float32) ui.Color { return ui.Oklch(l, c*t.tint, t.hue) }
	if dark {
		return tint(0.255, 0.02), tint(0.92, 0.015), ui.Oklch(0.74, 0.13, t.accent), ui.Oklch(0.38, 0.06, t.accent)
	}
	paper := 0.978 + 0.022*max(0, 1-t.tint/0.5)
	return tint(paper, 0.012), tint(0.28, 0.03), ui.Oklch(0.55, 0.13, t.accent), ui.Oklch(0.9, 0.045, t.accent)
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
