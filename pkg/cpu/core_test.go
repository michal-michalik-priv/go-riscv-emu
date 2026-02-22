package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
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

func TestSetRegister(t *testing.T) {
	core := NewCore(&devices.Bus{})
	regIndex := 10 // A0
	regValue := uint32(0x12345678)
	core.SetRegister(regIndex, regValue)

	if core.x[regIndex] != regValue {
		t.Errorf("Expected register %d to be %X, got %X", regIndex, regValue, core.x[regIndex])
	}
}

func TestGetRegister(t *testing.T) {
	core := NewCore(nil)
	core.SetRegister(1, 100)
	if core.GetRegister(1) != 100 {
		t.Errorf("Expected register 1 to be 100, got %d", core.GetRegister(1))
	}
	if core.GetRegister(0) != 0 {
		t.Errorf("Expected register 0 to be 0, got %d", core.GetRegister(0))
	}
}

func TestBootloader(t *testing.T) {
	core := NewCore(nil)
	entryPoint := uint32(0x80000000)
	core.Bootloader(entryPoint)

	if core.mode != ModeSupervisor {
		t.Errorf("Expected mode to be Supervisor (%d), got %d", ModeSupervisor, core.mode)
	}

	if core.pc != entryPoint {
		t.Errorf("Expected PC to be 0x%X, got 0x%X", entryPoint, core.pc)
	}

	mstatus := core.csrs[csrMstatus]
	if (mstatus & mstatusMPP) != 0 {
		t.Errorf("Expected MPP to be cleared after MRET (User mode), got %d", (mstatus&mstatusMPP)>>11)
	}

	mideleg := core.csrs[csrMideleg]
	if mideleg != 0xFFFF {
		t.Errorf("Expected mideleg to be 0xFFFF, got 0x%X", mideleg)
	}

	medeleg := core.csrs[csrMedeleg]
	if medeleg != 0xFFFF {
		t.Errorf("Expected medeleg to be 0xFFFF, got 0x%X", medeleg)
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
