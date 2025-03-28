package compiler

import (
	"fmt"
	"math"
	"testing"

	"github.com/AR1011/wazero/internal/asm"
	"github.com/AR1011/wazero/internal/testing/require"
	"github.com/AR1011/wazero/internal/wasm"
	"github.com/AR1011/wazero/internal/wazeroir"
)

func TestCompiler_compileReinterpret(t *testing.T) {
	for _, kind := range []wazeroir.OperationKind{
		wazeroir.OperationKindF32ReinterpretFromI32,
		wazeroir.OperationKindF64ReinterpretFromI64,
		wazeroir.OperationKindI32ReinterpretFromF32,
		wazeroir.OperationKindI64ReinterpretFromF64,
	} {
		kind := kind
		t.Run(kind.String(), func(t *testing.T) {
			for _, originOnStack := range []bool{false, true} {
				originOnStack := originOnStack
				t.Run(fmt.Sprintf("%v", originOnStack), func(t *testing.T) {
					for _, v := range []uint64{
						0, 1, 1 << 16, 1 << 31, 1 << 32, 1 << 63,
						math.MaxInt32, math.MaxUint32, math.MaxUint64,
					} {
						v := v
						t.Run(fmt.Sprintf("%d", v), func(t *testing.T) {
							env := newCompilerEnvironment()
							compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
							err := compiler.compilePreamble()
							require.NoError(t, err)

							if originOnStack {
								loc := compiler.runtimeValueLocationStack().pushRuntimeValueLocationOnStack()
								env.stack()[loc.stackPointer] = v
								env.setStackPointer(1)
							}

							var is32Bit bool
							switch kind {
							case wazeroir.OperationKindF32ReinterpretFromI32:
								is32Bit = true
								if !originOnStack {
									err = compiler.compileConstI32(operationPtr(wazeroir.NewOperationConstI32(uint32(v))))
									require.NoError(t, err)
								}
								err = compiler.compileF32ReinterpretFromI32()
								require.NoError(t, err)
							case wazeroir.OperationKindF64ReinterpretFromI64:
								if !originOnStack {
									err = compiler.compileConstI64(operationPtr(wazeroir.NewOperationConstI64(v)))
									require.NoError(t, err)
								}
								err = compiler.compileF64ReinterpretFromI64()
								require.NoError(t, err)
							case wazeroir.OperationKindI32ReinterpretFromF32:
								is32Bit = true
								if !originOnStack {
									err = compiler.compileConstF32(operationPtr(wazeroir.NewOperationConstF32(math.Float32frombits(uint32(v)))))
									require.NoError(t, err)
								}
								err = compiler.compileI32ReinterpretFromF32()
								require.NoError(t, err)
							case wazeroir.OperationKindI64ReinterpretFromF64:
								if !originOnStack {
									err = compiler.compileConstF64(operationPtr(wazeroir.NewOperationConstF64(math.Float64frombits(v))))
									require.NoError(t, err)
								}
								err = compiler.compileI64ReinterpretFromF64()
								require.NoError(t, err)
							default:
								t.Fail()
							}

							err = compiler.compileReturnFunction()
							require.NoError(t, err)

							code := asm.CodeSegment{}
							defer func() { require.NoError(t, code.Unmap()) }()

							// Generate and run the code under test.
							_, err = compiler.compile(code.NextCodeSection())
							require.NoError(t, err)
							env.exec(code.Bytes())

							// Reinterpret must preserve the bit-pattern.
							if is32Bit {
								require.Equal(t, uint32(v), env.stackTopAsUint32())
							} else {
								require.Equal(t, v, env.stackTopAsUint64())
							}
						})
					}
				})
			}
		})
	}
}

func TestCompiler_compileExtend(t *testing.T) {
	for _, signed := range []bool{false, true} {
		signed := signed
		t.Run(fmt.Sprintf("signed=%v", signed), func(t *testing.T) {
			for _, v := range []uint32{
				0, 1, 1 << 14, 1 << 31, math.MaxUint32, 0xFFFFFFFF, math.MaxInt32,
			} {
				v := v
				t.Run(fmt.Sprintf("%v", v), func(t *testing.T) {
					env := newCompilerEnvironment()
					compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
					err := compiler.compilePreamble()
					require.NoError(t, err)

					// Setup the promote target.
					err = compiler.compileConstI32(operationPtr(wazeroir.NewOperationConstI32(v)))
					require.NoError(t, err)

					err = compiler.compileExtend(operationPtr(wazeroir.NewOperationExtend(signed)))
					require.NoError(t, err)

					err = compiler.compileReturnFunction()
					require.NoError(t, err)

					code := asm.CodeSegment{}
					defer func() { require.NoError(t, code.Unmap()) }()

					// Generate and run the code under test.
					_, err = compiler.compile(code.NextCodeSection())
					require.NoError(t, err)
					env.exec(code.Bytes())

					require.Equal(t, uint64(1), env.stackPointer())
					if signed {
						expected := int64(int32(v))
						require.Equal(t, expected, env.stackTopAsInt64())
					} else {
						expected := uint64(uint32(v))
						require.Equal(t, expected, env.stackTopAsUint64())
					}
				})
			}
		})
	}
}

func TestCompiler_compileITruncFromF(t *testing.T) {
	tests := []struct {
		outputType  wazeroir.SignedInt
		inputType   wazeroir.Float
		nonTrapping bool
	}{
		{outputType: wazeroir.SignedInt32, inputType: wazeroir.Float32},
		{outputType: wazeroir.SignedInt32, inputType: wazeroir.Float64},
		{outputType: wazeroir.SignedInt64, inputType: wazeroir.Float32},
		{outputType: wazeroir.SignedInt64, inputType: wazeroir.Float64},
		{outputType: wazeroir.SignedUint32, inputType: wazeroir.Float32},
		{outputType: wazeroir.SignedUint32, inputType: wazeroir.Float64},
		{outputType: wazeroir.SignedUint64, inputType: wazeroir.Float32},
		{outputType: wazeroir.SignedUint64, inputType: wazeroir.Float64},
		{outputType: wazeroir.SignedInt32, inputType: wazeroir.Float32, nonTrapping: true},
		{outputType: wazeroir.SignedInt32, inputType: wazeroir.Float64, nonTrapping: true},
		{outputType: wazeroir.SignedInt64, inputType: wazeroir.Float32, nonTrapping: true},
		{outputType: wazeroir.SignedInt64, inputType: wazeroir.Float64, nonTrapping: true},
		{outputType: wazeroir.SignedUint32, inputType: wazeroir.Float32, nonTrapping: true},
		{outputType: wazeroir.SignedUint32, inputType: wazeroir.Float64, nonTrapping: true},
		{outputType: wazeroir.SignedUint64, inputType: wazeroir.Float32, nonTrapping: true},
		{outputType: wazeroir.SignedUint64, inputType: wazeroir.Float64, nonTrapping: true},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(fmt.Sprintf("%s from %s (non-trapping=%v)", tc.outputType, tc.inputType, tc.nonTrapping), func(t *testing.T) {
			for _, v := range []float64{
				1.0,
			} {
				v := v
				if v == math.MaxInt32 {
					// Note that math.MaxInt32 is rounded up to math.MaxInt32+1 in 32-bit float representation.
					require.Equal(t, float32(2147483648.0) /* = math.MaxInt32+1 */, float32(v))
				} else if v == math.MaxUint32 {
					// Note that math.MaxUint32 is rounded up to math.MaxUint32+1 in 32-bit float representation.
					require.Equal(t, float32(4294967296 /* = math.MaxUint32+1 */), float32(v))
				} else if v == math.MaxInt64 {
					// Note that math.MaxInt64 is rounded up to math.MaxInt64+1 in 32/64-bit float representation.
					require.Equal(t, float32(9223372036854775808.0) /* = math.MaxInt64+1 */, float32(v))
					require.Equal(t, float64(9223372036854775808.0) /* = math.MaxInt64+1 */, float64(v))
				} else if v == math.MaxUint64 {
					// Note that math.MaxUint64 is rounded up to math.MaxUint64+1 in 32/64-bit float representation.
					require.Equal(t, float32(18446744073709551616.0) /* = math.MaxInt64+1 */, float32(v))
					require.Equal(t, float64(18446744073709551616.0) /* = math.MaxInt64+1 */, float64(v))
				}

				t.Run(fmt.Sprintf("%v", v), func(t *testing.T) {
					env := newCompilerEnvironment()
					compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
					err := compiler.compilePreamble()
					require.NoError(t, err)

					// Setup the conversion target.
					if tc.inputType == wazeroir.Float32 {
						err = compiler.compileConstF32(operationPtr(wazeroir.NewOperationConstF32(float32(v))))
					} else {
						err = compiler.compileConstF64(operationPtr(wazeroir.NewOperationConstF64(v)))
					}
					require.NoError(t, err)

					err = compiler.compileITruncFromF(operationPtr(wazeroir.NewOperationITruncFromF(
						tc.inputType, tc.outputType, tc.nonTrapping,
					)))
					require.NoError(t, err)

					err = compiler.compileReturnFunction()
					require.NoError(t, err)

					code := asm.CodeSegment{}
					defer func() { require.NoError(t, code.Unmap()) }()

					// Generate and run the code under test.
					_, err = compiler.compile(code.NextCodeSection())
					require.NoError(t, err)
					env.exec(code.Bytes())

					// Check the result.
					expStatus := nativeCallStatusCodeReturned
					if math.IsNaN(v) {
						if tc.nonTrapping {
							v = 0
						} else {
							expStatus = nativeCallStatusCodeInvalidFloatToIntConversion
						}
					}
					if tc.inputType == wazeroir.Float32 && tc.outputType == wazeroir.SignedInt32 {
						f32 := float32(v)
						exp := int32(math.Trunc(float64(f32)))
						if f32 < math.MinInt32 || f32 >= math.MaxInt32 {
							if tc.nonTrapping {
								if f32 < 0 {
									exp = math.MinInt32
								} else {
									exp = math.MaxInt32
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsInt32())
						}
					} else if tc.inputType == wazeroir.Float32 && tc.outputType == wazeroir.SignedInt64 {
						f32 := float32(v)
						exp := int64(math.Trunc(float64(f32)))
						if f32 < math.MinInt64 || f32 >= math.MaxInt64 {
							if tc.nonTrapping {
								if f32 < 0 {
									exp = math.MinInt64
								} else {
									exp = math.MaxInt64
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsInt64())
						}
					} else if tc.inputType == wazeroir.Float64 && tc.outputType == wazeroir.SignedInt32 {
						if v < math.MinInt32 || v > math.MaxInt32 {
							if tc.nonTrapping {
								if v < 0 {
									v = math.MinInt32
								} else {
									v = math.MaxInt32
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, int32(math.Trunc(v)), env.stackTopAsInt32())
						}
					} else if tc.inputType == wazeroir.Float64 && tc.outputType == wazeroir.SignedInt64 {
						exp := int64(math.Trunc(v))
						if v < math.MinInt64 || v >= math.MaxInt64 {
							if tc.nonTrapping {
								if v < 0 {
									exp = math.MinInt64
								} else {
									exp = math.MaxInt64
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsInt64())
						}
					} else if tc.inputType == wazeroir.Float32 && tc.outputType == wazeroir.SignedUint32 {
						f32 := float32(v)
						exp := uint32(math.Trunc(float64(f32)))
						if f32 < 0 || f32 >= math.MaxUint32 {
							if tc.nonTrapping {
								if v < 0 {
									exp = 0
								} else {
									exp = math.MaxUint32
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsUint32())
						}
					} else if tc.inputType == wazeroir.Float64 && tc.outputType == wazeroir.SignedUint32 {
						exp := uint32(math.Trunc(v))
						if v < 0 || v > math.MaxUint32 {
							if tc.nonTrapping {
								if v < 0 {
									exp = 0
								} else {
									exp = math.MaxUint32
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsUint32())
						}
					} else if tc.inputType == wazeroir.Float32 && tc.outputType == wazeroir.SignedUint64 {
						f32 := float32(v)
						exp := uint64(math.Trunc(float64(f32)))
						if f32 < 0 || f32 >= math.MaxUint64 {
							if tc.nonTrapping {
								if v < 0 {
									exp = 0
								} else {
									exp = math.MaxUint64
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsUint64())
						}
					} else if tc.inputType == wazeroir.Float64 && tc.outputType == wazeroir.SignedUint64 {
						exp := uint64(math.Trunc(v))
						if v < 0 || v >= math.MaxUint64 {
							if tc.nonTrapping {
								if v < 0 {
									exp = 0
								} else {
									exp = math.MaxUint64
								}
							} else {
								expStatus = nativeCallStatusIntegerOverflow
							}
						}
						if expStatus == nativeCallStatusCodeReturned {
							require.Equal(t, exp, env.stackTopAsUint64())
						}
					}
					require.Equal(t, expStatus, env.compilerStatus())
				})
			}
		})
	}
}

func TestCompiler_compileFConvertFromI(t *testing.T) {
	tests := []struct {
		inputType  wazeroir.SignedInt
		outputType wazeroir.Float
	}{
		{inputType: wazeroir.SignedInt32, outputType: wazeroir.Float32},
		{inputType: wazeroir.SignedInt32, outputType: wazeroir.Float64},
		{inputType: wazeroir.SignedInt64, outputType: wazeroir.Float32},
		{inputType: wazeroir.SignedInt64, outputType: wazeroir.Float64},
		{inputType: wazeroir.SignedUint32, outputType: wazeroir.Float32},
		{inputType: wazeroir.SignedUint32, outputType: wazeroir.Float64},
		{inputType: wazeroir.SignedUint64, outputType: wazeroir.Float32},
		{inputType: wazeroir.SignedUint64, outputType: wazeroir.Float64},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(fmt.Sprintf("%s from %s", tc.outputType, tc.inputType), func(t *testing.T) {
			for _, v := range []uint64{
				0, 1, 12345, 1 << 31, 1 << 32, 1 << 54, 1 << 63,
				0xffff_ffff_ffff_ffff, 0xffff_ffff,
				0xffff_ffff_ffff_fffe, 0xffff_fffe,
				math.MaxUint32, math.MaxUint64, math.MaxInt32, math.MaxInt64,
			} {
				t.Run(fmt.Sprintf("%d", v), func(t *testing.T) {
					env := newCompilerEnvironment()
					compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
					err := compiler.compilePreamble()
					require.NoError(t, err)

					// Setup the conversion target.
					if tc.inputType == wazeroir.SignedInt32 || tc.inputType == wazeroir.SignedUint32 {
						err = compiler.compileConstI32(operationPtr(wazeroir.NewOperationConstI32(uint32(v))))
					} else {
						err = compiler.compileConstI64(operationPtr(wazeroir.NewOperationConstI64(uint64(v))))
					}
					require.NoError(t, err)

					err = compiler.compileFConvertFromI(operationPtr(wazeroir.NewOperationFConvertFromI(
						tc.inputType, tc.outputType,
					)))
					require.NoError(t, err)

					err = compiler.compileReturnFunction()
					require.NoError(t, err)

					code := asm.CodeSegment{}
					defer func() { require.NoError(t, code.Unmap()) }()

					// Generate and run the code under test.
					_, err = compiler.compile(code.NextCodeSection())
					require.NoError(t, err)
					env.exec(code.Bytes())

					// Check the result.
					require.Equal(t, uint64(1), env.stackPointer())
					actualBits := env.stackTopAsUint64()
					if tc.outputType == wazeroir.Float32 && tc.inputType == wazeroir.SignedInt32 {
						exp := float32(int32(v))
						actual := math.Float32frombits(uint32(actualBits))
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float32 && tc.inputType == wazeroir.SignedInt64 {
						exp := float32(int64(v))
						actual := math.Float32frombits(uint32(actualBits))
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float64 && tc.inputType == wazeroir.SignedInt32 {
						exp := float64(int32(v))
						actual := math.Float64frombits(actualBits)
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float64 && tc.inputType == wazeroir.SignedInt64 {
						exp := float64(int64(v))
						actual := math.Float64frombits(actualBits)
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float32 && tc.inputType == wazeroir.SignedUint32 {
						exp := float32(uint32(v))
						actual := math.Float32frombits(uint32(actualBits))
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float64 && tc.inputType == wazeroir.SignedUint32 {
						exp := float64(uint32(v))
						actual := math.Float64frombits(actualBits)
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float32 && tc.inputType == wazeroir.SignedUint64 {
						exp := float32(v)
						actual := math.Float32frombits(uint32(actualBits))
						require.Equal(t, exp, actual)
					} else if tc.outputType == wazeroir.Float64 && tc.inputType == wazeroir.SignedUint64 {
						exp := float64(v)
						actual := math.Float64frombits(actualBits)
						require.Equal(t, exp, actual)
					}
				})
			}
		})
	}
}

func TestCompiler_compileF64PromoteFromF32(t *testing.T) {
	for _, v := range []float32{
		0, 100, -100, 1, -1,
		100.01234124, -100.01234124, 200.12315,
		math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		float32(math.Inf(1)), float32(math.Inf(-1)), float32(math.NaN()),
	} {
		t.Run(fmt.Sprintf("%f", v), func(t *testing.T) {
			env := newCompilerEnvironment()
			compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
			err := compiler.compilePreamble()
			require.NoError(t, err)

			// Setup the promote target.
			err = compiler.compileConstF32(operationPtr(wazeroir.NewOperationConstF32(v)))
			require.NoError(t, err)

			err = compiler.compileF64PromoteFromF32()
			require.NoError(t, err)

			err = compiler.compileReturnFunction()
			require.NoError(t, err)

			code := asm.CodeSegment{}
			defer func() { require.NoError(t, code.Unmap()) }()

			// Generate and run the code under test.
			_, err = compiler.compile(code.NextCodeSection())
			require.NoError(t, err)
			env.exec(code.Bytes())

			// Check the result.
			require.Equal(t, uint64(1), env.stackPointer())
			if math.IsNaN(float64(v)) {
				require.True(t, math.IsNaN(env.stackTopAsFloat64()))
			} else {
				exp := float64(v)
				actual := env.stackTopAsFloat64()
				require.Equal(t, exp, actual)
			}
		})
	}
}

func TestCompiler_compileF32DemoteFromF64(t *testing.T) {
	for _, v := range []float64{
		0, 100, -100, 1, -1,
		100.01234124, -100.01234124, 200.12315,
		math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		6.8719476736e+10,  /* = 1 << 36 */
		1.37438953472e+11, /* = 1 << 37 */
		math.Inf(1), math.Inf(-1), math.NaN(),
	} {
		t.Run(fmt.Sprintf("%f", v), func(t *testing.T) {
			env := newCompilerEnvironment()
			compiler := env.requireNewCompiler(t, &wasm.FunctionType{}, newCompiler, nil)
			err := compiler.compilePreamble()
			require.NoError(t, err)

			// Setup the demote target.
			err = compiler.compileConstF64(operationPtr(wazeroir.NewOperationConstF64(v)))
			require.NoError(t, err)

			err = compiler.compileF32DemoteFromF64()
			require.NoError(t, err)

			err = compiler.compileReturnFunction()
			require.NoError(t, err)

			code := asm.CodeSegment{}
			defer func() { require.NoError(t, code.Unmap()) }()

			// Generate and run the code under test.
			_, err = compiler.compile(code.NextCodeSection())
			require.NoError(t, err)
			env.exec(code.Bytes())

			// Check the result.
			require.Equal(t, uint64(1), env.stackPointer())
			if math.IsNaN(v) {
				require.True(t, math.IsNaN(float64(env.stackTopAsFloat32())))
			} else {
				exp := float32(v)
				actual := env.stackTopAsFloat32()
				require.Equal(t, exp, actual)
			}
		})
	}
}
