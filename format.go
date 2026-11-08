package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func formatBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div := int64(unit)
	exp := 0
	for v := n / unit; v >= unit && exp < 4; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func tilde(home, path string) string {
	if path == "" {
		return ""
	}
	home = filepath.Clean(home)
	path = filepath.Clean(path)
	if home == "" || home == "." {
		return path
	}
	if path == home {
		return "~"
	}
	sep := string(filepath.Separator)
	if strings.HasPrefix(path, home+sep) {
		return "~" + path[len(home):]
	}
	return path
}
