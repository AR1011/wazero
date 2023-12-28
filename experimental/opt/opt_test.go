package opt_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/AR1011/wazero"
	"github.com/AR1011/wazero/experimental/opt"
	"github.com/AR1011/wazero/internal/testing/require"
)

func TestUseOptimizingCompiler(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		return
	}
	c := opt.NewRuntimeConfigOptimizingCompiler()
	r := wazero.NewRuntimeWithConfig(context.Background(), c)
	require.NoError(t, r.Close(context.Background()))
}
