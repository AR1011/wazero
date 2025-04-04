package wasm

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/AR1011/wazero/api"
	"github.com/AR1011/wazero/internal/testing/require"
)

func TestMemoryPageConsts(t *testing.T) {
	require.Equal(t, MemoryPageSize, uint32(1)<<MemoryPageSizeInBits)
	require.Equal(t, MemoryPageSize, uint32(1<<16))
	require.Equal(t, MemoryLimitPages, uint32(1<<16))
}

func TestMemoryPagesToBytesNum(t *testing.T) {
	for _, numPage := range []uint32{0, 1, 5, 10} {
		require.Equal(t, uint64(numPage*MemoryPageSize), MemoryPagesToBytesNum(numPage))
	}
}

func TestMemoryBytesNumToPages(t *testing.T) {
	for _, numbytes := range []uint32{0, MemoryPageSize * 1, MemoryPageSize * 10} {
		require.Equal(t, numbytes/MemoryPageSize, memoryBytesNumToPages(uint64(numbytes)))
	}
}

func TestMemoryInstance_Grow_Size(t *testing.T) {
	tests := []struct {
		name         string
		capEqualsMax bool
	}{
		{name: ""},
		{name: "capEqualsMax", capEqualsMax: true},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			max := uint32(10)
			maxBytes := MemoryPagesToBytesNum(max)
			var m *MemoryInstance
			if tc.capEqualsMax {
				m = &MemoryInstance{Cap: max, Max: max, Buffer: make([]byte, 0, maxBytes)}
			} else {
				m = &MemoryInstance{Max: max, Buffer: make([]byte, 0)}
			}

			res, ok := m.Grow(5)
			require.True(t, ok)
			require.Equal(t, uint32(0), res)
			require.Equal(t, uint32(5), m.PageSize())

			// Zero page grow is well-defined, should return the current page correctly.
			res, ok = m.Grow(0)
			require.True(t, ok)
			require.Equal(t, uint32(5), res)
			require.Equal(t, uint32(5), m.PageSize())

			res, ok = m.Grow(4)
			require.True(t, ok)
			require.Equal(t, uint32(5), res)
			require.Equal(t, uint32(9), m.PageSize())

			// At this point, the page size equal 9,
			// so trying to grow two pages should result in failure.
			_, ok = m.Grow(2)
			require.False(t, ok)
			require.Equal(t, uint32(9), m.PageSize())

			// But growing one page is still permitted.
			res, ok = m.Grow(1)
			require.True(t, ok)
			require.Equal(t, uint32(9), res)

			// Ensure that the current page size equals the max.
			require.Equal(t, max, m.PageSize())

			if tc.capEqualsMax { // Ensure the capacity isn't more than max.
				require.Equal(t, maxBytes, uint64(cap(m.Buffer)))
			} else { // Slice doubles, so it should have a higher capacity than max.
				require.True(t, maxBytes < uint64(cap(m.Buffer)))
			}
		})
	}
}

func TestMemoryInstance_ReadByte(t *testing.T) {
	mem := &MemoryInstance{Buffer: []byte{0, 0, 0, 0, 0, 0, 0, 16}, Min: 1}
	v, ok := mem.ReadByte(7)
	require.True(t, ok)
	require.Equal(t, byte(16), v)

	_, ok = mem.ReadByte(8)
	require.False(t, ok)

	_, ok = mem.ReadByte(9)
	require.False(t, ok)
}

func TestPagesToUnitOfBytes(t *testing.T) {
	tests := []struct {
		name     string
		pages    uint32
		expected string
	}{
		{
			name:     "zero",
			pages:    0,
			expected: "0 Ki",
		},
		{
			name:     "one",
			pages:    1,
			expected: "64 Ki",
		},
		{
			name:     "megs",
			pages:    100,
			expected: "6 Mi",
		},
		{
			name:     "max memory",
			pages:    MemoryLimitPages,
			expected: "4 Gi",
		},
		{
			name:     "max uint32",
			pages:    math.MaxUint32,
			expected: "3 Ti",
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, PagesToUnitOfBytes(tc.pages))
		})
	}
}

func TestMemoryInstance_HasSize(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, MemoryPageSize)}

	tests := []struct {
		name        string
		offset      uint32
		sizeInBytes uint64
		expected    bool
	}{
		{
			name:        "simple valid arguments",
			offset:      0, // arbitrary valid offset
			sizeInBytes: 8, // arbitrary valid size
			expected:    true,
		},
		{
			name:        "maximum valid sizeInBytes",
			offset:      memory.Size() - 8,
			sizeInBytes: 8,
			expected:    true,
		},
		{
			name:        "sizeInBytes exceeds the valid size by 1",
			offset:      100, // arbitrary valid offset
			sizeInBytes: uint64(memory.Size() - 99),
			expected:    false,
		},
		{
			name:        "offset exceeds the memory size",
			offset:      memory.Size(),
			sizeInBytes: 1, // arbitrary size
			expected:    false,
		},
		{
			name:        "offset + sizeInBytes overflows in uint32",
			offset:      math.MaxUint32 - 1, // invalid too large offset
			sizeInBytes: 4,                  // if there's overflow, offset + sizeInBytes is 3, and it may pass the check
			expected:    false,
		},
		{
			name:        "address.wast:200",
			offset:      4294967295,
			sizeInBytes: 1,
			expected:    false,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, memory.hasSize(tc.offset, tc.sizeInBytes))
		})
	}
}

func TestMemoryInstance_ReadUint16Le(t *testing.T) {
	tests := []struct {
		name       string
		memory     []byte
		offset     uint32
		expected   uint16
		expectedOk bool
	}{
		{
			name:       "valid offset with an endian-insensitive v",
			memory:     []byte{0xff, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint16,
			expectedOk: true,
		},
		{
			name:       "valid offset with an endian-sensitive v",
			memory:     []byte{0xfe, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint16 - 1,
			expectedOk: true,
		},
		{
			name:       "maximum boundary valid offset",
			offset:     1,
			memory:     []byte{0x00, 0x1, 0x00},
			expected:   1, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:   "offset exceeds the maximum valid offset by 1",
			memory: []byte{0xff, 0xff},
			offset: 1,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			memory := &MemoryInstance{Buffer: tc.memory}

			v, ok := memory.ReadUint16Le(tc.offset)
			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestMemoryInstance_ReadUint32Le(t *testing.T) {
	tests := []struct {
		name       string
		memory     []byte
		offset     uint32
		expected   uint32
		expectedOk bool
	}{
		{
			name:       "valid offset with an endian-insensitive v",
			memory:     []byte{0xff, 0xff, 0xff, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint32,
			expectedOk: true,
		},
		{
			name:       "valid offset with an endian-sensitive v",
			memory:     []byte{0xfe, 0xff, 0xff, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint32 - 1,
			expectedOk: true,
		},
		{
			name:       "maximum boundary valid offset",
			offset:     1,
			memory:     []byte{0x00, 0x1, 0x00, 0x00, 0x00},
			expected:   1, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:   "offset exceeds the maximum valid offset by 1",
			memory: []byte{0xff, 0xff, 0xff, 0xff},
			offset: 1,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			memory := &MemoryInstance{Buffer: tc.memory}

			v, ok := memory.ReadUint32Le(tc.offset)
			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestMemoryInstance_ReadUint64Le(t *testing.T) {
	tests := []struct {
		name       string
		memory     []byte
		offset     uint32
		expected   uint64
		expectedOk bool
	}{
		{
			name:       "valid offset with an endian-insensitive v",
			memory:     []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint64,
			expectedOk: true,
		},
		{
			name:       "valid offset with an endian-sensitive v",
			memory:     []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.MaxUint64 - 1,
			expectedOk: true,
		},
		{
			name:       "maximum boundary valid offset",
			offset:     1,
			memory:     []byte{0x00, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:   1, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:   "offset exceeds the maximum valid offset by 1",
			memory: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			offset: 1,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			memory := &MemoryInstance{Buffer: tc.memory}

			v, ok := memory.ReadUint64Le(tc.offset)
			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestMemoryInstance_ReadFloat32Le(t *testing.T) {
	tests := []struct {
		name       string
		memory     []byte
		offset     uint32
		expected   float32
		expectedOk bool
	}{
		{
			name:       "valid offset with an endian-insensitive v",
			memory:     []byte{0xff, 0x00, 0x00, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.Float32frombits(uint32(0xff0000ff)),
			expectedOk: true,
		},
		{
			name:       "valid offset with an endian-sensitive v",
			memory:     []byte{0xfe, 0x00, 0x00, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.Float32frombits(uint32(0xff0000fe)),
			expectedOk: true,
		},
		{
			name:       "maximum boundary valid offset",
			offset:     1,
			memory:     []byte{0x00, 0xcd, 0xcc, 0xcc, 0x3d},
			expected:   0.1, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:   "offset exceeds the maximum valid offset by 1",
			memory: []byte{0xff, 0xff, 0xff, 0xff},
			offset: 1,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			memory := &MemoryInstance{Buffer: tc.memory}

			v, ok := memory.ReadFloat32Le(tc.offset)
			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestMemoryInstance_ReadFloat64Le(t *testing.T) {
	tests := []struct {
		name       string
		memory     []byte
		offset     uint32
		expected   float64
		expectedOk bool
	}{
		{
			name:       "valid offset with an endian-insensitive v",
			memory:     []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			offset:     0, // arbitrary valid offset.
			expected:   math.Float64frombits(uint64(0xff000000000000ff)),
			expectedOk: true,
		},
		{
			name:       "valid offset with an endian-sensitive v",
			memory:     []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
			offset:     0,               // arbitrary valid offset.
			expected:   math.MaxFloat64, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:       "maximum boundary valid offset",
			offset:     1,
			memory:     []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
			expected:   math.MaxFloat64, // arbitrary valid v
			expectedOk: true,
		},
		{
			name:   "offset exceeds the maximum valid offset by 1",
			memory: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			offset: 1,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			memory := &MemoryInstance{Buffer: tc.memory}

			v, ok := memory.ReadFloat64Le(tc.offset)
			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestMemoryInstance_Read(t *testing.T) {
	mem := &MemoryInstance{Buffer: []byte{0, 0, 0, 0, 16, 0, 0, 0}, Min: 1}

	buf, ok := mem.Read(4, 4)
	require.True(t, ok)
	require.Equal(t, []byte{16, 0, 0, 0}, buf)

	// Test write-through
	buf[3] = 4
	require.Equal(t, []byte{16, 0, 0, 4}, buf)
	require.Equal(t, []byte{0, 0, 0, 0, 16, 0, 0, 4}, mem.Buffer)

	_, ok = mem.Read(5, 4)
	require.False(t, ok)

	_, ok = mem.Read(9, 4)
	require.False(t, ok)
}

func TestMemoryInstance_WriteUint16Le(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, 100)}

	tests := []struct {
		name          string
		offset        uint32
		v             uint16
		expectedOk    bool
		expectedBytes []byte
	}{
		{
			name:          "valid offset with an endian-insensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint16,
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0xff},
		},
		{
			name:          "valid offset with an endian-sensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint16 - 1,
			expectedOk:    true,
			expectedBytes: []byte{0xfe, 0xff},
		},
		{
			name:          "maximum boundary valid offset",
			offset:        memory.Size() - 2, // 2 is the size of uint16
			v:             1,                 // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0x1, 0x00},
		},
		{
			name:          "offset exceeds the maximum valid offset by 1",
			offset:        memory.Size() - 2 + 1, // 2 is the size of uint16
			v:             1,                     // arbitrary valid v
			expectedBytes: []byte{0xff, 0xff},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOk, memory.WriteUint16Le(tc.offset, tc.v))
			if tc.expectedOk {
				require.Equal(t, tc.expectedBytes, memory.Buffer[tc.offset:tc.offset+2]) // 2 is the size of uint16
			}
		})
	}
}

func TestMemoryInstance_WriteUint32Le(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, 100)}

	tests := []struct {
		name          string
		offset        uint32
		v             uint32
		expectedOk    bool
		expectedBytes []byte
	}{
		{
			name:          "valid offset with an endian-insensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint32,
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:          "valid offset with an endian-sensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint32 - 1,
			expectedOk:    true,
			expectedBytes: []byte{0xfe, 0xff, 0xff, 0xff},
		},
		{
			name:          "maximum boundary valid offset",
			offset:        memory.Size() - 4, // 4 is the size of uint32
			v:             1,                 // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0x1, 0x00, 0x00, 0x00},
		},
		{
			name:          "offset exceeds the maximum valid offset by 1",
			offset:        memory.Size() - 4 + 1, // 4 is the size of uint32
			v:             1,                     // arbitrary valid v
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOk, memory.WriteUint32Le(tc.offset, tc.v))
			if tc.expectedOk {
				require.Equal(t, tc.expectedBytes, memory.Buffer[tc.offset:tc.offset+4]) // 4 is the size of uint32
			}
		})
	}
}

func TestMemoryInstance_WriteUint64Le(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, 100)}
	tests := []struct {
		name          string
		offset        uint32
		v             uint64
		expectedOk    bool
		expectedBytes []byte
	}{
		{
			name:          "valid offset with an endian-insensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint64,
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:          "valid offset with an endian-sensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.MaxUint64 - 1,
			expectedOk:    true,
			expectedBytes: []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:          "maximum boundary valid offset",
			offset:        memory.Size() - 8, // 8 is the size of uint64
			v:             1,                 // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0x1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:       "offset exceeds the maximum valid offset by 1",
			offset:     memory.Size() - 8 + 1, // 8 is the size of uint64
			v:          1,                     // arbitrary valid v
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOk, memory.WriteUint64Le(tc.offset, tc.v))
			if tc.expectedOk {
				require.Equal(t, tc.expectedBytes, memory.Buffer[tc.offset:tc.offset+8]) // 8 is the size of uint64
			}
		})
	}
}

func TestMemoryInstance_WriteFloat32Le(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, 100)}

	tests := []struct {
		name          string
		offset        uint32
		v             float32
		expectedOk    bool
		expectedBytes []byte
	}{
		{
			name:          "valid offset with an endian-insensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.Float32frombits(uint32(0xff0000ff)),
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0x00, 0x00, 0xff},
		},
		{
			name:          "valid offset with an endian-sensitive v",
			offset:        0,                                        // arbitrary valid offset.
			v:             math.Float32frombits(uint32(0xff0000fe)), // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0xfe, 0x00, 0x00, 0xff},
		},
		{
			name:          "maximum boundary valid offset",
			offset:        memory.Size() - 4, // 4 is the size of float32
			v:             0.1,               // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0xcd, 0xcc, 0xcc, 0x3d},
		},
		{
			name:          "offset exceeds the maximum valid offset by 1",
			offset:        memory.Size() - 4 + 1, // 4 is the size of float32
			v:             math.MaxFloat32,       // arbitrary valid v
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOk, memory.WriteFloat32Le(tc.offset, tc.v))
			if tc.expectedOk {
				require.Equal(t, tc.expectedBytes, memory.Buffer[tc.offset:tc.offset+4]) // 4 is the size of float32
			}
		})
	}
}

func TestMemoryInstance_WriteFloat64Le(t *testing.T) {
	memory := &MemoryInstance{Buffer: make([]byte, 100)}
	tests := []struct {
		name          string
		offset        uint32
		v             float64
		expectedOk    bool
		expectedBytes []byte
	}{
		{
			name:          "valid offset with an endian-insensitive v",
			offset:        0, // arbitrary valid offset.
			v:             math.Float64frombits(uint64(0xff000000000000ff)),
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		},
		{
			name:          "valid offset with an endian-sensitive v",
			offset:        0,               // arbitrary valid offset.
			v:             math.MaxFloat64, // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
		},
		{
			name:          "maximum boundary valid offset",
			offset:        memory.Size() - 8, // 8 is the size of float64
			v:             math.MaxFloat64,   // arbitrary valid v
			expectedOk:    true,
			expectedBytes: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
		},
		{
			name:       "offset exceeds the maximum valid offset by 1",
			offset:     memory.Size() - 8 + 1, // 8 is the size of float64
			v:          math.MaxFloat64,       // arbitrary valid v
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOk, memory.WriteFloat64Le(tc.offset, tc.v))
			if tc.expectedOk {
				require.Equal(t, tc.expectedBytes, memory.Buffer[tc.offset:tc.offset+8]) // 8 is the size of float64
			}
		})
	}
}

func TestMemoryInstance_Write(t *testing.T) {
	mem := &MemoryInstance{Buffer: []byte{0, 0, 0, 0, 16, 0, 0, 0}, Min: 1}

	buf := []byte{16, 0, 0, 4}
	require.True(t, mem.Write(4, buf))
	require.Equal(t, []byte{0, 0, 0, 0, 16, 0, 0, 4}, mem.Buffer)

	// Test it isn't write-through
	buf[3] = 0
	require.Equal(t, []byte{16, 0, 0, 0}, buf)
	require.Equal(t, []byte{0, 0, 0, 0, 16, 0, 0, 4}, mem.Buffer)

	ok := mem.Write(5, buf)
	require.False(t, ok)

	ok = mem.Write(9, buf)
	require.False(t, ok)
}

func TestMemoryInstance_Write_overflow(t *testing.T) {
	mem := &MemoryInstance{Buffer: []byte{0, 0, 0, 0, 16, 0, 0, 0}, Min: 1}

	// Test overflow
	huge := uint64(math.MaxUint32 + 1 + 4)
	if huge != uint64(int(huge)) {
		t.Skip("Skipping on 32-bit")
	}

	buf := []byte{16, 0, 0, 4}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	header.Len = int(huge)
	header.Cap = int(huge)

	require.False(t, mem.Write(4, buf))
}

func TestMemoryInstance_WriteString(t *testing.T) {
	mem := &MemoryInstance{Buffer: []byte{0, 0, 0, 0, 16, 0, 0, 0}, Min: 1}

	s := "bear"
	require.True(t, mem.WriteString(4, s))
	require.Equal(t, []byte{0, 0, 0, 0, 'b', 'e', 'a', 'r'}, mem.Buffer)

	ok := mem.WriteString(5, s)
	require.False(t, ok)

	ok = mem.WriteString(9, s)
	require.False(t, ok)
}

func BenchmarkWriteString(b *testing.B) {
	tests := []string{
		"",
		"bear",
		"hello world",
		strings.Repeat("hello ", 10),
	}
	//nolint intentionally testing interface access
	var mem api.Memory
	mem = &MemoryInstance{Buffer: make([]byte, 1000), Min: 1}
	for _, tt := range tests {
		b.Run("", func(b *testing.B) {
			b.Run("Write", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if !mem.Write(0, []byte(tt)) {
						b.Fail()
					}
				}
			})
			b.Run("WriteString", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if !mem.WriteString(0, tt) {
						b.Fail()
					}
				}
			})
		})
	}
}
