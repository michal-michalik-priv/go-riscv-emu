package system

import (
	"fmt"
	"log/slog"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/cpu"
	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

const (
	// RAMOffset is the starting address of the RAM in the system's memory map.
	RAMOffset      = 0x80000000
	DummyTTYOffset = 0x10000000
	TimerOffset    = 0x02000000
)

// System represents the entire emulation system, including the CPU and memory.
type System struct {
	core *cpu.Core
	bus  *devices.Bus
}

// NewSystem initializes and returns a new System with a CPU core and RAM device.
func NewSystem(dummy_tty bool) *System {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(RAMOffset, 0x8000000) // 128 MB RAM
	bus.AddDevice(ramDevice)

	if dummy_tty {
		dummyTTYDevice := &devices.DummyTTYDevice{}
		dummyTTYDevice.Initialize(DummyTTYOffset, 0x1) // 1 byte of Dummy TTY
		bus.AddDevice(dummyTTYDevice)
	}

	system := System{
		core: cpu.NewCore(bus),
		bus:  bus,
	}

	timerDevice := &devices.TimerDevice{}
	timerDevice.Initialize(TimerOffset, 0x10000)
	timerDevice.SetInterruptHandler(system.core.SetInterrupt)
	timerDevice.Start()
	bus.AddDevice(timerDevice)

	return &system
}

// Core returns the CPU core of the system.
func (s *System) Core() *cpu.Core {
	return s.core
}

// Bootloader sets up the CPU state to boot into Supervisor mode.
func (s *System) Bootloader(entryPoint uint32) {
	s.core.Bootloader(entryPoint)
}

// Bus returns the device bus of the system.
func (s *System) Bus() *devices.Bus {
	return s.bus
}

// Step executes a single instruction cycle of the CPU core.
func (s *System) Step() {
	err := cpu.Step(s.core)
	if err != nil {
		slog.Error("Failed to execute CPU step:", "error", err)
		panic(err)
	}
}

// DumpState returns a string representation of the current system state.
func (s *System) DumpState() string {
	state := s.core.DumpState()

	// Add timer information if available
	for _, device := range s.bus.GetDevices() {
		if timer, ok := device.(*devices.TimerDevice); ok {
			state += "Timer Registers:\n"
			state += fmt.Sprintf(" mtime:    0x%016X\n", timer.GetMtime())
			state += fmt.Sprintf(" mtimecmp: 0x%016X\n", timer.GetMtimecmp())
			break
		}
	}

	return state
}
