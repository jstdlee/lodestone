package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestOpenCopyAndFilter(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(model, []byte("weight"), 0o644); err != nil {
		t.Fatal(err)
	}
	hub := filepath.Join(dir, "hub", "qwen")
	if err := os.MkdirAll(hub, 0o755); err != nil {
		t.Fatal(err)
	}
	qwen := filepath.Join(hub, "model.safetensors")
	if err := os.WriteFile(qwen, []byte("q"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := newApp()
	a.home = filepath.Dir(dir)
	a.snap = Snapshot{
		Unique: 100,
		Stores: []Store{
			{
				ID: "lmstudio", Name: "LM Studio", Tools: "lms", Category: "Runners",
				Blurb: "GGUF files.", Path: dir, Active: true, Bytes: 80,
				Entries: []Entry{{Name: "publisher/model", Path: model, Bytes: 80, Detail: "gguf"}},
			},
			{
				ID: "hf-hub", Name: "Hugging Face Hub", Tools: "hf", Category: "Hub",
				Path: hub, Active: true, Bytes: 20,
				Entries: []Entry{{Name: "Qwen/Qwen3-0.6B", Path: qwen, Bytes: 20, Detail: "model"}},
			},
		},
	}
	var opened, copied string
	a.reveal = func(p string) error { opened = p; return nil }
	a.copyText = func(s string) error { copied = s; return nil }
	a.selectedKey = "lmstudio\x00" + model

	tt := ui.NewTester(a.view, 1100, 740)
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("publisher/model") || !tt.HasText("Qwen/Qwen3-0.6B") {
		t.Fatalf("rows missing: %q", tt.Texts())
	}
	if err := tt.Click("Open location"); err != nil {
		t.Fatal(err)
	}
	if opened != model {
		t.Fatalf("opened %q", opened)
	}
	if err := tt.Click("Copy path"); err != nil {
		t.Fatal(err)
	}
	if copied != model {
		t.Fatalf("copied %q", copied)
	}

	if err := tt.Click("Runners"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("Qwen/Qwen3-0.6B") || !tt.HasText("publisher/model") {
		t.Fatalf("category filter: %q", tt.Texts())
	}
	if err := tt.Click("All"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Search"); err != nil {
		t.Fatal(err)
	}
	tt.Type("nothing-here")
	if !tt.HasText("No matches") || tt.HasText("publisher/model") {
		t.Fatalf("search: %q", tt.Texts())
	}
}

func TestSearchPathAndTheme(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "unique-cache", "weights.bin")
	if err := os.MkdirAll(filepath.Dir(model), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(model, []byte("w"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := newApp()
	a.home = dir
	a.snap = Snapshot{Stores: []Store{{
		ID: "hf-hub", Name: "Hugging Face Hub", Category: "Hub", Path: dir, Bytes: 10,
		Entries: []Entry{{Name: "publisher/model", Path: model, Bytes: 10}},
	}}}
	tt := ui.NewTester(a.view, 1100, 740)
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Search"); err != nil {
		t.Fatal(err)
	}
	tt.Type("unique-cache")
	if !tt.HasText("publisher/model") {
		t.Fatalf("path search: %q", tt.Texts())
	}
	if a.dark {
		t.Fatal("started dark")
	}
	if err := tt.Click("Theme"); err != nil {
		t.Fatal(err)
	}
	if !a.dark {
		t.Fatal("theme did not switch")
	}
}

func TestOpenMissingDoesNothing(t *testing.T) {
	a := newApp()
	opened := false
	a.reveal = func(string) error { opened = true; return nil }
	a.openRow("missing", "", false)
	if opened {
		t.Fatal("opened a row with no path")
	}
}

func TestLargeCategory(t *testing.T) {
	a := newApp()
	a.snap.Large = []Entry{{Name: "archive.tar", Path: "/tmp/archive.tar", Bytes: largeMin, Detail: "tar"}}
	a.snap.Unique = largeMin
	tt := ui.NewTester(a.view, 1100, 740)
	if err := tt.Click("Large files"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("archive.tar") {
		t.Fatalf("large file hidden: %q", tt.Texts())
	}
}

func TestRescanReplacesRows(t *testing.T) {
	a := newApp()
	a.scan = func(func(string)) Snapshot {
		return Snapshot{Unique: 5, Stores: []Store{{
			ID: "whisper", Name: "Whisper", Category: "Image & audio", Tools: "whisper",
			Path: "/tmp/whisper", Bytes: 5, Active: true,
			Entries: []Entry{{Name: "tiny.pt", Path: "/tmp/whisper/tiny.pt", Bytes: 5}},
		}}}
	}
	tt := ui.NewTester(a.view, 1100, 740)
	if err := tt.Click("Rescan"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("tiny.pt") {
		t.Fatalf("after rescan: %q", tt.Texts())
	}
	if a.scanning {
		t.Fatal("scan still marked running")
	}
}
