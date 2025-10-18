package cpu

import (
	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

// Core represents the CPU core with its registers and program counter.
type Core struct {
	pc  uint32
	x   [32]uint32
	bus *devices.Bus
}

// NewCore creates and initializes a new CPU core with the given bus.
func NewCore(bus *devices.Bus) *Core {
	return &Core{
		pc:  0,
		bus: bus,
		x:   [32]uint32{},
	}
}

// SetPc sets the program counter to the specified value.
func (c *Core) SetPc(value uint32) {
	c.pc = value
}

// GetPc returns the current value of the program counter.
func (c *Core) GetPc() uint32 {
	return c.pc
}

// Fetch retrieves the next instruction from memory at the current PC.
func (c *Core) Fetch() uint32 {
	device := c.bus.FindDevice(c.pc)
	if device == nil {
		panic("No device found at PC address")
	}

	byte1, _ := device.Read(c.pc)
	byte2, _ := device.Read(c.pc + 1)
	byte3, _ := device.Read(c.pc + 2)
	byte4, _ := device.Read(c.pc + 3)

	instruction := uint32(byte1) | (uint32(byte2) << 8) |
		(uint32(byte3) << 16) | (uint32(byte4) << 24)

	return instruction
}
