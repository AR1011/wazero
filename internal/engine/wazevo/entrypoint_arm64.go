package wazevo

import _ "unsafe"

// entrypoint is implemented by the backend.
//
//go:linkname entrypoint github.com/AR1011/wazero/internal/engine/wazevo/backend/isa/arm64.entrypoint
func entrypoint(preambleExecutable, functionExecutable *byte, executionContextPtr uintptr, moduleContextPtr *byte, paramResultStackPtr *uint64, goAllocatedStackSlicePtr uintptr)

// entrypoint is implemented by the backend.
//
//go:linkname afterGoFunctionCallEntrypoint github.com/AR1011/wazero/internal/engine/wazevo/backend/isa/arm64.afterGoFunctionCallEntrypoint
func afterGoFunctionCallEntrypoint(executable *byte, executionContextPtr uintptr, stackPointer uintptr)
