package gojs_test

import (
	"os"
	"testing"

	"github.com/AR1011/wazero"
	"github.com/AR1011/wazero/internal/gojs/config"
	"github.com/AR1011/wazero/internal/testing/require"
)

func Test_process(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Chdir("/.."))
	stdout, stderr, err := compileAndRun(testCtx, "process", func(moduleConfig wazero.ModuleConfig) (wazero.ModuleConfig, *config.Config) {
		return defaultConfig(moduleConfig.WithFS(testFS))
	})

	require.Zero(t, stderr)
	require.NoError(t, err)
	require.Equal(t, `syscall.Getpid()=1
syscall.Getppid()=0
syscall.Getuid()=0
syscall.Getgid()=0
syscall.Geteuid()=0
syscall.Umask(0077)=0o22
syscall.Getgroups()=[0]
os.FindProcess(1).Pid=1
wd ok
Not a directory
`, stdout)
}
