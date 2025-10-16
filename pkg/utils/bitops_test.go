package utils

import "testing"

func TestBitmask(t *testing.T) {
	tests := []struct {
		bits     uint8
		expected uint32
	}{
		{0, 0x0},
		{1, 0x1},
		{2, 0x3},
		{3, 0x7},
		{4, 0xF},
		{8, 0xFF},
		{16, 0xFFFF},
		{32, 0xFFFFFFFF},
		{33, 0xFFFFFFFF}, // Edge case: bits > 32
	}

	for _, test := range tests {
		result := Bitmask(test.bits)
		if result != test.expected {
			t.Errorf("Bitmask(%d) = %X; want %X", test.bits, result,
				test.expected)
		}
	}
}

func TestShiftedBitmask(t *testing.T) {
	tests := []struct {
		bits     uint8
		shift    uint8
		expected uint32
	}{
		{0, 0, 0x0},
		{1, 0, 0x1},
		{1, 1, 0x2},
		{2, 1, 0x6},
		{3, 2, 0x1C},
		{4, 4, 0xF0},
		{8, 8, 0xFF00},
		{16, 16, 0xFFFF0000},
	}

	for _, test := range tests {
		result := ShiftedBitmask(test.bits, test.shift)
		if result != test.expected {
			t.Errorf("ShiftedBitmask(%d, %d) = %X; want %X", test.bits,
				test.shift, result, test.expected)
		}
	}
}

func TestSignExtend(t *testing.T) {
	tests := []struct {
		value    uint32
		bits     uint8
		expected int32
	}{
		{0b00000000, 8, 0},
		{0b01111111, 8, 127},
		{0b10000000, 8, -128},
		{0b11111111, 8, -1},
		{0b00000000_00000000_00000000_01111111, 8, 127},
		{0b00000000_00000000_00000000_10000000, 8, -128},
		{0b00000000_00000000_00000000_11111111, 8, -1},
		{0b00000000_00000000_11111111_11111111, 16, -1},
		{0b00000000_00000000_10000000_00000000, 16, -32768},
		{0b00000000_00000000_01111111_11111111, 16, 32767},
		{0b00000000_00000001_00000000_00000000, 17, -65536},
		{0xFFFFFFFF, 32, -1},
		{0x7FFFFFFF, 32, 2147483647},
		{0x80000000, 32, -2147483648},
	}

	for _, test := range tests {
		result := SignExtend(test.value, test.bits)
		if result != test.expected {
			t.Errorf("SignExtend(%X, %d) = %d; want %d", test.value, test.bits,
				result, test.expected)
		}
	}
}

func TestBitsSlice(t *testing.T) {
	tests := []struct {
		value    uint32
		start    uint8
		end      uint8
		expected uint32
	}{
		{0b00000000, 0, 0, 0},
		{0b00000001, 0, 1, 1},
		{0b00000010, 1, 2, 1},
		{0b00000111, 0, 3, 7},
		{0b11111111, 4, 8, 15},
		{0b10101010_11110000, 4, 12, 0xAF},
		{0xFFFFFFFF, 16, 32, 0xFFFF},
		{0xFFFFFFFF, 0, 32, 0xFFFFFFFF},
		// Invalid ranges
		{0b00000000, 5, 3, 0},
		{0b00000000, 0, 33, 0},
	}

	for _, test := range tests {
		result := BitsSlice(test.value, test.start, test.end)
		if result != test.expected {
			t.Errorf("BitsSlice(%X, %d, %d) = %X; want %X", test.value,
				test.start, test.end, result, test.expected)
		}
	}
}
