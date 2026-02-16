package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestCSRPrivilege(t *testing.T) {
	core := NewCore(&devices.Bus{})

	// Machine mode should access everything
	core.mode = ModeMachine
	if !core.CanAccessCSR(csrMstatus, true) {
		t.Error("Machine mode should access mstatus")
	}
	if !core.CanAccessCSR(csrSstatus, true) {
		t.Error("Machine mode should access sstatus")
	}

	// Supervisor mode should not access Machine CSRs
	core.mode = ModeSupervisor
	if core.CanAccessCSR(csrMstatus, true) {
		t.Error("Supervisor mode should not access mstatus")
	}
	if !core.CanAccessCSR(csrSstatus, true) {
		t.Error("Supervisor mode should access sstatus")
	}

	// User mode should not access Supervisor or Machine CSRs
	core.mode = ModeUser
	if core.CanAccessCSR(csrSstatus, true) {
		t.Error("User mode should not access sstatus")
	}
	if core.CanAccessCSR(csrMstatus, true) {
		t.Error("User mode should not access mstatus")
	}
}

func TestSstatusShadowing(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.mode = ModeMachine

	// Write to mstatus
	val := uint32(mstatusMIE | mstatusSIE)
	core.WriteCSR(csrMstatus, val)

	// Read from sstatus
	sval, _ := core.ReadCSR(csrSstatus)
	if (sval & mstatusSIE) == 0 {
		t.Error("sstatus should show SIE from mstatus")
	}
	if (sval & mstatusMIE) != 0 {
		t.Error("sstatus should not show MIE from mstatus")
	}

	// Write to sstatus
	core.WriteCSR(csrSstatus, mstatusSPIE)
	mval, _ := core.ReadCSR(csrMstatus)
	if (mval & mstatusSPIE) == 0 {
		t.Error("mstatus should show SPIE written via sstatus")
	}
}

func TestTrapDelegation(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.mode = ModeUser
	core.pc = 0x1000

	// Delegate ecall from U-mode (cause 8) to S-mode
	core.csrs[csrMedeleg] = (1 << 8)
	core.csrs[csrStvec] = 0x2000
	core.csrs[csrMtvec] = 0x3000

	core.Trap(ExceptionEnvironmentCallFromUMode, 0)

	if core.mode != ModeSupervisor {
		t.Errorf("Expected mode to be Supervisor, got %d", core.mode)
	}
	if core.pc != 0x2000 {
		t.Errorf("Expected PC to be 0x2000 (stvec), got %X", core.pc)
	}
	if core.csrs[csrSepc] != 0x1000 {
		t.Errorf("Expected sepc to be 0x1000, got %X", core.csrs[csrSepc])
	}
}
