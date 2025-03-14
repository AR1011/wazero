package gojs_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AR1011/wazero"
	"github.com/AR1011/wazero/experimental"
	"github.com/AR1011/wazero/experimental/logging"
	"github.com/AR1011/wazero/internal/gojs/config"
	"github.com/AR1011/wazero/internal/testing/require"
)

func Test_exit(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	loggingCtx := context.WithValue(testCtx, experimental.FunctionListenerFactoryKey{},
		logging.NewHostLoggingListenerFactory(&log, logging.LogScopeProc))

	stdout, stderr, err := compileAndRun(loggingCtx, "exit", defaultConfig)

	require.Zero(t, stderr)
	require.EqualError(t, err, `module closed with exit_code(255)`)
	require.Zero(t, stdout)
	require.Equal(t, `==> go.runtime.wasmExit(code=255)
<==
`, logString(log)) // Note: gojs doesn't panic on exit, so you see "<=="
}

func Test_goroutine(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := compileAndRun(testCtx, "goroutine", defaultConfig)

	require.Zero(t, stderr)
	require.NoError(t, err)
	require.Equal(t, `producer
consumer
`, stdout)
}

func Test_mem(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	loggingCtx := context.WithValue(testCtx, experimental.FunctionListenerFactoryKey{},
		logging.NewHostLoggingListenerFactory(&log, logging.LogScopeMemory))

	stdout, stderr, err := compileAndRun(loggingCtx, "mem", defaultConfig)

	require.Zero(t, stderr)
	require.NoError(t, err)
	require.Zero(t, stdout)

	// The memory view is reset at least once.
	require.Contains(t, logString(log), `==> go.runtime.resetMemoryDataView()
<==
`)
}

func Test_stdio(t *testing.T) {
	t.Parallel()

	input := "stdin\n"
	stdout, stderr, err := compileAndRun(testCtx, "stdio", func(moduleConfig wazero.ModuleConfig) (wazero.ModuleConfig, *config.Config) {
		return defaultConfig(moduleConfig.WithStdin(strings.NewReader(input)))
	})

	require.Equal(t, "stderr 6\n", stderr)
	require.NoError(t, err)
	require.Equal(t, "stdout 6\n", stdout)
}

func Test_stdio_large(t *testing.T) {
	t.Parallel()

	// Large stdio will trigger GC which will trigger events.
	var log bytes.Buffer
	loggingCtx := context.WithValue(testCtx, experimental.FunctionListenerFactoryKey{},
		logging.NewHostLoggingListenerFactory(&log, logging.LogScopePoll))

	size := 2 * 1024 * 1024 // 2MB
	input := make([]byte, size)
	stdout, stderr, err := compileAndRun(loggingCtx, "stdio", func(moduleConfig wazero.ModuleConfig) (wazero.ModuleConfig, *config.Config) {
		return defaultConfig(moduleConfig.WithStdin(bytes.NewReader(input)))
	})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("stderr %d\n", size), stderr)
	require.Equal(t, fmt.Sprintf("stdout %d\n", size), stdout)

	// There's no guarantee of a timeout event (in Go 1.21 there isn't), so we
	// don't verify this. gojs is in maintenance mode until it is removed after
	// Go 1.22 is out.
}

func Test_gc(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := compileAndRun(testCtx, "gc", defaultConfig)

	require.NoError(t, err)
	require.Equal(t, "", stderr)
	require.Equal(t, "before gc\nafter gc\n", stdout)
}
