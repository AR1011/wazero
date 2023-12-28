//go:build !linux && !darwin && !windows

package sysfs

import (
	"net"

	"github.com/AR1011/wazero/experimental/sys"
	socketapi "github.com/AR1011/wazero/internal/sock"
)

// MSG_PEEK is a filler value.
const MSG_PEEK = 0x2

func newTCPListenerFile(tl *net.TCPListener) socketapi.TCPSock {
	return &unsupportedSockFile{}
}

type unsupportedSockFile struct {
	baseSockFile
}

// Accept implements the same method as documented on socketapi.TCPSock
func (f *unsupportedSockFile) Accept() (socketapi.TCPConn, sys.Errno) {
	return nil, sys.ENOSYS
}
