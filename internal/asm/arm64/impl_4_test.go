package arm64

import (
	"encoding/hex"
	"fmt"
	"math"
	"testing"

	"github.com/AR1011/wazero/internal/asm"
	"github.com/AR1011/wazero/internal/testing/require"
)

func TestAssemblerImpl_encodeJumpToRegister(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		tests := []struct {
			n      *nodeImpl
			expErr string
		}{
			{
				n:      &nodeImpl{instruction: ADD, types: operandTypesNoneToRegister},
				expErr: "ADD is unsupported for NoneToRegister type",
			},
			{
				n:      &nodeImpl{instruction: RET, dstReg: asm.NilRegister},
				expErr: "invalid destination register: nil is not integer",
			},
			{
				n:      &nodeImpl{instruction: RET, dstReg: RegV0},
				expErr: "invalid destination register: V0 is not integer",
			},
		}

		code := asm.CodeSegment{}
		defer func() { require.NoError(t, code.Unmap()) }()

		for _, tt := range tests {
			tc := tt
			a := NewAssembler(asm.NilRegister)
			buf := code.NextCodeSection()
			err := a.encodeJumpToRegister(buf, tc.n)
			require.EqualError(t, err, tc.expErr)
		}
	})

	tests := []struct {
		name   string
		expHex string
		inst   asm.Instruction
		reg    asm.Register
	}{
		{
			name:   "B",
			inst:   B,
			reg:    RegR0,
			expHex: "00001fd6",
		},
		{
			name:   "B",
			inst:   B,
			reg:    RegR5,
			expHex: "a0001fd6",
		},
		{
			name:   "B",
			inst:   B,
			reg:    RegR30,
			expHex: "c0031fd6",
		},
		{
			name:   "RET",
			inst:   RET,
			reg:    RegR0,
			expHex: "00005fd6",
		},
		{
			name:   "RET",
			inst:   RET,
			reg:    RegR5,
			expHex: "a0005fd6",
		},
		{
			name:   "RET",
			inst:   RET,
			reg:    RegR30,
			expHex: "c0035fd6",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			code := asm.CodeSegment{}
			defer func() { require.NoError(t, code.Unmap()) }()

			a := NewAssembler(asm.NilRegister)
			buf := code.NextCodeSection()
			err := a.encodeJumpToRegister(buf, &nodeImpl{instruction: tc.inst, dstReg: tc.reg})
			require.NoError(t, err)

			actual := buf.Bytes()
			require.Equal(t, tc.expHex, hex.EncodeToString(actual))
		})
	}
}

func TestAssemblerImpl_EncodeMemoryToRegister(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		tests := []struct {
			n      *nodeImpl
			expErr string
		}{
			{
				n:      &nodeImpl{instruction: SUB, types: operandTypesMemoryToRegister},
				expErr: "SUB is unsupported for MemoryToRegister type",
			},
		}

		code := asm.CodeSegment{}
		defer func() { require.NoError(t, code.Unmap()) }()

		for _, tt := range tests {
			tc := tt
			a := NewAssembler(asm.NilRegister)
			buf := code.NextCodeSection()
			err := a.encodeMemoryToRegister(buf, tc.n)
			require.EqualError(t, err, tc.expErr)
		}
	})

	tests := []struct {
		name string
		n    *nodeImpl
		exp  []byte
	}{
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x20, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x8, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x41, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x50, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0x28, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x4f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x48, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x50, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xab, 0xfc, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x6b, 0xfb, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x6b, 0xf7, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x6b, 0xf3, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0x7, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xe0, 0x43, 0x91, 0x6b, 0xff, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRD, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0x68, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0x23, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x5f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0xb, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x40, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x41, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x50, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0x2b, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x4f, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x48, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x50, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xcb, 0xff, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x6b, 0xfb, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x6b, 0xf7, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x6b, 0xf3, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0xf8}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0x7, 0x40, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xe3, 0x43, 0x91, 0x6b, 0xff, 0x7f, 0xf9}},
		{name: "LDRD/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRD, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0x68, 0xf8}},
		{name: "LDRW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRW, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x5f, 0xb8}},
		{name: "LDRW/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRW, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x40, 0xb9}},
		{name: "LDRW/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRW, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x40, 0xb8}},
		{name: "LDRW/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRW, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x20, 0x40, 0xb8}},
		{name: "LDRW/RegisterOffset/dst=R11,base=R5,offsetReg=R12", n: &nodeImpl{instruction: LDRW, srcReg: RegR5, srcReg2: RegR12, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0x6c, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x20, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x81, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x90, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0x50, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x8f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x90, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0xa0, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xbb, 0x10, 0x40, 0x91, 0x6b, 0xfb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x6b, 0xf3, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x6b, 0xeb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0xb, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xf0, 0x43, 0x91, 0x6b, 0xfb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0xa8, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0x23, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x9f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x80, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x81, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x90, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0x53, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x8f, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x90, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0xa0, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xdb, 0x13, 0x40, 0x91, 0x6b, 0xfb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x6b, 0xf3, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x6b, 0xeb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0xb8}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0xb, 0x80, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xf3, 0x43, 0x91, 0x6b, 0xfb, 0xbf, 0xb9}},
		{name: "LDRSW/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRSW, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0xa8, 0xb8}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x20, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x81, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x90, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0xa0, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x8f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0xa0, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xbb, 0x8, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xbb, 0x18, 0x40, 0x91, 0x6b, 0xf3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xd3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xc3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0x13, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xf8, 0x43, 0x91, 0x6b, 0xf3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0xa8, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0x7, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x9f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0x23, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x80, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x81, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x90, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0xa3, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x8f, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0xa0, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xdb, 0xb, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xdb, 0x1b, 0x40, 0x91, 0x6b, 0xf3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xd3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xc3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x78}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0x13, 0x80, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xfb, 0x43, 0x91, 0x6b, 0xf3, 0xbf, 0x79}},
		{name: "LDRSHD/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRSHD, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0xa8, 0x78}},
		{name: "LDRSHW/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSHW, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0xdf, 0x78}},
		{name: "LDRSHW/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRSHW, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0xc0, 0x79}},
		{name: "LDRSHW/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRSHW, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0xc0, 0x78}},
		{name: "LDRSHW/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRSHW, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0xc0, 0x79}},
		{name: "LDRSHW/RegisterOffset/dst=R11,base=R5,offsetReg=R12", n: &nodeImpl{instruction: LDRSHW, srcReg: RegR5, srcReg2: RegR12, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0xec, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x20, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x41, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x50, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0xa0, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x4f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x60, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xbb, 0x8, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xbb, 0x18, 0x40, 0x91, 0x6b, 0xf3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xe3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xd3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x38, 0x40, 0x91, 0x6b, 0xc3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0x13, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xf8, 0x43, 0x91, 0x6b, 0xf3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRH, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0x68, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0x7, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x5f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0x23, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x40, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x41, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x50, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0xa3, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x4f, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x60, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xdb, 0xb, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xdb, 0x1b, 0x40, 0x91, 0x6b, 0xf3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xe3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xd3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x3b, 0x40, 0x91, 0x6b, 0xc3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x78}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0x13, 0x40, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xfb, 0x43, 0x91, 0x6b, 0xf3, 0x7f, 0x79}},
		{name: "LDRH/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRH, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0x68, 0x78}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x8, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0x3c, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x40, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0x3c, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x44, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x90, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0x40, 0x81, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x98, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xfc, 0x83, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xbb, 0x4, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xbb, 0x8, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xbb, 0x1c, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0xc3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0xa3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0x83, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0x23, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xfc, 0x43, 0x91, 0x6b, 0xe3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0xa8, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x7, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0xb, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0x3f, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x9f, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0x43, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0x3f, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x47, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x90, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0x43, 0x81, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x98, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xff, 0x83, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xdb, 0x7, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xdb, 0xb, 0x40, 0x91, 0x6b, 0x3, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xdb, 0x1f, 0x40, 0x91, 0x6b, 0xe3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0xc3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0xa3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0x83, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0xbb, 0x38}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0x23, 0x80, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xff, 0x43, 0x91, 0x6b, 0xe3, 0xbf, 0x39}},
		{name: "LDRSBD/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRSBD, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0xa8, 0x38}},
		{name: "LDRSBW/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRSBW, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0xc0, 0x39}},
		{name: "LDRSBW/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRSBW, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0xc0, 0x39}},
		{name: "LDRSBW/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRSBW, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x8, 0xc0, 0x39}},
		{name: "LDRSBW/RegisterOffset/dst=R11,base=R5,offsetReg=R12", n: &nodeImpl{instruction: LDRSBW, srcReg: RegR5, srcReg2: RegR12, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0xec, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -1, dstReg: RegR11}, exp: []byte{0xab, 0xf0, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x0", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 0, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x1", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1, dstReg: RegR11}, exp: []byte{0xab, 0x4, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x2", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 2, dstReg: RegR11}, exp: []byte{0xab, 0x8, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -2, dstReg: RegR11}, exp: []byte{0xab, 0xe0, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0x3c, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -15, dstReg: RegR11}, exp: []byte{0xab, 0x10, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x10", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 16, dstReg: RegR11}, exp: []byte{0xab, 0x40, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xf", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 15, dstReg: RegR11}, exp: []byte{0xab, 0x3c, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x11", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 17, dstReg: RegR11}, exp: []byte{0xab, 0x44, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -256, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x50, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x50", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 80, dstReg: RegR11}, exp: []byte{0xab, 0x40, 0x41, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: -128, dstReg: RegR11}, exp: []byte{0xab, 0x0, 0x58, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xff", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 255, dstReg: RegR11}, exp: []byte{0xab, 0xfc, 0x43, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x1000", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xbb, 0x4, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x2000", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xbb, 0x8, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xbb, 0x1c, 0x40, 0x91, 0x6b, 0xe3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xfff0", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0xc3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffe8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0xa3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffe0", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xbb, 0x3c, 0x40, 0x91, 0x6b, 0x83, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x8000000", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x40000000", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x40000008", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x40000010", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x10000004", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xab, 0x68, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0x100008", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x6b, 0x23, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=0xffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xbb, 0xfc, 0x43, 0x91, 0x6b, 0xe3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R5,offset=RegR8", n: &nodeImpl{instruction: LDRB, srcReg: RegR5, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xab, 0x68, 0x68, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -1, dstReg: RegR11}, exp: []byte{0xcb, 0xf3, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x0", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 0, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x1", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1, dstReg: RegR11}, exp: []byte{0xcb, 0x7, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x2", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 2, dstReg: RegR11}, exp: []byte{0xcb, 0xb, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -2, dstReg: RegR11}, exp: []byte{0xcb, 0xe3, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0x3f, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -15, dstReg: RegR11}, exp: []byte{0xcb, 0x13, 0x5f, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x10", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 16, dstReg: RegR11}, exp: []byte{0xcb, 0x43, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xf", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 15, dstReg: RegR11}, exp: []byte{0xcb, 0x3f, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x11", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 17, dstReg: RegR11}, exp: []byte{0xcb, 0x47, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -256, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x50, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x50", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 80, dstReg: RegR11}, exp: []byte{0xcb, 0x43, 0x41, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: -128, dstReg: RegR11}, exp: []byte{0xcb, 0x3, 0x58, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xff", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 255, dstReg: RegR11}, exp: []byte{0xcb, 0xff, 0x43, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x1000", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 4096, dstReg: RegR11}, exp: []byte{0xdb, 0x7, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x2000", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 8192, dstReg: RegR11}, exp: []byte{0xdb, 0xb, 0x40, 0x91, 0x6b, 0x3, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 32760, dstReg: RegR11}, exp: []byte{0xdb, 0x1f, 0x40, 0x91, 0x6b, 0xe3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xfff0", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 65520, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0xc3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffe8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 65512, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0xa3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffe0", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 65504, dstReg: RegR11}, exp: []byte{0xdb, 0x3f, 0x40, 0x91, 0x6b, 0x83, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x8000000", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 134217728, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x40000000", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1073741824, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x40000008", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1073741832, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1073741816, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x40000010", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1073741840, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1073741808, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 2147483640, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x10000004", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 268435460, dstReg: RegR11}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xcb, 0x6b, 0x7b, 0x38}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0x100008", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1048584, dstReg: RegR11}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x6b, 0x23, 0x40, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=0xffff8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcConst: 1048568, dstReg: RegR11}, exp: []byte{0xdb, 0xff, 0x43, 0x91, 0x6b, 0xe3, 0x7f, 0x39}},
		{name: "LDRB/RegisterOffset/dst=R11,base=R30,offset=RegR8", n: &nodeImpl{instruction: LDRB, srcReg: RegR30, srcReg2: RegR8, dstReg: RegR11}, exp: []byte{0xcb, 0x6b, 0x68, 0x38}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -1, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 0, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x1", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x2", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 2, dstReg: RegV30}, exp: []byte{0xbe, 0x20, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -2, dstReg: RegV30}, exp: []byte{0xbe, 0xe0, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xf", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 15, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -15, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x10", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 16, dstReg: RegV30}, exp: []byte{0xbe, 0x8, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xf", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 15, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x11", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 17, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x41, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -128, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x58, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -256, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x50, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x50", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 80, dstReg: RegV30}, exp: []byte{0xbe, 0x28, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: -128, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x58, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xff", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 255, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x4f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x1000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 4096, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x48, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x2000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 8192, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x50, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 32760, dstReg: RegV30}, exp: []byte{0xbe, 0xfc, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xfff0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 65520, dstReg: RegV30}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x7e, 0xfb, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffe8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 65512, dstReg: RegV30}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x7e, 0xf7, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffe0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 65504, dstReg: RegV30}, exp: []byte{0xbb, 0x20, 0x40, 0x91, 0x7e, 0xf3, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x8000000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 134217728, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x40000000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1073741824, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x40000008", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1073741832, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1073741816, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x40000010", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1073741840, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1073741808, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 2147483640, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x10000004", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 268435460, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0x100008", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1048584, dstReg: RegV30}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x7e, 0x7, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=0xffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcConst: 1048568, dstReg: RegV30}, exp: []byte{0xbb, 0xe0, 0x43, 0x91, 0x7e, 0xff, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R5,offset=RegR8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR5, srcReg2: RegR8, dstReg: RegV30}, exp: []byte{0xbe, 0x68, 0x68, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -1, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 0, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x1", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x2", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 2, dstReg: RegV30}, exp: []byte{0xde, 0x23, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -2, dstReg: RegV30}, exp: []byte{0xde, 0xe3, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xf", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 15, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -15, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x5f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x10", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 16, dstReg: RegV30}, exp: []byte{0xde, 0xb, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xf", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 15, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x40, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x11", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 17, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x41, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -128, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x58, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -256, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x50, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x50", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 80, dstReg: RegV30}, exp: []byte{0xde, 0x2b, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: -128, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x58, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xff", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 255, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x4f, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x1000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 4096, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x48, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x2000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 8192, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x50, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 32760, dstReg: RegV30}, exp: []byte{0xde, 0xff, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xfff0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 65520, dstReg: RegV30}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x7e, 0xfb, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffe8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 65512, dstReg: RegV30}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x7e, 0xf7, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffe0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 65504, dstReg: RegV30}, exp: []byte{0xdb, 0x23, 0x40, 0x91, 0x7e, 0xf3, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x8000000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 134217728, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x40000000", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1073741824, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x40000008", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1073741832, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1073741816, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x40000010", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1073741840, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1073741808, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 2147483640, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x10000004", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 268435460, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xfc}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0x100008", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1048584, dstReg: RegV30}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x7e, 0x7, 0x40, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=0xffff8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcConst: 1048568, dstReg: RegV30}, exp: []byte{0xdb, 0xe3, 0x43, 0x91, 0x7e, 0xff, 0x7f, 0xfd}},
		{name: "FLDRD/RegisterOffset/dst=V30,base=R30,offset=RegR8", n: &nodeImpl{instruction: FLDRD, srcReg: RegR30, srcReg2: RegR8, dstReg: RegV30}, exp: []byte{0xde, 0x6b, 0x68, 0xfc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffffff", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -1, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 0, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x1", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x2", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 2, dstReg: RegV30}, exp: []byte{0xbe, 0x20, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -2, dstReg: RegV30}, exp: []byte{0xbe, 0xe0, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xf", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 15, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -15, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x10", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 16, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xf", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 15, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x11", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 17, dstReg: RegV30}, exp: []byte{0xbe, 0x10, 0x41, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -128, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x58, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff00", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -256, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x50, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x50", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 80, dstReg: RegV30}, exp: []byte{0xbe, 0x50, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: -128, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x58, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xff", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 255, dstReg: RegV30}, exp: []byte{0xbe, 0xf0, 0x4f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x1000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 4096, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x50, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x2000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 8192, dstReg: RegV30}, exp: []byte{0xbe, 0x0, 0x60, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x7ff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 32760, dstReg: RegV30}, exp: []byte{0xbb, 0x10, 0x40, 0x91, 0x7e, 0xfb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xfff0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 65520, dstReg: RegV30}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x7e, 0xf3, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffe8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 65512, dstReg: RegV30}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x7e, 0xeb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffe0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 65504, dstReg: RegV30}, exp: []byte{0xbb, 0x30, 0x40, 0x91, 0x7e, 0xe3, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x8000000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 134217728, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x40000000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1073741824, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x40000008", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1073741832, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x3ffffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1073741816, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x40000010", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1073741840, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x3ffffff0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1073741808, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x7ffffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 2147483640, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x10000004", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 268435460, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xbe, 0x68, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0x100008", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1048584, dstReg: RegV30}, exp: []byte{0xbb, 0x0, 0x44, 0x91, 0x7e, 0xb, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=0xffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcConst: 1048568, dstReg: RegV30}, exp: []byte{0xbb, 0xf0, 0x43, 0x91, 0x7e, 0xfb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R5,offset=RegR8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR5, srcReg2: RegR8, dstReg: RegV30}, exp: []byte{0xbe, 0x68, 0x68, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffffff", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -1, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 0, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x1", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x2", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 2, dstReg: RegV30}, exp: []byte{0xde, 0x23, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xfffffffffffffffe", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -2, dstReg: RegV30}, exp: []byte{0xde, 0xe3, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xf", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 15, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xfffffffffffffff1", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -15, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x5f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x10", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 16, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xf", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 15, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x40, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x11", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 17, dstReg: RegV30}, exp: []byte{0xde, 0x13, 0x41, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -128, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x58, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff00", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -256, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x50, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x50", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 80, dstReg: RegV30}, exp: []byte{0xde, 0x53, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffffffffffffff80", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: -128, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x58, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xff", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 255, dstReg: RegV30}, exp: []byte{0xde, 0xf3, 0x4f, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x1000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 4096, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x50, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x2000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 8192, dstReg: RegV30}, exp: []byte{0xde, 0x3, 0x60, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x7ff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 32760, dstReg: RegV30}, exp: []byte{0xdb, 0x13, 0x40, 0x91, 0x7e, 0xfb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xfff0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 65520, dstReg: RegV30}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x7e, 0xf3, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffe8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 65512, dstReg: RegV30}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x7e, 0xeb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffe0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 65504, dstReg: RegV30}, exp: []byte{0xdb, 0x33, 0x40, 0x91, 0x7e, 0xe3, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x8000000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 134217728, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x40000000", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1073741824, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x40000008", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1073741832, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x3ffffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1073741816, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x40000010", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1073741840, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x3ffffff0", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1073741808, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x7ffffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 2147483640, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x10000004", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 268435460, dstReg: RegV30}, exp: []byte{0x1b, 0x0, 0x0, 0x18, 0xde, 0x6b, 0x7b, 0xbc}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0x100008", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1048584, dstReg: RegV30}, exp: []byte{0xdb, 0x3, 0x44, 0x91, 0x7e, 0xb, 0x40, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=0xffff8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcConst: 1048568, dstReg: RegV30}, exp: []byte{0xdb, 0xf3, 0x43, 0x91, 0x7e, 0xfb, 0x7f, 0xbd}},
		{name: "FLDRS/RegisterOffset/dst=V30,base=R30,offset=RegR8", n: &nodeImpl{instruction: FLDRS, srcReg: RegR30, srcReg2: RegR8, dstReg: RegV30}, exp: []byte{0xde, 0x6b, 0x68, 0xbc}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code := asm.CodeSegment{}
			defer func() { require.NoError(t, code.Unmap()) }()

			a := NewAssembler(RegR27)
			buf := code.NextCodeSection()
			err := a.encodeMemoryToRegister(buf, tc.n)
			require.NoError(t, err)

			err = a.Assemble(buf)
			require.NoError(t, err)

			actual := buf.Bytes()
			require.Equal(t, tc.exp, actual, hex.EncodeToString(actual))
		})
	}
}

func TestAssemblerImpl_encodeReadInstructionAddress(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tests := []struct {
			name                   string
			expADRInstructionBytes []byte
			numDummyInstructions   int
		}{
			{
				name:                   "< 8-bit offset",
				numDummyInstructions:   1,
				expADRInstructionBytes: []byte{0x77, 0x0, 0x0, 0x10},
			},
			{
				name:                   "> 8-bit offset",
				numDummyInstructions:   5000,
				expADRInstructionBytes: []byte{0x57, 0x71, 0x2, 0x10},
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				code := asm.CodeSegment{}
				defer func() { require.NoError(t, code.Unmap()) }()

				const targetBeforeInstruction, dstReg = RET, RegR23
				a := NewAssembler(asm.NilRegister)

				a.CompileReadInstructionAddress(dstReg, targetBeforeInstruction)
				adrInst := a.current
				for i := 0; i < tc.numDummyInstructions; i++ {
					a.CompileJumpToRegister(B, RegR5)
				}
				a.CompileJumpToRegister(targetBeforeInstruction, RegR25)
				a.CompileConstToRegister(MOVD, 0x3e8, RegR10) // Target.
				target := a.current

				buf := code.NextCodeSection()
				err := a.Assemble(buf)
				require.NoError(t, err)
				// The binary should start with ADR instruction.
				actual := buf.Bytes()
				require.Equal(t, tc.expADRInstructionBytes, actual[:4], hex.EncodeToString(actual))
				// Then, follow the dummy B instructions.
				pos := 4
				for i := 0; i < tc.numDummyInstructions; i++ {
					require.Equal(t,
						// A0 00 1F D6    br   x5
						[]byte{0xa0, 0x0, 0x1f, 0xd6},
						actual[pos:pos+4], hex.EncodeToString(actual))
					pos += 4
				}
				// And targetBeforeInstruction follows: "20 03 5F D6    ret  x25"
				require.Equal(t, []byte{0x20, 0x03, 0x5F, 0xd6},
					actual[pos:pos+4], hex.EncodeToString(actual))

				// After that, we end with the target instruction "movz x10, #0x3e8"
				pos += 4
				require.Equal(t, []byte{0xa, 0x7d, 0x80, 0xd2},
					actual[pos:pos+4], hex.EncodeToString(actual))

				require.Equal(t, uint64(4+tc.numDummyInstructions*4+4),
					target.offsetInBinary-adrInst.offsetInBinary)
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		code := asm.CodeSegment{}
		defer func() { require.NoError(t, code.Unmap()) }()

		a := NewAssembler(asm.NilRegister)
		a.CompileReadInstructionAddress(RegR27, NOP)
		a.CompileConstToRegister(MOVD, 1000, RegR10)

		buf := code.NextCodeSection()
		err := a.Assemble(buf)
		require.EqualError(t, err, "BUG: target instruction NOP not found for ADR")
	})
	t.Run("offset too large", func(t *testing.T) {
		for _, offset := range []int64{
			1 << 20,
			-(1 << 20) - 1,
			math.MaxInt64, math.MinInt64,
		} {
			u64 := uint64(offset)
			t.Run(fmt.Sprintf("offset=%#b", u64), func(t *testing.T) {
				code := asm.CodeSegment{}
				defer func() { require.NoError(t, code.Unmap()) }()

				a := NewAssembler(asm.NilRegister)
				a.CompileReadInstructionAddress(RegR27, RET)
				a.CompileJumpToRegister(RET, RegR25)
				a.CompileConstToRegister(MOVD, 1000, RegR10)

				buf := code.NextCodeSection()

				for n := a.root; n != nil; n = n.next {
					n.offsetInBinary = uint64(buf.Len())

					err := a.encodeNode(buf, n)
					require.NoError(t, err)
				}

				targetNode := a.current
				targetNode.offsetInBinary = u64

				n := a.adrInstructionNodes[0]
				err := a.finalizeADRInstructionNode(nil, n)
				require.EqualError(t, err, fmt.Sprintf("BUG: too large offset for ADR: %#x", u64))
			})
		}
	})
}
