package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// maybeShots writes before and after light shots when LODESTONE_SHOTS
// names a directory, then closes the window.
func (a *app) maybeShots() {
	dir := os.Getenv("LODESTONE_SHOTS")
	if dir == "" || a.win == nil {
		return
	}
	os.Unsetenv("LODESTONE_SHOTS")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("shots: %v", err)
		return
	}
	go a.captureShots(dir)
}

func (a *app) captureShots(dir string) {
	time.Sleep(700 * time.Millisecond)
	a.win.SetContentSize(1180, 760)
	a.win.Update(func() {
		a.dark = true
		a.themeReady = true
		a.gallery = false
		a.detail = false
		a.placesOpen = false
		a.query = ""
		a.category = "all"
		a.selected = -1
		a.lastSel = -1
		a.selectedKey = ""
		a.pulseKey = ""
	})
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "map.png"))

	a.win.Update(func() { a.gallery = true })
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "gallery.png"))

	a.win.Update(func() {
		a.gallery = false
		a.detail = true
		rows, _ := a.visibleRows()
		if len(rows) > 0 {
			a.selected = 0
			a.lastSel = 0
			a.selectedKey = rows[0].key
			a.anchorPath = rows[0].path
			a.pulseKey = rows[0].key
			a.pulseN = 6
		}
	})
	time.Sleep(900 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "details.png"))

	// a search that filters everything away, with the way back
	a.win.Update(func() {
		a.detail = false
		a.query = "nothing-here"
	})
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "search-empty.png"))

	// the chips wrap at a narrow window
	a.win.Update(func() { a.query = "" })
	a.win.SetContentSize(520, 720)
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "narrow.png"))

	// the same three, in the light chrome
	a.win.Update(func() {
		a.dark = !a.dark
		a.gallery = false
		a.detail = false
		a.selected = -1
		a.lastSel = -1
		a.selectedKey = ""
		a.pulseKey = ""
	})
	a.win.SetContentSize(1180, 760)
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "light-map.png"))

	a.win.Update(func() { a.gallery = true })
	time.Sleep(800 * time.Millisecond)
	a.writeShot(filepath.Join(dir, "light-gallery.png"))
	a.win.Close()
}

func (a *app) writeShot(path string) {
	png, err := a.win.CapturePage()
	if err != nil {
		log.Printf("shot %s: %v", path, err)
		return
	}
	if err := os.WriteFile(path, png, 0o644); err != nil {
		log.Printf("shot %s: %v", path, err)
	}
}
