//go:build !js

package writefs

import (
	"github.com/AR1011/wazero/experimental/sys"
)

// statFields isn't used outside JS, it is only for compilation
func statFields(string) (atimeNsec, mtimeNsec int64, dev, inode uint64) {
	panic(sys.ENOSYS)
}
