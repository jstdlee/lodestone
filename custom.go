package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// CustomDir is a directory the user asked Lodestone to scan.
type CustomDir struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

var customSeq atomic.Uint64

func newCustomID() string {
	n := customSeq.Add(1)
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatUint(n, 10)
}

func (a *app) customFile() string {
	if a.customPath != "" {
		return a.customPath
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = filepath.Join(a.home, ".config")
	}
	return filepath.Join(dir, "lodestone", "dirs.json")
}

func (a *app) loadCustoms() error {
	b, err := os.ReadFile(a.customFile())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var doc struct {
		Dirs []CustomDir `json:"dirs"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	var out []CustomDir
	seen := map[string]bool{}
	for _, d := range doc.Dirs {
		path, name := cleanCustom(d, a.home)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if d.ID == "" {
			d.ID = newCustomID()
		}
		d.Path = path
		d.Name = name
		out = append(out, d)
	}
	a.customs = out
	return nil
}

func (a *app) saveCustoms() error {
	path := a.customFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	doc := struct {
		Dirs []CustomDir `json:"dirs"`
	}{Dirs: a.customs}
	if doc.Dirs == nil {
		doc.Dirs = []CustomDir{}
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func cleanCustom(d CustomDir, home string) (string, string) {
	path := strings.TrimSpace(d.Path)
	if path == "" {
		return "", ""
	}
	if path == "~" && home != "" {
		path = home
	} else if strings.HasPrefix(path, "~/") && home != "" {
		path = filepath.Join(home, path[2:])
	} else if !filepath.IsAbs(path) && home != "" {
		path = filepath.Join(home, path)
	}
	path = filepath.Clean(path)
	name := strings.TrimSpace(d.Name)
	if name == "" {
		name = filepath.Base(path)
	}
	return path, name
}

func (a *app) beginAdd() {
	a.placesOpen = true
	a.drafting = true
	a.draftName = ""
	a.draftPath = ""
	a.confirmDel = ""
}

func (a *app) beginDuplicate(id string) {
	var src CustomDir
	var ok bool
	for _, d := range a.customs {
		if d.ID == id {
			src, ok = d, true
			break
		}
	}
	if !ok {
		return
	}
	a.placesOpen = true
	a.drafting = true
	a.confirmDel = ""
	a.draftPath = src.Path
	a.draftName = src.Name
	if !strings.HasSuffix(src.Name, " copy") {
		a.draftName += " copy"
	}
}

func (a *app) askDelete(id string) {
	if a.drafting {
		a.status = "Confirm or cancel the new directory first"
		return
	}
	if a.confirmDel == id {
		a.confirmDel = ""
		return
	}
	a.confirmDel = id
}

func (a *app) cancelPending() {
	a.drafting = false
	a.confirmDel = ""
	a.draftName = ""
	a.draftPath = ""
}

func (a *app) confirmPending() error {
	if id := a.confirmDel; id != "" {
		name := id
		var next []CustomDir
		for _, d := range a.customs {
			if d.ID == id {
				name = d.Name
				continue
			}
			next = append(next, d)
		}
		a.customs = next
		a.confirmDel = ""
		if err := a.saveCustoms(); err != nil {
			a.status = err.Error()
			return err
		}
		a.status = "Removed " + name
		a.rescan()
		return nil
	}
	if !a.drafting {
		return nil
	}
	path, name := cleanCustom(CustomDir{Name: a.draftName, Path: a.draftPath}, a.home)
	if path == "" || path == "." {
		a.status = "Enter a directory path"
		return errors.New(a.status)
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		a.status = path + " is a file, not a directory"
		return errors.New(a.status)
	}
	for _, d := range a.customs {
		have, _ := cleanCustom(d, a.home)
		if have == path {
			a.status = "Change the path. That directory is already listed"
			return errors.New(a.status)
		}
	}
	a.customs = append(a.customs, CustomDir{ID: newCustomID(), Name: name, Path: path})
	a.drafting = false
	a.draftName = ""
	a.draftPath = ""
	if err := a.saveCustoms(); err != nil {
		a.status = err.Error()
		return err
	}
	a.status = "Added " + name
	a.rescan()
	return nil
}

func (a *app) opLabel(kind string, d CustomDir) string {
	n := 0
	for _, o := range a.customs {
		if o.Name == d.Name {
			n++
		}
	}
	label := kind + " " + d.Name
	if n > 1 {
		label += " " + filepath.Base(d.Path)
	}
	return label
}

// addCustomStores measures directories the user added. A path already counted
// by a built-in store is listed again but not added to the unique total.
func addCustomStores(snap *Snapshot, extra []CustomDir, measured map[string]string, global *meter, home string, progress func(string)) {
	if len(extra) == 0 {
		return
	}
	seen := map[string]bool{}
	for _, d := range extra {
		path, name := cleanCustom(d, home)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		st := Store{
			ID: "custom:" + d.ID, Name: name, Tools: "custom", Category: "Custom",
			Blurb: "A directory you added.", Path: path, Note: "custom", Active: true,
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			st.Active = false
			if err == nil {
				st.Note = "not a directory"
			} else {
				st.Note = "missing"
			}
			snap.Stores = append(snap.Stores, st)
			continue
		}
		covered := false
		note := "custom"
		if owner, ok := measured[path]; ok {
			covered = true
			note = "already counted with " + owner
		} else if parent, ok := nestedIn(path, measured); ok {
			covered = true
			note = "inside " + parent
		}
		progress("Reading " + name + "…")
		meter := global
		if covered {
			meter = newMeter()
		}
		got, merr := measure(path, "child", meter)
		if merr != nil {
			snap.Warnings = append(snap.Warnings, name+": "+merr.Error())
		}
		st.Bytes = got.Bytes
		st.Entries = got.Entries
		st.Note = note
		if !covered {
			measured[path] = name
			snap.Large = append(snap.Large, got.Large...)
		}
		snap.Stores = append(snap.Stores, st)
	}
	snap.Large = dedupeEntries(snap.Large)
}

func dedupeEntries(in []Entry) []Entry {
	seen := map[string]bool{}
	var out []Entry
	for _, e := range in {
		if seen[e.Path] {
			continue
		}
		seen[e.Path] = true
		out = append(out, e)
	}
	return out
}
