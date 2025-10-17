package system

import (
	"github.com/Keisim/go-riscv-emu/pkg/cpu"
	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

const (
	RAMOffset = 0x80000000
)

type System struct {
	core *cpu.Core
	bus  devices.Bus
}

func NewSystem() *System {
	bus := devices.Bus{}
	ramDevice := devices.RAMDevice{}
	ramDevice.Initialize(RAMOffset, 0x10000000) // 256 MB RAM
	bus.AddDevice(&ramDevice)

	system := System{
		core: cpu.NewCore(),
		bus:  bus,
	}

	return &system
}

func (s *System) Core() *cpu.Core {
	return s.core
}

func (s *System) Bus() *devices.Bus {
	return &s.bus
}
