package main

import "github.com/egoist/mygo/ui"

// keroTheme is Kero's paper chrome: cool gray wash, white cards, blue accent.
// Colors follow kero's lightPalette and darkPalette.
func keroTheme(dark bool) *ui.Theme {
	if dark {
		t := ui.DarkTheme()
		t.Background = ui.Hex("#141418")
		t.Surface = ui.Hex("#232328")
		t.SurfaceHover = ui.Hex("#2c2c32")
		t.SurfacePressed = ui.Hex("#404047")
		t.Border = ui.RGBA(255, 255, 255, 0.10)
		t.Text = ui.Hex("#ececf0")
		t.TextMuted = ui.Hex("#a9a9b4")
		t.Accent = ui.Hex("#6b93ff")
		t.AccentHover = ui.Hex("#8fb3ff")
		t.AccentPressed = ui.Hex("#3b6cf6")
		t.AccentText = ui.Hex("#161618")
		t.Danger = ui.Hex("#ff6961")
		t.Warning = ui.Hex("#f0a33a")
		t.Success = ui.Hex("#4fd17a")
		t.Selection = ui.RGBA(107, 147, 255, 0.35)
		t.Focus = ui.RGBA(107, 147, 255, 0.55)
		t.Scrollbar = ui.RGBA(255, 255, 255, 0.28)
		t.Radius = 8
		t.InactiveAccent = t.Accent
		return t
	}
	t := ui.LightTheme()
	t.Background = ui.Hex("#eceef4")
	t.Surface = ui.Hex("#ffffff")
	t.SurfaceHover = ui.Hex("#f7f7f9")
	t.SurfacePressed = ui.Hex("#eceef4")
	t.Border = ui.RGBA(0, 0, 0, 0.08)
	t.Text = ui.Hex("#1d1d1f")
	t.TextMuted = ui.Hex("#6e6e78")
	t.Accent = ui.Hex("#3b6cf6")
	t.AccentHover = ui.Hex("#2f5bd8")
	t.AccentPressed = ui.Hex("#2f5bd8")
	t.AccentText = ui.Hex("#ffffff")
	t.Danger = ui.Hex("#d93a35")
	t.Warning = ui.Hex("#c27b0f")
	t.Success = ui.Hex("#22863a")
	t.Selection = ui.RGBA(59, 108, 246, 0.18)
	t.Focus = ui.RGBA(59, 108, 246, 0.45)
	t.Scrollbar = ui.RGBA(0, 0, 0, 0.28)
	t.Radius = 8
	t.InactiveAccent = t.Accent
	return t
}

func categoryColor(cat string) ui.Color {
	switch cat {
	case "Hub":
		return ui.Hex("#3b6cf6")
	case "Runners":
		return ui.Hex("#0f8199")
	case "Image & audio":
		return ui.Hex("#a033c4")
	case "Training":
		return ui.Hex("#22863a")
	case "Engines":
		return ui.Hex("#c27b0f")
	case "Runtime":
		return ui.Hex("#5b6fd6")
	case "Loose weights":
		return ui.Hex("#d93a35")
	case "Large files":
		return ui.Hex("#5c6370")
	default:
		return ui.Hex("#6e6e78")
	}
}

const folderSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>`

const copySVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>`

const refreshSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/><path d="M8 16H3v5"/></svg>`

const sunSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="M4.93 4.93l1.41 1.41"/><path d="M17.66 17.66l1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="M6.34 17.66l-1.41 1.41"/><path d="M19.07 4.93l-1.41 1.41"/></svg>`

const moonSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>`

const listSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 6h13"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M3 6h.01"/><path d="M3 12h.01"/><path d="M3 18h.01"/></svg>`

const backSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12 19-7-7 7-7"/><path d="M19 12H5"/></svg>`

var (
	iconFolder  = ui.MustParseSVG([]byte(folderSVG))
	iconCopy    = ui.MustParseSVG([]byte(copySVG))
	iconRefresh = ui.MustParseSVG([]byte(refreshSVG))
	iconSun     = ui.MustParseSVG([]byte(sunSVG))
	iconMoon    = ui.MustParseSVG([]byte(moonSVG))
	iconList    = ui.MustParseSVG([]byte(listSVG))
	iconBack    = ui.MustParseSVG([]byte(backSVG))
)
