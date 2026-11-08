//go:build !unix

package main

import "os"

type inode struct {
	dev uint64
	ino uint64
}

func inodeOf(os.FileInfo) (inode, bool) { return inode{}, false }
