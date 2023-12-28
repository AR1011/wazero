package v2

import (
	"context"
	"runtime"
	"testing"

	"github.com/AR1011/wazero"
	"github.com/AR1011/wazero/api"
	"github.com/AR1011/wazero/experimental/opt"
	"github.com/AR1011/wazero/internal/integration_test/spectest"
	"github.com/AR1011/wazero/internal/platform"
)

const enabledFeatures = api.CoreFeaturesV2

func TestCompiler(t *testing.T) {
	if !platform.CompilerSupported() {
		t.Skip()
	}
	spectest.Run(t, Testcases, context.Background(), wazero.NewRuntimeConfigCompiler().WithCoreFeatures(enabledFeatures))
}

func TestInterpreter(t *testing.T) {
	spectest.Run(t, Testcases, context.Background(), wazero.NewRuntimeConfigInterpreter().WithCoreFeatures(enabledFeatures))
}

func TestWazevo(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		t.Skip()
	}
	c := opt.NewRuntimeConfigOptimizingCompiler().WithCoreFeatures(enabledFeatures)
	spectest.Run(t, Testcases, context.Background(), c)
}
