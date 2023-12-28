//go:build linux

package sysfs

import (
	"os"
	"syscall"

	"github.com/AR1011/wazero/experimental/sys"
)

func datasync(f *os.File) sys.Errno {
	return sys.UnwrapOSError(syscall.Fdatasync(int(f.Fd())))
}
