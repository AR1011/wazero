package ssa

import (
	"testing"

	"github.com/AR1011/wazero/internal/testing/require"
)

func TestInstruction_InvertConditionalBrx(t *testing.T) {
	i := &Instruction{opcode: OpcodeBrnz}
	i.InvertBrx()
	require.Equal(t, OpcodeBrz, i.opcode)
	i.InvertBrx()
	require.Equal(t, OpcodeBrnz, i.opcode)
}
