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

	// Should NOT have advanced PC by 4 in ecall because it should have trapped
	// But ecall() calls core.Trap which handles PC saving and jumping to vector.
	// In our core.Trap, it doesn't advance PC before saving it to mepc/sepc.
	if core.pc == 0x1004 {
		t.Errorf("PC should not have advanced for non-SBI ecall")
	}

	if core.csrs[csrMcause] != ExceptionEnvironmentCallFromUMode {
		t.Errorf("Expected cause %d, got %d", ExceptionEnvironmentCallFromUMode, core.csrs[csrMcause])
	}
}
