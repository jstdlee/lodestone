package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLargeFilesInStoreAndHome(t *testing.T) {
	home := t.TempDir()
	blob := filepath.Join(home, ".cache", "huggingface", "hub", "models--acme--tiny", "blobs", "weights.bin")
	if err := os.MkdirAll(filepath.Dir(blob), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := truncateFile(blob, largeMin); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(home, "notes", "archive.tar")
	if err := os.MkdirAll(filepath.Dir(archive), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := truncateFile(archive, largeMin); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "notes", "readme.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(home, "credentials")
	if err := truncateFile(secret, largeMin); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := truncateFile(filepath.Join(home, "node_modules", "pkg.tar"), largeMin); err != nil {
		t.Fatal(err)
	}

	snap := Scan(Environ{Home: home, Get: func(string) string { return "" }}, nil)
	names := map[string]bool{}
	for _, e := range snap.Large {
		names[e.Name] = true
		if strings.Contains(strings.ToLower(e.Path), "credential") {
			t.Fatalf("secret listed: %s", e.Path)
		}
	}
	if !names["weights.bin"] || !names["archive.tar"] {
		t.Fatalf("large files: %+v", snap.Large)
	}
	if names["pkg.tar"] || names["readme.txt"] || names["credentials"] {
		t.Fatalf("unexpected large files: %+v", snap.Large)
	}
	var weightListed int
	for _, e := range snap.Large {
		if e.Name == "weights.bin" {
			weightListed++
		}
	}
	if weightListed != 1 {
		t.Fatalf("weights.bin listed %d times", weightListed)
	}
}

func truncateFile(path string, n int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Truncate(n)
}
