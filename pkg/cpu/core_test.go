package cpu

import (
	"testing"

	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

func TestSetPc(t *testing.T) {
	core := NewCore(&devices.Bus{})
	newPc := uint32(0x1000)
	core.SetPc(newPc)

	if core.pc != newPc {
		t.Errorf("Expected PC to be %X, got %X", newPc, core.pc)
	}
}

func TestGetPc(t *testing.T) {
	core := NewCore(&devices.Bus{})
	expectedPc := uint32(0x2000)
	core.pc = expectedPc

	if core.GetPc() != expectedPc {
		t.Errorf("Expected PC to be %X, got %X", expectedPc, core.GetPc())
	}
}

type MockDevice struct {
	memory map[uint32]byte
}

func (d *MockDevice) Initialize(baseAddr uint32, size uint32) {
	// No initialization needed for mock
}

func (d *MockDevice) Read(addr uint32) (byte, error) {
	if val, ok := d.memory[addr]; ok {
		return val, nil
	}
	return 0, nil
}

func (d *MockDevice) Write(addr uint32, value byte) error {
	d.memory[addr] = value
	return nil
}

func (d *MockDevice) BaseAddress() uint32 {
	return 0x1000
}

func (d *MockDevice) Size() uint32 {
	return 0x100
}

func TestFetch(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &MockDevice{
		memory: map[uint32]byte{
			0x1000: 0x13,
			0x1001: 0x05,
			0x1002: 0x00,
			0x1003: 0x00,
		},
	}
	ramDevice.Initialize(0x1000, 0x100)
	bus.AddDevice(ramDevice)

	core := NewCore(bus)
	core.SetPc(0x1000)

	instruction := core.Fetch()
	expectedInstruction := uint32(0x00000513) // ADDI x10, x0, 0

	if instruction != expectedInstruction {
		t.Errorf("Expected instruction %X, got %X", expectedInstruction, instruction)
	}
}
