package devices

import (
	"fmt"

	"github.com/Keisim/go-riscv-emu/pkg/memory"
)

// RAMDevice represents a block of RAM accessible via MMIO.
type RAMDevice struct {
	baseAddress uint32
	size        uint32
	memory      *memory.RandomAccessMemory
}

// Read reads a byte from the RAM device at the specified address.
// Returns an error if the address is out of bounds.
func (r *RAMDevice) Read(address uint32) (byte, error) {
	if address < r.baseAddress || address >= r.baseAddress+r.size {
		return 0, fmt.Errorf(
			"attempted to read from invalid MMIO RAM address %X", address)
	}
	return r.memory.Read(address - r.baseAddress)
}

// Write writes a byte to the RAM device at the specified address.
// Returns an error if the address is out of bounds.
func (r *RAMDevice) Write(address uint32, value byte) error {
	if address < r.baseAddress || address >= r.baseAddress+r.size {
		return fmt.Errorf(
			"attempted to write %X to invalid MMIO RAM address %X",
			value, address)
	}
	return r.memory.Write(address-r.baseAddress, value)
}

// BaseAddress returns the base address of the RAM device.
func (r *RAMDevice) BaseAddress() uint32 {
	return r.baseAddress
}

// Size returns the size of the RAM device in bytes.
func (r *RAMDevice) Size() uint32 {
	return r.size
}

// Initialize sets up the RAM device with the specified base address and
// size.
func (r *RAMDevice) Initialize(baseAddress, size uint32) {
	r.baseAddress = baseAddress
	r.size = size
	r.memory = memory.NewRAM(size)
}
