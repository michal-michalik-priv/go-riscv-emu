package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestWFI(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)
	core.mode = ModeMachine
	core.pc = 0x1000

	// Test WFI
	// WFI instruction: 0x10500073
	err := execute(core, 0x10500073)
	if err != nil {
		t.Fatalf("WFI execution failed: %v", err)
	}

	if !core.wfi {
		t.Error("Expected core to be in WFI state")
	}
	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}
}

func TestWFIPrivilege(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// Test User Mode (should trap)
	core.mode = ModeUser
	core.pc = 0x1000
	err := execute(core, 0x10500073)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	// In User Mode, WFI should trap with Illegal Instruction
	if core.csrs[csrMcause] != ExceptionIllegalInstruction {
		t.Errorf("Expected Illegal Instruction trap, got cause %d", core.csrs[csrMcause])
	}

	// Test Supervisor Mode with TW=1 (should trap)
	core = NewCore(bus)
	core.mode = ModeSupervisor
	core.csrs[csrMstatus] |= mstatusTW
	core.pc = 0x1000
	err = execute(core, 0x10500073)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if core.csrs[csrMcause] != ExceptionIllegalInstruction {
		t.Errorf("Expected Illegal Instruction trap due to TW=1, got cause %d", core.csrs[csrMcause])
	}
}
