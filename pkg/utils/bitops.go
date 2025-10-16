package utils

// Bitmask returns a bitmask with the specified number of least
// significant bits set to 1.
func Bitmask(bits uint8) uint32 {
	if bits >= 32 {
		return 0xFFFFFFFF
	}
	return (1 << bits) - 1
}

// ShiftedBitmask returns a bitmask with the specified number of least
// significant bits set to 1, shifted left by the specified amount.
func ShiftedBitmask(bits uint8, shift uint8) uint32 {
	return Bitmask(bits) << shift
}

// BitsSlice extracts a slice of bits from the given value, starting
// from 'start' (inclusive) to 'end' (exclusive).
// Returns 0 if the specified range is invalid.
func BitsSlice(value uint32, start, end uint8) uint32 {
	if start > end || end > 32 {
		return 0
	}
	mask := Bitmask(end - start)
	return (value >> start) & mask
}

// SignExtend sign-extends the given value from the specified bit width
// to a 32-bit signed integer.
func SignExtend(value uint32, bits uint8) int32 {
	if bits >= 32 {
		return int32(value)
	}

	signBit := uint32(1 << (bits - 1))
	// Negative number
	if value&signBit != 0 {
		extended := value | ^Bitmask(bits)
		return int32(extended)
	}

	// Positive number
	return int32(value)
}
