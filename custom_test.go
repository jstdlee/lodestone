package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCustomAddDuplicateDelete(t *testing.T) {
	home := t.TempDir()
	a := newApp()
	a.home = home
	a.customPath = filepath.Join(home, "cfg", "dirs.json")
	a.scan = func(func(string)) Snapshot { return Snapshot{} }

	a.beginAdd()
	if err := a.confirmPending(); err == nil {
		t.Fatal("empty path confirmed")
	}
	a.draftName = "Weights"
	a.draftPath = "pile"
	pile := filepath.Join(home, "pile")
	if err := os.MkdirAll(pile, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := a.confirmPending(); err != nil {
		t.Fatal(err)
	}
	if len(a.customs) != 1 || a.customs[0].Path != pile || a.customs[0].Name != "Weights" {
		t.Fatalf("%+v", a.customs)
	}
	if a.drafting {
		t.Fatal("draft still open")
	}

	id := a.customs[0].ID
	a.beginDuplicate(id)
	if a.draftName != "Weights copy" || a.draftPath != pile {
		t.Fatalf("draft %q %q", a.draftName, a.draftPath)
	}
	if err := a.confirmPending(); err == nil {
		t.Fatal("duplicate path confirmed")
	}
	other := filepath.Join(home, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	a.draftPath = other
	if err := a.confirmPending(); err != nil {
		t.Fatal(err)
	}
	if len(a.customs) != 2 {
		t.Fatalf("%+v", a.customs)
	}

	file := filepath.Join(home, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	a.beginAdd()
	a.draftPath = file
	if err := a.confirmPending(); err == nil {
		t.Fatal("file confirmed")
	}
	a.cancelPending()

	a.askDelete(id)
	if a.confirmDel != id {
		t.Fatal("delete was not armed")
	}
	a.askDelete(id)
	if a.confirmDel != "" {
		t.Fatal("second delete did not disarm")
	}
	a.askDelete(id)
	if err := a.confirmPending(); err != nil {
		t.Fatal(err)
	}
	if len(a.customs) != 1 || a.customs[0].Path != other {
		t.Fatalf("%+v", a.customs)
	}

	b := newApp()
	b.home = home
	b.customPath = a.customPath
	if err := b.loadCustoms(); err != nil {
		t.Fatal(err)
	}
	if len(b.customs) != 1 || b.customs[0].Path != other {
		t.Fatalf("reloaded %+v", b.customs)
	}
}

func TestCustomDirScan(t *testing.T) {
	home := t.TempDir()
	parent := filepath.Join(home, "pile")
	child := filepath.Join(parent, "inner")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := make([]byte, 64)
	if err := os.WriteFile(filepath.Join(child, "w.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	env := Environ{Home: home, Get: func(string) string { return "" }}
	extra := []CustomDir{
		{ID: "p", Name: "Pile", Path: parent},
		{ID: "c", Name: "Inner", Path: child},
		{ID: "m", Name: "Gone", Path: filepath.Join(home, "missing")},
	}
	parentOnly := scanDirs(env, extra[:1], home, nil)
	snap := scanDirs(env, extra, home, nil)
	if snap.Unique != parentOnly.Unique {
		t.Fatalf("nested directory counted twice: %d vs %d", snap.Unique, parentOnly.Unique)
	}
	var pile, inner, gone *Store
	for i := range snap.Stores {
		st := &snap.Stores[i]
		switch st.ID {
		case "custom:p":
			pile = st
		case "custom:c":
			inner = st
		case "custom:m":
			gone = st
		}
	}
	if pile == nil || pile.Bytes != 64 || len(pile.Entries) != 1 || pile.Entries[0].Name != "inner" {
		t.Fatalf("pile %+v", pile)
	}
	if inner == nil || inner.Bytes != 64 || inner.Note != "inside Pile" {
		t.Fatalf("inner %+v", inner)
	}
	if gone == nil || gone.Active || gone.Note != "missing" {
		t.Fatalf("gone %+v", gone)
	}
	rows := diskRows(snap)
	var sawMissing bool
	for _, r := range rows {
		if r.name == "Gone" {
			sawMissing = true
			if r.exists {
				t.Fatal("missing dir looked present")
			}
		}
	}
	if !sawMissing {
		t.Fatal("missing custom dir hidden")
	}
}

func TestCustomDirectoryIcons(t *testing.T) {
	home := t.TempDir()
	weights := filepath.Join(home, "weights")
	if err := os.MkdirAll(weights, 0o755); err != nil {
		t.Fatal(err)
	}
	a := newApp()
	a.home = home
	a.customPath = filepath.Join(home, "cfg", "dirs.json")
	a.scan = func(func(string)) Snapshot { return Snapshot{} }
	a.pickDir = func() (string, error) { return weights, nil }
	tt := ui.NewTester(a.view, 1180, 780)
	if err := tt.Click("Directories"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("None yet.") {
		t.Fatalf("panel missing: %q", tt.Texts())
	}
	if err := tt.Click("Add directory"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Choose folder"); err != nil {
		t.Fatal(err)
	}
	if a.draftPath != weights || a.draftName != "weights" {
		t.Fatalf("picked %q %q", a.draftName, a.draftPath)
	}
	if err := tt.Click("Directory name"); err != nil {
		t.Fatal(err)
	}
	tt.Type(" box")
	if err := tt.Click("Confirm"); err != nil {
		t.Fatal(err)
	}
	if len(a.customs) != 1 || a.customs[0].Name != "weights box" {
		t.Fatalf("%+v", a.customs)
	}
	if !tt.HasText("weights box") {
		t.Fatalf("row missing: %q", tt.Texts())
	}
	if err := tt.Click("Duplicate weights box"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Confirm"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("already listed") {
		t.Fatalf("duplicate not refused: %q", tt.Texts())
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Delete weights box"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Remove?") {
		t.Fatalf("delete not armed: %q", tt.Texts())
	}
	if err := tt.Click("Confirm"); err != nil {
		t.Fatal(err)
	}
	if len(a.customs) != 0 {
		t.Fatalf("not removed: %+v", a.customs)
	}
	if !tt.HasText("None yet.") {
		t.Fatalf("empty state missing: %q", tt.Texts())
	}
}
