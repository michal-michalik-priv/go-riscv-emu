package cpu

import (
	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

type Core struct {
	pc  uint32
	x   [32]uint32
	bus *devices.Bus
}

func NewCore(bus *devices.Bus) *Core {
	return &Core{
		pc:  0,
		bus: bus,
		x:   [32]uint32{},
	}
}

func (c *Core) SetPc(value uint32) {
	c.pc = value
}

func (c *Core) GetPc() uint32 {
	return c.pc
}

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
