package memory

import (
	"fmt"
	"log/slog"
)

// RandomAccessMemory simulates a simple RAM module.
type RandomAccessMemory struct {
	size uint32
	data []byte
}

// NewRAM creates a new RAM instance with the specified size in bytes.
func NewRAM(size uint32) *RandomAccessMemory {
	slog.Debug(fmt.Sprintf("Initializing RAM of size %d bytes\n", size))
	return &RandomAccessMemory{
		size: size,
		data: make([]byte, size),
	}
}

// Read reads a byte from the specified address.
// Returns an error if the address is out of bounds.
func (ram *RandomAccessMemory) Read(address uint32) (byte, error) {
	if address >= ram.size {
		return 0, fmt.Errorf("read address out of bounds")
	}
	return ram.data[address], nil
}

// Write writes a byte to the specified address.
// Returns an error if the address is out of bounds.
func (ram *RandomAccessMemory) Write(address uint32, value byte) error {
	if address >= ram.size {
		return fmt.Errorf("write address out of bounds")
	}
	ram.data[address] = value
	return nil
}

// Size returns the size of the RAM in bytes.
func (ram *RandomAccessMemory) Size() uint32 {
	return ram.size
}
