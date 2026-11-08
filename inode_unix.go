//go:build unix

package main

import (
	"os"
	"syscall"
)

type inode struct {
	dev uint64
	ino uint64
}

func inodeOf(info os.FileInfo) (inode, bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return inode{}, false
	}
	return inode{dev: uint64(st.Dev), ino: uint64(st.Ino)}, true
}
