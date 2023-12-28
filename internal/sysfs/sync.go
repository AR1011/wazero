//go:build !windows

package sysfs

import (
	"os"

	"github.com/AR1011/wazero/experimental/sys"
)

func fsync(f *os.File) sys.Errno {
	return sys.UnwrapOSError(f.Sync())
}
