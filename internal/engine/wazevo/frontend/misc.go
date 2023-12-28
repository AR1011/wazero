package frontend

import (
	"github.com/AR1011/wazero/internal/engine/wazevo/ssa"
	"github.com/AR1011/wazero/internal/wasm"
)

func FunctionIndexToFuncRef(idx wasm.Index) ssa.FuncRef {
	return ssa.FuncRef(idx)
}
