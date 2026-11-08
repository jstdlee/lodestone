package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatAndTilde(t *testing.T) {
	if got := formatBytes(1536); got != "1.5 KB" {
		t.Fatalf("1536 -> %q", got)
	}
	if got := formatBytes(0); got != "0 B" {
		t.Fatalf("0 -> %q", got)
	}
	home := filepath.Join(string(filepath.Separator), "home", "ada")
	if got := tilde(home, filepath.Join(home, "models", "a")); got != "~/models/a" && got != "~\\models\\a" {
		t.Fatalf("tilde %q", got)
	}
}

func TestResolveHFPrefersHubCache(t *testing.T) {
	home := t.TempDir()
	env := Environ{Home: home, Get: func(k string) string {
		if k == "HF_HUB_CACHE" {
			return filepath.Join(home, "custom-hub")
		}
		return ""
	}}
	var hf Pattern
	for _, p := range Patterns {
		if p.ID == "hf-hub" {
			hf = p
			break
		}
	}
	places := resolve(env, hf)
	if len(places) == 0 || !places[0].Active || places[0].Path != filepath.Join(home, "custom-hub") {
		t.Fatalf("active place %+v", places)
	}
	var shadow bool
	for _, pl := range places {
		if strings.Contains(pl.Note, "leftover") {
			shadow = true
		}
	}
	if !shadow {
		t.Fatalf("leftover default missing: %+v", places)
	}
}

func TestScanFixture(t *testing.T) {
	home := t.TempDir()
	secret := "SECRETTOKENVALUE1234567890"
	hub := filepath.Join(home, ".cache", "huggingface", "hub")
	repo := filepath.Join(hub, "models--Qwen--Qwen3-0.6B")
	blob := filepath.Join(repo, "blobs", "abc")
	if err := os.MkdirAll(filepath.Dir(blob), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blob, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, "snapshots", "main", "model.safetensors")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(blob, link); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hub, "stored_tokens"), []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(repo, "blobs", "abc-link")
	if err := os.Link(blob, other); err != nil {
		t.Fatal(err)
	}

	ollama := filepath.Join(home, ".ollama", "models")
	manifest := filepath.Join(ollama, "manifests", "registry.ollama.ai", "library", "tiny", "latest")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ollama, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ollama, "blobs", "sha256-abc"), bytesOf(40), 0o644); err != nil {
		t.Fatal(err)
	}
	body := `{"layers":[{"digest":"sha256:abc","size":999999}]}`
	if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	comfy := filepath.Join(home, "dev", "comfyui")
	if err := os.MkdirAll(filepath.Join(comfy, "models", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(comfy, "main.py"), []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(comfy, "models", "checkpoints", "demo.safetensors"), bytesOf(20), 0o644); err != nil {
		t.Fatal(err)
	}
	central := filepath.Join(home, "central")
	if err := os.MkdirAll(filepath.Join(central, "models", "loras"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(central, "models", "loras", "x.safetensors"), bytesOf(8), 0o644); err != nil {
		t.Fatal(err)
	}
	yaml := "comfyui:\n  base_path: " + central + "\n"
	if err := os.WriteFile(filepath.Join(comfy, "extra_model_paths.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := filepath.Join(home, ".unsloth", "studio", "cache")
	nested := filepath.Join(cache, "huggingface", "hub", "models--org--name", "blobs")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "note.bin"), bytesOf(50), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "w"), bytesOf(80), 0o644); err != nil {
		t.Fatal(err)
	}

	loose := filepath.Join(home, "weights", "stray.safetensors")
	if err := os.MkdirAll(filepath.Dir(loose), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(loose)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(looseMin); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := os.MkdirAll(filepath.Join(home, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	skip := filepath.Join(home, "node_modules", "skip.safetensors")
	sf, err := os.Create(skip)
	if err != nil {
		t.Fatal(err)
	}
	if err := sf.Truncate(looseMin); err != nil {
		t.Fatal(err)
	}
	sf.Close()

	settings := filepath.Join(home, ".lmstudio", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	lms := filepath.Join(home, "lms-models", "pub", "model")
	if err := os.MkdirAll(lms, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lms, "m.gguf"), bytesOf(12), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := `{"downloadsFolder":"` + lms + `","hfDownloadToken":"` + secret + `"}`
	if err := os.WriteFile(settings, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	env := Environ{Home: home, Get: func(string) string { return "" }}
	var phases []string
	snap := Scan(env, func(p string) { phases = append(phases, p) })
	report := snapText(snap)
	if strings.Contains(report, secret) || strings.Contains(report, "stored_tokens") || strings.Contains(report, "hfDownloadToken") {
		t.Fatalf("secret leaked into\n%s", report)
	}
	if !strings.Contains(report, "Qwen/Qwen3-0.6B") {
		t.Fatalf("hub name missing\n%s", report)
	}
	if !strings.Contains(report, "tiny:latest") {
		t.Fatalf("ollama name missing\n%s", report)
	}
	if !strings.Contains(report, "stray.safetensors") {
		t.Fatalf("loose file missing\n%s", report)
	}
	if strings.Contains(report, "skip.safetensors") {
		t.Fatalf("node_modules was scanned\n%s", report)
	}
	var hubBytes, ollamaBytes int64
	var sawCentral, sawNestedHub bool
	for _, st := range snap.Stores {
		switch st.ID {
		case "hf-hub":
			hubBytes = st.Bytes
		case "ollama":
			if st.Path != ollama {
				continue
			}
			ollamaBytes = st.Bytes
			var found bool
			for _, e := range st.Entries {
				if e.Name == "tiny:latest" && e.Bytes == 40 && e.Path == filepath.Join(ollama, "blobs", "sha256-abc") {
					found = true
				}
			}
			if !found {
				t.Fatalf("ollama entries %+v", st.Entries)
			}
		case "comfyui":
			if st.Path == filepath.Join(central, "models") {
				sawCentral = true
			}
		case "unsloth-hub":
			sawNestedHub = true
		}
	}
	// 10-byte blob, hard-linked, must count twice in the store and once uniquely.
	if hubBytes != 20 {
		t.Fatalf("hub bytes %d, want 20", hubBytes)
	}
	if ollamaBytes < 40 {
		t.Fatalf("ollama bytes %d", ollamaBytes)
	}
	if !sawCentral {
		t.Fatalf("comfy extra root missing: %+v", snap.Stores)
	}
	if sawNestedHub {
		t.Fatal("nested unsloth hub was measured separately")
	}
	if snap.Unique < looseMin {
		t.Fatalf("unique %d", snap.Unique)
	}
	if len(phases) == 0 {
		t.Fatal("no progress")
	}
}

func TestJSONStringIgnoresOtherKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"downloadsFolder":"/tmp/lms","token":"nope"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := jsonString(path, "downloadsFolder"); got != "/tmp/lms" {
		t.Fatalf("folder %q", got)
	}
	if got := jsonString(path, "missing"); got != "" {
		t.Fatalf("missing %q", got)
	}
}

func TestOllamaRowOpensLargestBlob(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "manifests", "registry.ollama.ai", "tinyrick", "demo", "Q4_K_P")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}
	model := filepath.Join(root, "blobs", "sha256-model")
	proj := filepath.Join(root, "blobs", "sha256-proj")
	if err := os.WriteFile(model, bytesOf(100), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proj, bytesOf(10), 0o644); err != nil {
		t.Fatal(err)
	}
	body := `{"layers":[{"digest":"sha256:model","size":1},{"digest":"sha256:proj","size":999},{"digest":"sha256:missing","size":50}]}`
	if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	entries := ollamaEntries(root, nil, map[string]string{
		filepath.Join("manifests", "registry.ollama.ai", "tinyrick", "demo", "Q4_K_P"): "tinyrick/demo:Q4_K_P",
	})
	if len(entries) != 1 {
		t.Fatalf("entries %+v", entries)
	}
	e := entries[0]
	if e.Path != model {
		t.Fatalf("path %s, want the model blob", e.Path)
	}
	if e.Bytes != 160 {
		t.Fatalf("bytes %d, want model+projector+declared missing layer", e.Bytes)
	}
	if e.Detail != "weight" {
		t.Fatalf("detail %q", e.Detail)
	}
}

func bytesOf(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return b
}

func snapText(snap Snapshot) string {
	var b strings.Builder
	for _, st := range snap.Stores {
		b.WriteString(st.Name)
		b.WriteByte(' ')
		b.WriteString(st.Path)
		b.WriteByte('\n')
		for _, e := range st.Entries {
			b.WriteString(e.Name)
			b.WriteByte(' ')
			b.WriteString(e.Path)
			b.WriteByte('\n')
		}
	}
	for _, e := range snap.Loose {
		b.WriteString(e.Name)
		b.WriteByte(' ')
		b.WriteString(e.Path)
		b.WriteByte('\n')
	}
	for _, e := range snap.Large {
		b.WriteString(e.Name)
		b.WriteByte(' ')
		b.WriteString(e.Path)
		b.WriteByte('\n')
	}
	for _, w := range snap.Warnings {
		b.WriteString(w)
		b.WriteByte('\n')
	}
	return b.String()
}
