//go:build !windows && !plan9

package sysfs

import (
	"syscall"

	"github.com/AR1011/wazero/experimental/sys"
)

func setNonblock(fd uintptr, enable bool) sys.Errno {
	return sys.UnwrapOSError(syscall.SetNonblock(int(fd), enable))
}

func isNonblock(f *osFile) bool {
	return f.flag&sys.O_NONBLOCK == sys.O_NONBLOCK
}
