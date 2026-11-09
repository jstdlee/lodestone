package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// app is the state the window shows. The view reads it on the main thread.
type app struct {
	win           *mygo.Window
	home          string
	snap          Snapshot
	rows          []row
	counts        map[string]int
	query         string
	category      string
	dark          bool
	themeReady    bool
	split         float32
	pane          float32
	detail        bool
	gallery       bool
	cards         ui.GridState
	wantScroll    bool
	pendingScroll int
	pulseKey      string
	pulseDown     bool
	pulseN        int
	list          ui.ListState
	selected      int
	lastSel       int
	// selectedKey and anchorPath survive a rebuilt list; the index does not.
	selectedKey string
	anchorPath  string
	scanning    bool
	phase       string
	status      string
	again       bool
	scanGen     int
	customs     []CustomDir
	customPath  string
	placesOpen  bool
	drafting    bool
	draftName   string
	draftPath   string
	confirmDel  string
	pickDir     func() (string, error)
	reveal      func(string) error
	copyText    func(string) error
	scan        func(func(string)) Snapshot
}

func newApp() *app {
	home, _ := os.UserHomeDir()
	return &app{
		home:          home,
		category:      "all",
		selected:      -1,
		lastSel:       -1,
		split:         232,
		pane:          72,
		pendingScroll: -1,
	}
}

func (a *app) rescan() {
	if a.scanning {
		a.again = true
		return
	}
	a.scanning = true
	a.phase = "Reading caches…"
	if a.win == nil {
		a.finishScan(a.runScan(func(string) {}))
		return
	}
	a.scanGen++
	gen := a.scanGen
	win := a.win
	go func() {
		snap := a.runScan(func(phase string) {
			win.Update(func() {
				if a.scanGen != gen || phase == a.phase {
					return
				}
				a.phase = phase
			})
		})
		win.Update(func() {
			if a.scanGen != gen {
				return
			}
			a.finishScan(snap)
		})
	}()
}

func (a *app) runScan(progress func(string)) Snapshot {
	if a.scan != nil {
		return a.scan(progress)
	}
	return scanDirs(realEnviron(), a.customs, a.home, progress)
}

func (a *app) finishScan(snap Snapshot) {
	a.snap = snap
	a.scanning = false
	a.phase = ""
	a.status = itoa(len(snap.Stores)) + " stores · " + itoa(len(snap.Large)) + " large files · " + formatBytes(snap.Unique) + " unique"
	if len(snap.Warnings) > 0 {
		a.status += " · " + snap.Warnings[0]
	}
	if a.again {
		a.again = false
		a.rescan()
	}
	a.maybeShots()
}

func mygoShellShow(path string) { mygo.Shell.ShowItemInFolder(path) }

func mygoCopy(text string) error {
	mygo.Clipboard.WriteText(text)
	return nil
}

func main() {
	a := newApp()
	if err := a.loadCustoms(); err != nil {
		a.status = "Custom directories: " + err.Error()
	}
	mygo.App.WhenReady(func() {
		a.scanning = true
		a.phase = "Reading caches…"
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:     "Lodestone",
			Width:     1180,
			Height:    780,
			MinWidth:  920,
			MinHeight: 640,
			StateKey:  "main",
			Content:   ui.View(a.view),
		})
		a.scanning = false
		a.rescan()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
