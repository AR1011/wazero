package vs

import (
	"bytes"
	_ "embed"
	"fmt"
	"testing"

	"github.com/AR1011/wazero/internal/testing/require"
)

var (
	// allocationWasm is compiled from ../../../examples/allocation/tinygo/testdata/src/greet.go
	// We can't use go:embed as it is outside this directory. Copying it isn't ideal due to size and drift.
	allocationWasmPath = "../../../examples/allocation/tinygo/testdata/greet.wasm"
	allocationWasm     []byte
	allocationParam    = "wazero"
	allocationResult   = []byte("wasm >> Hello, wazero!")
	allocationConfig   *RuntimeConfig
)

func init() {
	allocationWasm = readRelativeFile(allocationWasmPath)
	allocationConfig = &RuntimeConfig{
		ModuleName: "greet",
		ModuleWasm: allocationWasm,
		FuncNames:  []string{"malloc", "free", "greet"},
		NeedsWASI:  true, // Needed for TinyGo
		LogFn: func(buf []byte) error {
			if !bytes.Equal(allocationResult, buf) {
				return fmt.Errorf("expected %q, but was %q", allocationResult, buf)
			}
			return nil
		},
	}
}

func allocationCall(m Module, _ int) error {
	nameSize := uint32(len(allocationParam))
	// Instead of an arbitrary memory offset, use TinyGo's allocator. Notice
	// there is nothing string-specific in this allocation function. The same
	// function could be used to pass binary serialized data to Wasm.
	namePtr, err := m.CallI32_I32(testCtx, "malloc", nameSize)
	if err != nil {
		return err
	}

	// The pointer is a linear memory offset, which is where we write the name.
	if err = m.WriteMemory(namePtr, []byte(allocationParam)); err != nil {
		return err
	}

	// Now, we can call "greeting", which reads the string we wrote to memory!
	fnErr := m.CallI32I32_V(testCtx, "greet", namePtr, nameSize)
	if fnErr != nil {
		return fnErr
	}

	// This pointer was allocated by TinyGo, but owned by Go, So, we have to
	// deallocate it when finished
	if err := m.CallI32_V(testCtx, "free", namePtr); err != nil {
		return err
	}

	return nil
}

func RunTestAllocation(t *testing.T, runtime func() Runtime) {
	testCall(t, runtime, allocationConfig, testAllocationCall)
}

func testAllocationCall(t *testing.T, m Module, instantiation, iteration int) {
	err := allocationCall(m, iteration)
	require.NoError(t, err, "instantiation[%d] iteration[%d] failed: %v", instantiation, iteration, err)
}

func RunTestBenchmarkAllocation_Call_CompilerFastest(t *testing.T, vsRuntime Runtime) {
	runTestBenchmark_Call_CompilerFastest(t, allocationConfig, "Allocation", allocationCall, vsRuntime)
}

func RunBenchmarkAllocation(b *testing.B, runtime func() Runtime) {
	benchmark(b, runtime, allocationConfig, allocationCall)
}
