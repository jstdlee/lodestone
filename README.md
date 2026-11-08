# Lodestone

A native window for the model files already on disk: hub caches, runner stores, loose weights, and other large files under your home directory.

```sh
go run .
```

Needs Go 1.27. The window is MyGo native UI, so it does not need a web frontend.

The window opens on the map. Category totals sit above a treemap drawn by size, and the type on each block scales with the block. Details opens the category list: pick a row to read its path, reveal it in the file manager, or copy the path. The same row lights up on the map, and a block opens that row. Search sits under the title and matches a name or a path. Rescan and the theme are the icons at the right.

The scan lists a place only when that directory or file exists. The total is unique bytes: a hub blob that also shows up as a large file is counted once. A category total is the listed size of that group, which can overlap.

```sh
go test
```
