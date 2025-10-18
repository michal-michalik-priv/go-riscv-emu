package system

import (
	"log/slog"

	"github.com/Keisim/go-riscv-emu/pkg/cpu"
	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

const (
	// RAMOffset is the starting address of the RAM in the system's memory map.
	RAMOffset      = 0x80000000
	DummyTTYOffset = 0x10000000
)

// System represents the entire emulation system, including the CPU and memory.
type System struct {
	core *cpu.Core
	bus  devices.Bus
}

// NewSystem initializes and returns a new System with a CPU core and RAM device.
func NewSystem(dummy_tty bool) *System {
	bus := devices.Bus{}
	ramDevice := devices.RAMDevice{}
	ramDevice.Initialize(RAMOffset, 0x10000000) // 256 MB RAM
	bus.AddDevice(&ramDevice)

	if dummy_tty {
		dummyTTYDevice := devices.DummyTTYDevice{}
		dummyTTYDevice.Initialize(DummyTTYOffset, 0x1) // 1 byte of Dummy TTY
		bus.AddDevice(&dummyTTYDevice)
	}

	system := System{
		core: cpu.NewCore(&bus),
		bus:  bus,
	}

	return &system
}

// Core returns the CPU core of the system.
func (s *System) Core() *cpu.Core {
	return s.core
}

// Bus returns the device bus of the system.
func (s *System) Bus() *devices.Bus {
	return &s.bus
}

// Step executes a single instruction cycle of the CPU core.
func (s *System) Step() {
	err := cpu.Step(s.core)
	if err != nil {
		slog.Error("Failed to execute CPU step:", "error", err)
		panic(err)
	}
}
