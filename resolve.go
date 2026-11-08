package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Pattern is one tool's on-disk convention.
type Pattern struct {
	ID          string
	Name        string
	Tools       string
	Category    string
	Blurb       string
	Layout      string
	Env         string
	List        string // hub, child, pair, ollama, or empty
	CatalogOnly bool
	ShareOf     string
	Cands       []cand
}

type cand struct {
	Kind     string // env, envjoin, home, xdg, abs, discover
	Env      string
	Join     string
	Path     string
	Note     string
	Discover string
	Unless   []string
	Shadow   bool
}

// Place is one resolved directory.
type Place struct {
	Path   string
	Note   string
	Active bool
	Exists bool
	Shared bool
}

// Environ is the home directory and environment the scan reads.
type Environ struct {
	Home string
	Get  func(string) string
}

func realEnviron() Environ {
	home, _ := os.UserHomeDir()
	return Environ{Home: home, Get: os.Getenv}
}

func (e Environ) get(k string) string {
	if e.Get == nil {
		return ""
	}
	return strings.TrimSpace(e.Get(k))
}

func (e Environ) blocked(keys []string) bool {
	for _, k := range keys {
		if e.get(k) != "" {
			return true
		}
	}
	return false
}

func (e Environ) home(rel string) string {
	return filepath.Clean(filepath.Join(e.Home, rel))
}

func resolve(env Environ, p Pattern) []Place {
	var out []Place
	seen := map[string]bool{}
	haveActive := false
	add := func(path, note string, shadow bool) {
		path = strings.TrimSpace(path)
		if path == "" || path == "." {
			return
		}
		path = filepath.Clean(path)
		if seen[path] {
			return
		}
		seen[path] = true
		active := !shadow && !haveActive
		if active {
			haveActive = true
		}
		out = append(out, Place{
			Path:   path,
			Note:   note,
			Active: active,
			Exists: pathExists(path),
		})
	}
	for _, c := range p.Cands {
		if c.Kind != "discover" && env.blocked(c.Unless) {
			continue
		}
		switch c.Kind {
		case "env":
			add(env.get(c.Env), c.Note, c.Shadow)
		case "envjoin":
			if base := env.get(c.Env); base != "" {
				add(filepath.Join(base, c.Join), c.Note, c.Shadow)
			}
		case "home":
			add(env.home(c.Path), c.Note, c.Shadow)
		case "xdg":
			if base := env.get("XDG_CACHE_HOME"); base != "" {
				add(filepath.Join(base, c.Path), c.Note, c.Shadow)
			}
		case "abs":
			add(c.Path, c.Note, c.Shadow)
		case "discover":
			for _, path := range discover(env, c.Discover) {
				add(path, c.Note, c.Shadow)
			}
		}
	}
	return out
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func discover(env Environ, kind string) []string {
	switch kind {
	case "lmstudio-settings":
		if p := jsonString(env.home(".lmstudio/settings.json"), "downloadsFolder"); p != "" {
			return []string{p}
		}
	case "strata-settings":
		base := env.get("XDG_CONFIG_HOME")
		if base == "" {
			base = env.home(".config")
		}
		if p := jsonString(filepath.Join(base, "strata", "settings.json"), "data_dir"); p != "" {
			return []string{p}
		}
	case "strata-documents":
		docs := env.home("Documents")
		if p := userDirs(env, "XDG_DOCUMENTS_DIR"); p != "" {
			docs = p
		}
		return []string{filepath.Join(docs, "Strata", "Models")}
	case "tensorfold-kits":
		return tensorfoldKits(env)
	case "comfy":
		return comfyModels(env)
	case "a1111":
		return a1111Models(env)
	}
	return nil
}

// jsonString reads one top-level string. Other fields, including tokens, are discarded.
func jsonString(path, key string) string {
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 || len(b) > 1<<20 {
		return ""
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

func userDirs(env Environ, key string) string {
	base := env.get("XDG_CONFIG_HOME")
	if base == "" {
		base = env.home(".config")
	}
	b, err := os.ReadFile(filepath.Join(base, "user-dirs.dirs"))
	if err != nil || len(b) > 1<<20 {
		return ""
	}
	prefix := key + "="
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		v := strings.Trim(strings.TrimPrefix(line, prefix), `"`)
		v = strings.ReplaceAll(v, "$HOME", env.Home)
		if v != "" {
			return filepath.Clean(v)
		}
	}
	return ""
}

func tensorfoldKits(env Environ) []string {
	home := hfHome(env)
	entries, err := os.ReadDir(home)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "tensorfold-") {
			out = append(out, filepath.Join(home, e.Name()))
		}
	}
	return out
}

func hfHome(env Environ) string {
	if h := env.get("HF_HOME"); h != "" {
		return h
	}
	if x := env.get("XDG_CACHE_HOME"); x != "" {
		return filepath.Join(x, "huggingface")
	}
	return env.home(".cache/huggingface")
}

func comfyModels(env Environ) []string {
	var seeds []string
	for _, rel := range []string{"ComfyUI", "comfy", "comfyui", filepath.Join("dev", "comfyui"), filepath.Join("dev", "ComfyUI")} {
		seeds = append(seeds, env.home(rel))
	}
	if dev, err := os.ReadDir(env.home("dev")); err == nil {
		for _, e := range dev {
			if e.IsDir() {
				seeds = append(seeds, env.home(filepath.Join("dev", e.Name())))
			}
		}
	}
	return installsWithModels(seeds, "main.py", "folder_paths.py")
}
func a1111Models(env Environ) []string {
	var seeds []string
	for _, rel := range []string{"stable-diffusion-webui", "stable-diffusion-webui-forge", "forge", "sd-webui", "SD.Next"} {
		seeds = append(seeds, env.home(rel))
	}
	if dev, err := os.ReadDir(env.home("dev")); err == nil {
		for _, e := range dev {
			if e.IsDir() {
				seeds = append(seeds, env.home(filepath.Join("dev", e.Name())))
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, root := range seeds {
		models := filepath.Join(root, "models")
		if !dirExists(models) {
			continue
		}
		if !fileExists(filepath.Join(root, "webui.py")) && !dirExists(filepath.Join(models, "Stable-diffusion")) {
			continue
		}
		if seen[models] {
			continue
		}
		seen[models] = true
		out = append(out, models)
	}
	return out
}

func installsWithModels(seeds []string, markers ...string) []string {
	var out []string
	seen := map[string]bool{}
	for _, root := range seeds {
		models := filepath.Join(root, "models")
		if !dirExists(models) {
			continue
		}
		marked := false
		for _, m := range markers {
			if fileExists(filepath.Join(root, m)) {
				marked = true
				break
			}
		}
		if !marked || seen[models] {
			continue
		}
		seen[models] = true
		out = append(out, models)
		for _, extra := range extraModelPaths(root) {
			extra = filepath.Clean(extra)
			if seen[extra] || !dirExists(extra) {
				continue
			}
			seen[extra] = true
			out = append(out, extra)
		}
	}
	return out
}

// extraModelPaths reads base_path and absolute folder values from a ComfyUI
// extra_model_paths.yaml next to the install.
func extraModelPaths(root string) []string {
	path := filepath.Join(root, "extra_model_paths.yaml")
	if strings.Contains(filepath.ToSlash(path), "/tests/") {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil || len(b) > 1<<20 {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") || !strings.Contains(trim, ":") {
			continue
		}
		key, val, _ := strings.Cut(trim, ":")
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if val == "" || val == "|" || val == ">" {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "base_path" {
			if !filepath.IsAbs(val) {
				val = filepath.Join(root, val)
			}
			val = filepath.Clean(val)
			if !dirExists(val) {
				continue
			}
			if dirExists(filepath.Join(val, "models")) {
				out = append(out, filepath.Join(val, "models"))
			} else {
				out = append(out, val)
			}
			continue
		}
		if filepath.IsAbs(val) {
			out = append(out, filepath.Clean(val))
		}
	}
	return out
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
