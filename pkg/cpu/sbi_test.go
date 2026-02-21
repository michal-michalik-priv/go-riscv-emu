package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestSBIPutchar(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.mode = ModeSupervisor
	core.SetRegister(17, 1)   // a7 = SBI_CONSOLE_PUTCHAR
	core.SetRegister(10, 'A') // a0 = 'A'
	core.pc = 0x1000

	err := ecall(core)
	if err != nil {
		t.Fatalf("ecall failed: %v", err)
	}

	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}

	a0 := core.GetRegister(10)
	if a0 != uint32(SBI_SUCCESS) {
		t.Errorf("Expected a0 to be SBI_SUCCESS (0), got %v", a0)
	}
}

func TestSBIWrongMode(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.mode = ModeUser      // Not Supervisor
	core.SetRegister(17, 1)   // a7 = SBI_CONSOLE_PUTCHAR
	core.SetRegister(10, 'A') // a0 = 'A'
	core.pc = 0x1000

	err := ecall(core)
	if err != nil {
		t.Fatalf("ecall failed: %v", err)
	}

	// Now it should have advanced PC because handleSBI handles it by returning error code
	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}

	a0 := core.GetRegister(10)
	errCode := int32(SBI_ERR_NOT_SUPPORTED)
	expectedA0 := uint32(errCode)
	if a0 != expectedA0 {
		t.Errorf("Expected a0 to be %X (SBI_ERR_NOT_SUPPORTED), got %X", expectedA0, a0)
	}
}

func TestSBIUnsupported(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.mode = ModeSupervisor
	core.SetRegister(17, 99) // a7 = Unsupported EID
	core.pc = 0x1000

	err := ecall(core)
	if err != nil {
		t.Fatalf("ecall failed: %v", err)
	}

	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}

	a0 := core.GetRegister(10)
	errCode := int32(SBI_ERR_NOT_SUPPORTED)
	expectedA0 := uint32(errCode)
	if a0 != expectedA0 {
		t.Errorf("Expected a0 to be %X (SBI_ERR_NOT_SUPPORTED), got %X", expectedA0, a0)
	}
}
