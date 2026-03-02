package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestCounterEn(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// Machine mode should always have access
	core.mode = ModeMachine
	if !core.CanAccessCSR(csrTime, false) {
		t.Errorf("Machine mode denied access to csrTime")
	}

	// Disable TM bit in mcounteren
	core.WriteCSR(csrMcounteren, 0)

	// Supervisor mode should be denied access if mcounteren bit is 0
	core.mode = ModeSupervisor
	if core.CanAccessCSR(csrTime, false) {
		t.Errorf("Supervisor mode allowed access to csrTime when mcounteren[1] is 0")
	}

	// Enable TM bit in mcounteren
	core.mode = ModeMachine
	core.WriteCSR(csrMcounteren, 1<<1)

	// Supervisor mode should now have access
	core.mode = ModeSupervisor
	if !core.CanAccessCSR(csrTime, false) {
		t.Errorf("Supervisor mode denied access to csrTime when mcounteren[1] is 1")
	}

	// User mode should be denied access if scounteren bit is 0
	core.mode = ModeUser
	core.csrs[csrScounteren] = 0
	if core.CanAccessCSR(csrTime, false) {
		t.Errorf("User mode allowed access to csrTime when scounteren[1] is 0")
	}

	// Enable TM bit in scounteren
	core.mode = ModeMachine
	core.WriteCSR(csrScounteren, 1<<1)

	// User mode should now have access
	core.mode = ModeUser
	if !core.CanAccessCSR(csrTime, false) {
		t.Errorf("User mode denied access to csrTime when mcounteren[1] and scounteren[1] are 1")
	}
}
