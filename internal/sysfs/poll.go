//go:build windows || linux || darwin

package sysfs

import (
	"github.com/AR1011/wazero/experimental/sys"
	"github.com/AR1011/wazero/internal/fsapi"
)

// poll implements `Poll` as documented on sys.File via a file descriptor.
func poll(fd uintptr, flag fsapi.Pflag, timeoutMillis int32) (ready bool, errno sys.Errno) {
	if flag != fsapi.POLLIN {
		return false, sys.ENOTSUP
	}
	fds := []pollFd{newPollFd(fd, _POLLIN, 0)}
	count, errno := _poll(fds, timeoutMillis)
	return count > 0, errno
}
