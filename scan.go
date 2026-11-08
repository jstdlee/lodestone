package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const looseMin = 64 << 20

// largeMin is the size of a file listed under Large files, model-related or not.
const largeMin = 100 << 20

var weightExt = map[string]bool{
	".safetensors": true, ".gguf": true, ".ggml": true, ".pt": true, ".pth": true,
	".ckpt": true, ".onnx": true, ".sft": true, ".engine": true, ".plan": true, ".ot": true,
}

// Snapshot is one finished read of the disk.
type Snapshot struct {
	Stores    []Store
	Patterns  []PatternView
	Loose     []Entry
	Large     []Entry
	Unique    int64
	Apparent  int64
	Warnings  []string
	FilesSeen int
}

// Store is one directory a tool actually has on disk.
type Store struct {
	ID       string
	Name     string
	Tools    string
	Category string
	Blurb    string
	Layout   string
	Env      string
	Path     string
	Note     string
	Active   bool
	Bytes    int64
	Entries  []Entry
}

// Entry is a model, category folder, or loose weight inside a store.
type Entry struct {
	Name   string
	Path   string
	Bytes  int64
	Detail string
}

// PatternView is a catalog row, present or not.
type PatternView struct {
	ID           string
	Name         string
	Tools        string
	Category     string
	Blurb        string
	Layout       string
	Env          string
	CatalogOnly  bool
	ShareOf      string
	ShareName    string
	SharePath    string
	ActivePath   string
	ActiveExists bool
	ActiveNote   string
	Places       []Place
	Bytes        int64
}

type meter struct {
	seen     map[inode]struct{}
	unique   int64
	apparent int64
}

func newMeter() *meter { return &meter{seen: map[inode]struct{}{}} }

func (m *meter) add(info os.FileInfo) {
	if info == nil || !info.Mode().IsRegular() {
		return
	}
	n := info.Size()
	m.apparent += n
	if id, ok := inodeOf(info); ok {
		if _, dup := m.seen[id]; dup {
			return
		}
		m.seen[id] = struct{}{}
	}
	m.unique += n
}

// Scan reads every pattern and then loose weight files outside those stores.
func Scan(env Environ, progress func(string)) Snapshot {
	return scanDirs(env, nil, "", progress)
}

// scanDirs is Scan plus directories the user added.
func scanDirs(env Environ, extra []CustomDir, home string, progress func(string)) Snapshot {
	if progress == nil {
		progress = func(string) {}
	}
	if env.Home == "" {
		env.Home, _ = os.UserHomeDir()
	}
	progress("Resolving cache locations…")
	var snap Snapshot
	snap.Patterns = make([]PatternView, 0, len(Patterns))
	global := newMeter()
	measured := map[string]string{} // path -> pattern name
	byID := map[string]int{}

	for i := range Patterns {
		p := Patterns[i]
		pv := PatternView{
			ID: p.ID, Name: p.Name, Tools: p.Tools, Category: p.Category,
			Blurb: p.Blurb, Layout: p.Layout, Env: p.Env,
			CatalogOnly: p.CatalogOnly, ShareOf: p.ShareOf, Places: resolve(env, p),
		}
		for _, pl := range pv.Places {
			if pl.Active {
				pv.ActivePath = pl.Path
				pv.ActiveExists = pl.Exists
				pv.ActiveNote = pl.Note
			}
		}
		if pv.ActivePath == "" {
			for _, pl := range pv.Places {
				if pl.Exists {
					pv.ActivePath = pl.Path
					pv.ActiveExists = true
					pv.ActiveNote = pl.Note
					break
				}
			}
		}
		snap.Patterns = append(snap.Patterns, pv)
		id := len(snap.Patterns) - 1
		byID[p.ID] = id
		if p.CatalogOnly {
			continue
		}
		for _, pl := range pv.Places {
			if !pl.Exists || !dirExists(pl.Path) {
				continue
			}
			if _, ok := measured[pl.Path]; ok {
				snap.Patterns[id].Places = markShared(snap.Patterns[id].Places, pl.Path)
				continue
			}
			if _, ok := nestedIn(pl.Path, measured); ok {
				snap.Patterns[id].Places = markShared(snap.Patterns[id].Places, pl.Path)
				continue
			}
			progress("Reading " + p.Name + "…")
			st, err := measure(pl.Path, p.List, global)
			if err != nil {
				snap.Warnings = append(snap.Warnings, p.Name+": "+err.Error())
			}
			measured[pl.Path] = p.Name
			snap.Patterns[id].Bytes += st.Bytes
			snap.Large = append(snap.Large, st.Large...)
			snap.Stores = append(snap.Stores, Store{
				ID: p.ID, Name: p.Name, Tools: p.Tools, Category: p.Category,
				Blurb: p.Blurb, Layout: p.Layout, Env: p.Env,
				Path: pl.Path, Note: pl.Note, Active: pl.Active,
				Bytes: st.Bytes, Entries: st.Entries,
			})
		}
	}
	for i := range snap.Patterns {
		pv := &snap.Patterns[i]
		if pv.ShareOf == "" {
			continue
		}
		other, ok := byID[pv.ShareOf]
		if !ok {
			continue
		}
		src := snap.Patterns[other]
		pv.ShareName = src.Name
		pv.SharePath = src.ActivePath
		if pv.ActivePath == "" {
			pv.ActivePath = src.ActivePath
			pv.ActiveExists = src.ActiveExists
			pv.ActiveNote = "shared with " + src.Name
		}
	}

	addCustomStores(&snap, extra, measured, global, home, progress)
	progress("Looking for loose weights and large files…")
	var roots []string
	for path := range measured {
		roots = append(roots, path)
	}
	loose, homeLarge, seen, warn := looseWalk(env.Home, roots, global)
	snap.Loose = loose
	snap.Large = append(snap.Large, homeLarge...)
	sortEntries(snap.Large)
	if len(snap.Large) > 400 {
		snap.Large = snap.Large[:400]
	}
	snap.FilesSeen = seen
	if warn != "" {
		snap.Warnings = append(snap.Warnings, warn)
	}
	snap.Unique = global.unique
	snap.Apparent = global.apparent
	return snap
}
func nestedIn(path string, measured map[string]string) (string, bool) {
	path = filepath.Clean(path)
	sep := string(filepath.Separator)
	for root, name := range measured {
		root = filepath.Clean(root)
		if path != root && strings.HasPrefix(path, root+sep) {
			return name, true
		}
	}
	return "", false
}

func markShared(places []Place, path string) []Place {
	for i := range places {
		if places[i].Path == path {
			places[i].Shared = true
		}
	}
	return places
}

type measuredDir struct {
	Bytes   int64
	Entries []Entry
	Large   []Entry
}

func measure(root, list string, global *meter) (measuredDir, error) {
	var out measuredDir
	childBytes := map[string]int64{}
	childDetail := map[string]string{}
	var firstErr error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if firstErr == nil && path != root {
				firstErr = err
			}
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (skipDir(name) || secretName(name)) {
				return filepath.SkipDir
			}
			if d.Type()&os.ModeSymlink != 0 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || secretName(name) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		n := info.Size()
		out.Bytes += n
		global.add(info)
		if n >= largeMin {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
			out.Large = append(out.Large, Entry{Name: name, Path: path, Bytes: n, Detail: ext})
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}
		parts := strings.Split(rel, string(filepath.Separator))
		switch list {
		case "hub", "child":
			if secretName(parts[0]) || strings.HasPrefix(parts[0], ".") {
				return nil
			}
			childBytes[parts[0]] += n
			if list == "hub" {
				if title, kind, ok := decodeHub(parts[0]); ok {
					childDetail[parts[0]] = title + "\x00" + kind
				}
			}
		case "pair":
			key := parts[0]
			if len(parts) >= 2 {
				key = parts[0] + "/" + parts[1]
			}
			if secretName(parts[0]) || strings.HasPrefix(parts[0], ".") {
				return nil
			}
			childBytes[key] += n
		case "ollama":
			if len(parts) >= 2 && parts[0] == "manifests" && !d.IsDir() {
				childBytes[rel] += n
				childDetail[rel] = ollamaName(parts)
			}
		}
		return nil
	})
	if err != nil && firstErr == nil {
		firstErr = err
	}
	if list == "ollama" {
		out.Entries = ollamaEntries(root, childBytes, childDetail)
		return out, firstErr
	}
	for key, n := range childBytes {
		if n <= 0 || secretName(filepath.Base(key)) {
			continue
		}
		title := key
		detail := ""
		if v, ok := childDetail[key]; ok && list == "hub" {
			title, detail, _ = strings.Cut(v, "\x00")
		}
		entryPath := filepath.Join(root, filepath.FromSlash(key))
		out.Entries = append(out.Entries, Entry{Name: title, Path: entryPath, Bytes: n, Detail: detail})
	}
	sortEntries(out.Entries)
	return out, firstErr
}

func ollamaName(parts []string) string {
	// manifests/<registry>/<namespace>/<name>/<tag>
	if len(parts) < 5 {
		return strings.Join(parts[1:], "/")
	}
	name := parts[len(parts)-2] + ":" + parts[len(parts)-1]
	ns := parts[len(parts)-3]
	if ns != "library" {
		name = ns + "/" + name
	}
	return name
}

func ollamaEntries(root string, files map[string]int64, names map[string]string) []Entry {
	blobs := blobSizes(root)
	var out []Entry
	for rel, name := range names {
		manifest := filepath.Join(root, rel)
		n, weight := manifestBytes(manifest, root, blobs)
		if n == 0 {
			n = files[rel]
		}
		// The manifest is a small JSON file. The bytes it reports live in blobs/.
		path := weight
		detail := "weight"
		if path == "" {
			path = manifest
			detail = "tag"
		}
		out = append(out, Entry{Name: name, Path: path, Bytes: n, Detail: detail})
	}
	sortEntries(out)
	return out
}

func blobSizes(root string) map[string]int64 {
	dir := filepath.Join(root, "blobs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := map[string]int64{}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		out[e.Name()] = info.Size()
	}
	return out
}

// manifestBytes sums the layers and returns the largest blob that exists.
// A missing blob still counts by the size written in the manifest.
func manifestBytes(path, root string, blobs map[string]int64) (int64, string) {
	b, err := os.ReadFile(path)
	if err != nil || len(b) > 2<<20 {
		return 0, ""
	}
	var doc struct {
		Layers []struct {
			Digest string `json:"digest"`
			Size   int64  `json:"size"`
		} `json:"layers"`
	}
	if json.Unmarshal(b, &doc) != nil {
		return 0, ""
	}
	var n, best int64
	var bestPath string
	for _, layer := range doc.Layers {
		key := strings.ReplaceAll(layer.Digest, ":", "-")
		sz, ok := int64(0), false
		if blobs != nil {
			sz, ok = blobs[key]
		}
		blobPath := ""
		if ok {
			n += sz
			blobPath = filepath.Join(root, "blobs", key)
		} else if layer.Size > 0 {
			n += layer.Size
			sz = layer.Size
		}
		if blobPath != "" && sz >= best {
			best = sz
			bestPath = blobPath
		}
	}
	return n, bestPath
}

func sortEntries(entries []Entry) {
	for i := 1; i < len(entries); i++ {
		j := i
		for j > 0 && (entries[j].Bytes > entries[j-1].Bytes || (entries[j].Bytes == entries[j-1].Bytes && entries[j].Name < entries[j-1].Name)) {
			entries[j], entries[j-1] = entries[j-1], entries[j]
			j--
		}
	}
}

func decodeHub(folder string) (title, kind string, ok bool) {
	switch {
	case strings.HasPrefix(folder, "models--"):
		kind, folder = "model", strings.TrimPrefix(folder, "models--")
	case strings.HasPrefix(folder, "datasets--"):
		kind, folder = "dataset", strings.TrimPrefix(folder, "datasets--")
	case strings.HasPrefix(folder, "spaces--"):
		kind, folder = "space", strings.TrimPrefix(folder, "spaces--")
	default:
		return "", "", false
	}
	org, repo, found := strings.Cut(folder, "--")
	if !found || org == "" || repo == "" {
		return folder, kind, folder != ""
	}
	return org + "/" + repo, kind, true
}

func skipDir(name string) bool {
	switch name {
	case "node_modules", ".git", "__pycache__", "site-packages", ".npm", ".cargo",
		".rustup", ".gradle", ".m2", "snap", ".steam", ".var", "mcp-oauth-tokens":
		return true
	}
	return strings.HasPrefix(name, ".venv")
}

func secretName(name string) bool {
	switch strings.ToLower(name) {
	case "token", "stored_tokens", "credentials", "credential", ".env", "auth",
		"mcp-oauth-tokens", "hfsearchtoken", "hfdownloadtoken":
		return true
	}
	return false
}

func looseWalk(home string, stores []string, global *meter) (loose, large []Entry, seen int, warn string) {
	if home == "" || !dirExists(home) {
		return nil, nil, 0, ""
	}
	const maxVisit = 1500000
	stopped := false
	_ = filepath.WalkDir(home, func(path string, d fs.DirEntry, err error) error {
		if stopped {
			return filepath.SkipAll
		}
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if path != home && (skipDir(name) || secretName(name) || underAny(path, stores)) {
				return filepath.SkipDir
			}
			return nil
		}
		seen++
		if seen > maxVisit {
			stopped = true
			return filepath.SkipAll
		}
		if d.Type()&os.ModeSymlink != 0 || secretName(name) {
			return nil
		}
		info, err := d.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		size := info.Size()
		ext := strings.ToLower(filepath.Ext(name))
		isWeight := size >= looseMin && (weightExt[ext] || ext == ".bin")
		if ext == ".bin" && (size < 200<<20 || info.Mode()&0o111 != 0) {
			isWeight = false
		}
		isLarge := size >= largeMin
		if !isWeight && !isLarge {
			return nil
		}
		if underAny(path, stores) {
			return nil
		}
		global.add(info)
		detail := strings.TrimPrefix(ext, ".")
		if isWeight {
			loose = append(loose, Entry{Name: name, Path: path, Bytes: size, Detail: detail})
		}
		if isLarge {
			large = append(large, Entry{Name: name, Path: path, Bytes: size, Detail: detail})
		}
		return nil
	})
	sortEntries(loose)
	if len(loose) > 400 {
		loose = loose[:400]
	}
	if stopped {
		return loose, large, seen, "loose scan stopped early"
	}
	return loose, large, seen, ""
}

func underAny(path string, roots []string) bool {
	path = filepath.Clean(path)
	sep := string(filepath.Separator)
	for _, root := range roots {
		root = filepath.Clean(root)
		if path == root || strings.HasPrefix(path, root+sep) {
			return true
		}
	}
	return false
}
