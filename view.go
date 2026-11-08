package main

import (
	"time"

	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
)

const treemapCap = 48

// row is one line in the list.
type row struct {
	key      string
	name     string
	sub      string
	path     string
	category string
	tools    string
	blurb    string
	layout   string
	env      string
	note     string
	detail   string
	bytes    int64
	exists   bool
}

type catStat struct {
	name  string
	n     int
	bytes int64
}

type mapEntry struct {
	key, name, path, cat string
	bytes                int64
	other                bool
}

func (a *app) view(c *ui.Context) {
	if a.split < 40 {
		a.split = 208
	}
	if a.pane < 20 || a.pane > 90 {
		a.pane = 68
	}
	t := a.applyTheme(c)
	a.rows, a.counts = a.visibleRows()
	a.syncSelection()
	matched := a.matchedRows()
	ui.Column(c).Fill().Padding(14, 16).Gap(10).Children(func() {
		a.header(c, t)
		ui.SearchField(c, &a.query).Placeholder("Name or path").Label("Search").Width(360)
		if !a.detail {
			a.dashboard(c, t, matched)
			a.treemap(c, t, true)
		} else {
			a.treemap(c, t, false)
			ui.Split(c, &a.split, func() { a.sidebar(c) }, func() {
				ui.Column(c).Fill().Background(t.Surface).Radius(t.Radius).Border(1, t.Border).Clip().Children(func() {
					if _, ok := a.current(); ok {
						ui.SplitPercent(c, &a.pane, func() { a.listPane(c, t) }, func() { a.inspector(c, t) }).Fill()
						return
					}
					a.listPane(c, t)
				})
			}).Fill().Grow(1)
		}
		if status := a.statusLine(); status != "" {
			ui.Text(c, status).TextColor(t.TextMuted).Role(ui.RoleStatus)
		}
	})
	a.finishSelection()
}

func (a *app) applyTheme(c *ui.Context) *ui.Theme {
	if !a.themeReady {
		a.dark = c.Theme().Dark
		a.themeReady = true
	}
	t := keroTheme(a.dark)
	c.SetTheme(t)
	return t
}

func (a *app) header(c *ui.Context, t *ui.Theme) {
	ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
		if a.detail {
			a.headerIcon(c, iconBack, "Home", "Back to the map", true, func() { a.detail = false })
		}
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, "Lodestone").FontSize(20).Bold()
			ui.Text(c, "Model caches, loose weights, and large files.").TextColor(t.TextMuted)
		})
		a.headerIcon(c, iconRefresh, "Rescan", "Scan again", !a.scanning, a.rescan)
		themeIcon := iconMoon
		themeTip := "Dark theme"
		if a.dark {
			themeIcon = iconSun
			themeTip = "Light theme"
		}
		a.headerIcon(c, themeIcon, "Theme", themeTip, true, func() { a.dark = !a.dark })
		if !a.detail {
			a.headerIcon(c, iconList, "Details", "Category list", true, func() { a.detail = true })
		}
	})
}

func (a *app) headerIcon(c *ui.Context, icon *ui.SVG, label, tip string, enabled bool, fn func()) {
	ui.Button(c, "").Label(label).Tooltip(tip).Disabled(!enabled).Shrink(0).Padding(6).OnClick(fn).Children(func() {
		ui.Icon(c, icon).FontSize(16)
	})
}

func (a *app) statusLine() string {
	if a.scanning && a.phase != "" {
		return a.phase
	}
	if a.status != "" {
		return a.status
	}
	if n := len(a.snap.Warnings); n > 0 {
		return a.snap.Warnings[0]
	}
	return ""
}

func (a *app) dashboard(c *ui.Context, t *ui.Theme, matched []row) {
	stats := categoryStats(matched)
	ui.Row(c).Gap(8).AlignItems(ui.Stretch).Children(func() {
		a.statChip(c, t, "all", "All", formatBytes(a.snap.Unique), "unique", a.category == "" || a.category == "all")
		for _, st := range stats {
			a.statChip(c, t, st.name, st.name, formatBytes(st.bytes), itoa(st.n), a.category == st.name)
		}
	})
}

func (a *app) statChip(c *ui.Context, t *ui.Theme, id, label, size, hint string, on bool) {
	border := t.Border
	if on {
		border = t.Accent
	}
	chip := ui.Column(c).Padding(7, 10).Gap(1).Radius(8).Background(t.Surface).Border(1, border).MinWidth(96).OnClick(func() {
		a.category = id
	})
	chip.Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			if id != "all" {
				ui.Box(c).Size(8, 8).Radius(4).Background(categoryColor(id)).Shrink(0)
			}
			ui.Text(c, label).FontSize(12).TextColor(t.TextMuted).SingleLine()
		})
		ui.Text(c, size).Bold().SingleLine()
		ui.Text(c, hint).FontSize(11).TextColor(t.TextMuted)
	})
}

func (a *app) treemap(c *ui.Context, t *ui.Theme, hero bool) {
	entries := treemapEntries(a.rows)
	sizes := make([]int64, len(entries))
	for i := range entries {
		sizes[i] = entries[i].bytes
	}
	box := ui.Box(c).FillWidth().Clip().Radius(8).Background(t.Surface).Border(1, t.Border)
	if hero {
		box = box.Grow(1).MinHeight(280)
	} else {
		box = box.Height(132)
	}
	pulse := a.blink(box)
	box.Children(func() {
		if len(entries) == 0 {
			msg := "Nothing to map"
			if strings.TrimSpace(a.query) != "" {
				msg = "No matches"
			}
			ui.Column(c).Fill().Center().Children(func() {
				ui.Text(c, msg).TextColor(t.TextMuted)
			})
			return
		}
		for i := range entries {
			entry := entries[i]
			ui.Box(c).Absolute().Fill().
				HitTest(func(x, y float32, r ui.Rect) bool {
					return cellContains(layoutTreemap(sizes, r.W, r.H), i, x-r.X, y-r.Y)
				}).
				OnClick(func() {
					if entry.other || entry.key == "" {
						return
					}
					a.selectedKey = entry.key
					a.anchorPath = entry.path
					a.armPulse(entry.key)
					a.wantScroll = true
					a.detail = true
				}).
				Draw(func(p *ui.Painter, r ui.Rect) {
					cells := layoutTreemap(sizes, r.W, r.H)
					cell, ok := cellByIndex(cells, i)
					if !ok || cell.w < 2 || cell.h < 2 {
						return
					}
					color := categoryColor(entry.cat)
					if entry.other {
						color = t.Border
					}
					rect := ui.Rect{X: r.X + cell.x + 1, Y: r.Y + cell.y + 1, W: cell.w - 2, H: cell.h - 2}
					if rect.W < 1 || rect.H < 1 {
						return
					}
					p.Fill(rect, color, 4)
					selected := entry.key != "" && entry.key == a.selectedKey
					if selected {
						wash := float32(0.14)
						if pulse > 0 {
							wash = 0.08 + 0.18*pulse
						}
						p.Fill(rect, ui.Hex("#ffffff").Alpha(wash), 4)
						inset := ui.Rect{X: rect.X + 5, Y: rect.Y + 5, W: rect.W - 10, H: rect.H - 10}
						if inset.W > 8 && inset.H > 8 {
							p.Stroke(inset, ui.Hex("#ffffff"), 4, 4)
						}
					}
					a.paintBlock(p, rect, entry, selected)
				})
		}
	})
}

func (a *app) paintBlock(p *ui.Painter, box ui.Rect, entry mapEntry, selected bool) {
	nameSize, subSize := blockFont(box.W, box.H)
	if nameSize < 9 {
		return
	}
	ink := blockInk(entry)
	if selected {
		ink = ui.Hex("#ffffff")
		if entry.cat == "Engines" || entry.other {
			ink = ui.Hex("#1d1d1f")
		}
	}
	pad := float32(6)
	span, _, th, fit := fitLabel(p, entry.name, box.W-pad*2, nameSize, ink)
	if !fit {
		return
	}
	x := box.X + pad
	sub := ui.Span{}
	var sh float32
	showSub := false
	if subSize > 0 {
		sub = ui.Span{Text: formatBytes(entry.bytes), Size: subSize, Weight: 500, Color: ink.Alpha(0.88)}
		if sw, h := p.MeasureText(0, sub); sw <= box.W-pad*2 && th+h+4 <= box.H-pad {
			sh = h
			showSub = true
		}
	}
	y := box.Y + (box.H-th)/2
	if showSub {
		y = box.Y + (box.H-(th+sh+2))/2
		p.RichText(x, y+th+2, 0, sub)
	}
	p.RichText(x, y, 0, span)
}

func blockFont(w, h float32) (name, sub float32) {
	if w < 24 || h < 16 {
		return 0, 0
	}
	short := w
	if h < short {
		short = h
	}
	name = short * 0.26
	if name > 22 {
		name = 22
	}
	if name < 10 {
		name = 10
	}
	if h < name+8 {
		name = h * 0.55
		if name < 9 {
			return 0, 0
		}
		return name, 0
	}
	sub = name * 0.72
	if sub < 9 {
		sub = 9
	}
	if h < name+sub+8 {
		return name, 0
	}
	return name, sub
}

func blockInk(entry mapEntry) ui.Color {
	if entry.other || entry.cat == "Engines" {
		return ui.Hex("#1d1d1f")
	}
	return ui.Hex("#ffffff")
}

func (a *app) blink(el ui.Element) float32 {
	if a.selectedKey == "" || a.pulseKey != a.selectedKey {
		return 0
	}
	if a.pulseN >= 6 {
		return 1
	}
	target := float32(1)
	if a.pulseDown {
		target = 0.55
	}
	v := el.AnimateTo("blink", target, ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.EaseInOut})
	if abs32(v-target) < 0.05 {
		a.pulseDown = !a.pulseDown
		a.pulseN++
	}
	if a.pulseN >= 6 {
		return 1
	}
	return v
}

func (a *app) armPulse(key string) {
	if key == "" {
		return
	}
	a.pulseKey = key
	a.pulseN = 0
	a.pulseDown = false
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func cellContains(cells []treemapCell, index int, x, y float32) bool {
	cell, ok := cellByIndex(cells, index)
	if !ok {
		return false
	}
	return x >= cell.x && x < cell.x+cell.w && y >= cell.y && y < cell.y+cell.h
}

func cellByIndex(cells []treemapCell, index int) (treemapCell, bool) {
	for _, cell := range cells {
		if cell.index == index {
			return cell, cell.w > 0 && cell.h > 0
		}
	}
	return treemapCell{}, false
}

func fitLabel(p *ui.Painter, text string, maxW, size float32, color ui.Color) (ui.Span, float32, float32, bool) {
	if maxW <= 8 || text == "" || size < 8 {
		return ui.Span{}, 0, 0, false
	}
	full := ui.Span{Text: text, Size: size, Weight: 600, Color: color}
	w, h := p.MeasureText(0, full)
	if w <= maxW {
		return full, w, h, true
	}
	rs := []rune(text)
	best := ""
	var bw, bh float32
	lo, hi := 1, len(rs)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		span := ui.Span{Text: string(rs[:mid]) + "…", Size: size, Weight: 600, Color: color}
		w, h = p.MeasureText(0, span)
		if w <= maxW {
			best = span.Text
			bw, bh = w, h
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best == "" {
		return ui.Span{}, 0, 0, false
	}
	return ui.Span{Text: best, Size: size, Weight: 600, Color: color}, bw, bh, true
}

func treemapEntries(rows []row) []mapEntry {
	if len(rows) == 0 {
		return nil
	}
	n := len(rows)
	var rest int64
	if n > treemapCap {
		for _, r := range rows[treemapCap:] {
			rest += r.bytes
		}
		n = treemapCap
	}
	out := make([]mapEntry, 0, n+1)
	for _, r := range rows[:n] {
		if r.bytes <= 0 {
			continue
		}
		out = append(out, mapEntry{key: r.key, name: r.name, path: r.path, cat: r.category, bytes: r.bytes})
	}
	if rest > 0 {
		out = append(out, mapEntry{name: "Other", cat: "Large files", bytes: rest, other: true})
	}
	return out
}

func (a *app) sidebar(c *ui.Context) {
	ui.Sidebar(c, &a.category, func() {
		ui.SidebarSection(c, "Category", nil, func() {
			a.sideItem(c, "all", "All")
			for _, cat := range Categories {
				a.sideItem(c, cat, cat)
			}
		})
	}).Fill()
}

func (a *app) sideItem(c *ui.Context, id, label string) {
	ui.SidebarItem(c, id, nil, label).Children(func() {
		ui.Badge(c, itoa(a.counts[id]))
	})
}

func (a *app) listPane(c *ui.Context, t *ui.Theme) {
	if a.pendingScroll >= 0 && a.pendingScroll < len(a.rows) {
		a.list.ScrollIntoView(a.pendingScroll)
	}
	a.pendingScroll = -1
	if len(a.rows) == 0 {
		msg := "Nothing in this view"
		if a.scanning {
			msg = "Reading caches…"
		} else if strings.TrimSpace(a.query) != "" {
			msg = "No matches"
		}
		ui.Column(c).Fill().Center().Children(func() {
			ui.Text(c, msg).TextColor(t.TextMuted)
		})
		return
	}
	var max int64 = 1
	for _, r := range a.rows {
		if r.bytes > max {
			max = r.bytes
		}
	}
	a.list.Selected = &a.selected
	list := ui.List(c, &a.list, len(a.rows)).Fill().Grow(1).
		ItemKey(func(i int) any {
			if i < 0 || i >= len(a.rows) {
				return i
			}
			return a.rows[i].key
		}).
		Rows(func(lr ui.ListRow) {
			r := a.rows[lr.Index]
			rc := lr.Context
			ui.Row(rc).Padding(6, 12).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(rc).Size(8, 8).Radius(4).Background(categoryColor(r.category)).Shrink(0)
				ui.Column(rc).Grow(1).Gap(1).Children(func() {
					ui.Text(rc, r.name).SingleLine()
					ui.Text(rc, a.rowSub(r)).SingleLine().FontSize(12).TextColor(t.TextMuted)
				})
				ui.Column(rc).Width(132).Gap(3).Shrink(0).Children(func() {
					ui.Meter(rc, float64(r.bytes), 0, float64(max), nil).Height(6)
					label := formatBytes(r.bytes)
					if !r.exists && r.bytes == 0 {
						label = "missing"
					}
					ui.Text(rc, label).FontSize(12).TextAlign(ui.End)
				})
				key, path, exists := r.key, r.path, r.exists
				a.iconButton(rc, iconFolder, "Open location", "Show in folder", exists && path != "", func() {
					a.openRow(key, path, exists)
				})
				a.iconButton(rc, iconCopy, "Copy path", "Copy path", path != "", func() {
					a.copyRow(key, path)
				})
			})
		})
	if list.Submitted() {
		a.openCurrent()
	}
}

func (a *app) iconButton(c *ui.Context, icon *ui.SVG, label, tip string, enabled bool, fn func()) {
	ui.Button(c, "").Label(label).Tooltip(tip).Disabled(!enabled).Shrink(0).Padding(4).OnClick(fn).Children(func() {
		ui.Icon(c, icon).FontSize(15)
	})
}

func (a *app) rowSub(r row) string {
	parts := make([]string, 0, 3)
	if r.sub != "" {
		parts = append(parts, r.sub)
	}
	if r.detail != "" {
		parts = append(parts, r.detail)
	}
	if r.path != "" {
		parts = append(parts, tilde(a.home, r.path))
	}
	return strings.Join(parts, " · ")
}

func (a *app) inspector(c *ui.Context, t *ui.Theme) {
	r, ok := a.current()
	ui.Column(c).Fill().Padding(12, 14).Gap(6).Children(func() {
		if !ok {
			ui.Text(c, "Select a row").TextColor(t.TextMuted)
			return
		}
		ui.Text(c, r.name).Bold().SingleLine()
		if r.sub != "" {
			ui.Text(c, r.sub).TextColor(t.TextMuted).SingleLine()
		}
		ui.Text(c, formatBytes(r.bytes)+" · "+r.category).FontSize(12).TextColor(t.TextMuted)
		if r.blurb != "" {
			ui.Text(c, r.blurb).MaxLines(4).TextColor(t.TextMuted)
		}
		if r.path != "" {
			ui.Text(c, "Path").FontSize(12).TextColor(t.TextMuted)
			ui.Text(c, tilde(a.home, r.path)).Selectable().MaxLines(3)
		}
		if r.layout != "" {
			ui.Text(c, r.layout).FontSize(12).MaxLines(2)
		}
		if r.env != "" {
			ui.Text(c, r.env).FontSize(12).TextColor(t.TextMuted).MaxLines(2)
		}
		if r.note != "" {
			ui.Text(c, r.note).FontSize(12).TextColor(t.TextMuted).MaxLines(2)
		}
	})
}

func (a *app) visibleRows() ([]row, map[string]int) {
	all := diskRows(a.snap)
	counts := map[string]int{}
	var out []row
	for _, r := range all {
		if !rowMatches(a.home, a.query, r) {
			continue
		}
		counts[r.category]++
		counts["all"]++
		if a.category != "" && a.category != "all" && r.category != a.category {
			continue
		}
		out = append(out, r)
	}
	return out, counts
}

func (a *app) matchedRows() []row {
	all := diskRows(a.snap)
	var out []row
	for _, r := range all {
		if rowMatches(a.home, a.query, r) {
			out = append(out, r)
		}
	}
	return out
}

func categoryStats(rows []row) []catStat {
	bytes := map[string]int64{}
	count := map[string]int{}
	for _, r := range rows {
		bytes[r.category] += r.bytes
		count[r.category]++
	}
	var out []catStat
	seen := map[string]bool{}
	for _, cat := range Categories {
		seen[cat] = true
		if count[cat] == 0 {
			continue
		}
		out = append(out, catStat{name: cat, n: count[cat], bytes: bytes[cat]})
	}
	for cat, n := range count {
		if seen[cat] || n == 0 {
			continue
		}
		out = append(out, catStat{name: cat, n: n, bytes: bytes[cat]})
	}
	return out
}

func diskRows(snap Snapshot) []row {
	var out []row
	for _, st := range snap.Stores {
		if len(st.Entries) == 0 {
			if st.Bytes <= 0 {
				continue
			}
			out = append(out, row{
				key: "store\x00" + st.Path, name: st.Name, sub: st.Note,
				path: st.Path, category: st.Category, tools: st.Tools,
				blurb: st.Blurb, layout: st.Layout, env: st.Env, note: st.Note,
				bytes: st.Bytes, exists: true,
			})
			continue
		}
		for _, e := range st.Entries {
			out = append(out, row{
				key: st.ID + "\x00" + e.Path, name: e.Name, sub: st.Name,
				path: e.Path, category: st.Category, tools: st.Tools,
				blurb: st.Blurb, layout: st.Layout, env: st.Env,
				detail: e.Detail, bytes: e.Bytes, exists: true,
			})
		}
	}
	for _, e := range snap.Loose {
		out = append(out, row{
			key: "loose\x00" + e.Path, name: e.Name, sub: "Loose file",
			path: e.Path, category: "Loose weights", detail: e.Detail,
			bytes: e.Bytes, exists: true,
			blurb: "A weight file outside the known caches.",
		})
	}
	for _, e := range snap.Large {
		out = append(out, row{
			key: "large\x00" + e.Path, name: e.Name, sub: "Large file",
			path: e.Path, category: "Large files", detail: e.Detail,
			bytes: e.Bytes, exists: true,
			blurb: "Any file of at least 100 MB, inside a model store or elsewhere under the home directory.",
		})
	}
	sortRows(out)
	return out
}

func sortRows(rows []row) {
	for i := 1; i < len(rows); i++ {
		j := i
		for j > 0 && (rows[j].bytes > rows[j-1].bytes || (rows[j].bytes == rows[j-1].bytes && rows[j].name < rows[j-1].name)) {
			rows[j], rows[j-1] = rows[j-1], rows[j]
			j--
		}
	}
}

func rowMatches(home, query string, r row) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	fields := []string{r.name, r.path, filepath.Base(r.path), tilde(home, r.path), r.tools, r.category, r.sub, r.blurb, r.note, r.detail}
	for _, seg := range strings.FieldsFunc(r.path, func(ch rune) bool { return ch == '/' || ch == '\\' }) {
		fields = append(fields, seg)
	}
	for _, s := range fields {
		if strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}

func (a *app) syncSelection() {
	if a.selected != a.lastSel && a.selected >= 0 && a.selected < len(a.rows) {
		key := a.rows[a.selected].key
		if key != a.selectedKey {
			a.armPulse(key)
		}
		a.selectedKey = key
		a.anchorPath = a.rows[a.selected].path
		return
	}
	if i := rowIndex(a.rows, a.selectedKey); i >= 0 {
		a.selected = i
		if a.wantScroll {
			a.pendingScroll = i
			a.wantScroll = false
		}
		return
	}
	if i := pathIndex(a.rows, a.anchorPath); i >= 0 {
		a.selected = i
		a.selectedKey = a.rows[i].key
		return
	}
	a.selected = -1
}

func (a *app) finishSelection() {
	if a.selected >= 0 && a.selected < len(a.rows) {
		key := a.rows[a.selected].key
		if key != a.selectedKey {
			a.armPulse(key)
		}
		a.selectedKey = key
		a.anchorPath = a.rows[a.selected].path
		a.lastSel = a.selected
		return
	}
	a.selected = -1
	a.lastSel = -1
}

func (a *app) current() (row, bool) {
	if i := rowIndex(a.rows, a.selectedKey); i >= 0 {
		return a.rows[i], true
	}
	if a.selected >= 0 && a.selected < len(a.rows) {
		return a.rows[a.selected], true
	}
	return row{}, false
}

func (a *app) openCurrent() {
	r, ok := a.current()
	if !ok {
		return
	}
	a.openRow(r.key, r.path, r.exists)
}

func (a *app) openRow(key, path string, exists bool) {
	if key != "" {
		a.selectedKey = key
		a.anchorPath = path
	}
	if !exists || path == "" {
		return
	}
	if a.reveal != nil {
		if err := a.reveal(path); err != nil {
			a.status = err.Error()
			return
		}
		a.status = "Opened " + tilde(a.home, path)
		return
	}
	mygoShellShow(path)
	a.status = "Opened " + tilde(a.home, path)
}

func (a *app) copyRow(key, path string) {
	if path == "" {
		return
	}
	if key != "" {
		a.selectedKey = key
		a.anchorPath = path
	}
	if a.copyText != nil {
		if err := a.copyText(path); err != nil {
			a.status = err.Error()
			return
		}
		a.status = "Copied " + tilde(a.home, path)
		return
	}
	if err := mygoCopy(path); err != nil {
		a.status = err.Error()
		return
	}
	a.status = "Copied " + tilde(a.home, path)
}

func rowIndex(rows []row, key string) int {
	if key == "" {
		return -1
	}
	for i := range rows {
		if rows[i].key == key {
			return i
		}
	}
	return -1
}

func pathIndex(rows []row, path string) int {
	if path == "" {
		return -1
	}
	for i := range rows {
		if rows[i].path == path {
			return i
		}
	}
	return -1
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
