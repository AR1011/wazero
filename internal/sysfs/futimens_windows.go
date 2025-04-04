package sysfs

import (
	"syscall"

	"github.com/AR1011/wazero/experimental/sys"
	"github.com/AR1011/wazero/internal/platform"
)

func utimens(path string, atim, mtim int64) sys.Errno {
	return chtimes(path, atim, mtim)
}

func futimens(fd uintptr, atim, mtim int64) error {
	// Before Go 1.20, ERROR_INVALID_HANDLE was returned for too many reasons.
	// Kick out so that callers can use path-based operations instead.
	if !platform.IsAtLeastGo120 {
		return sys.ENOSYS
	}

	// Per docs, zero isn't a valid timestamp as it cannot be differentiated
	// from nil. In both cases, it is a marker like sys.UTIME_OMIT.
	// See https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-setfiletime
	a, w := timespecToFiletime(atim, mtim)

	if a == nil && w == nil {
		return nil // both omitted, so nothing to change
	}

	// Attempt to get the stat by handle, which works for normal files
	h := syscall.Handle(fd)

	// Note: This returns ERROR_ACCESS_DENIED when the input is a directory.
	return syscall.SetFileTime(h, nil, a, w)
}

func timespecToFiletime(atim, mtim int64) (a, w *syscall.Filetime) {
	a = timespecToFileTime(atim)
	w = timespecToFileTime(mtim)
	return
}

func timespecToFileTime(tim int64) *syscall.Filetime {
	if tim == sys.UTIME_OMIT {
		return nil
	}
	ft := syscall.NsecToFiletime(tim)
	return &ft
}
