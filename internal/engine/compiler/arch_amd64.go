package compiler

import (
	"github.com/AR1011/wazero/internal/asm"
	"github.com/AR1011/wazero/internal/asm/amd64"
)

// init initializes variables for the amd64 architecture
func init() {
	newArchContext = newArchContextImpl
	registerNameFn = amd64.RegisterName
	unreservedGeneralPurposeRegisters = amd64UnreservedGeneralPurposeRegisters
	unreservedVectorRegisters = amd64UnreservedVectorRegisters
}

// archContext is embedded in callEngine in order to store architecture-specific data.
// For amd64, this is empty.
type archContext struct{}

// newArchContextImpl implements newArchContext for amd64 architecture.
func newArchContextImpl() (ret archContext) { return }

// newCompiler returns a new compiler interface which can be used to compile the given function instance.
// Note: ir param can be nil for host functions.
func newCompiler() compiler {
	return newAmd64Compiler()
}

func registerMaskShift(r asm.Register) int {
	return int(r - amd64.RegAX)
}

func registerFromMaskShift(s int) asm.Register {
	return amd64.RegAX + asm.Register(s)
}
