//go:build !windows && !plan9

package sysfs

import (
	"io/fs"
	"syscall"

	experimentalsys "github.com/AR1011/wazero/experimental/sys"
	"github.com/AR1011/wazero/sys"
)

func inoFromFileInfo(_ string, info fs.FileInfo) (sys.Inode, experimentalsys.Errno) {
	switch v := info.Sys().(type) {
	case *sys.Stat_t:
		return v.Ino, 0
	case *syscall.Stat_t:
		return v.Ino, 0
	default:
		return 0, 0
	}
}
