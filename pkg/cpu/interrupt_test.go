package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestInterruptDelegation(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// Setup: Supervisor mode, S-mode timer interrupt delegated
	core.mode = ModeSupervisor
	core.csrs[csrMideleg] = (1 << 5)    // Delegate STIP
	core.csrs[csrMie] = (1 << 5)        // Enable STIP in mie
	core.csrs[csrMstatus] |= mstatusSIE // Enable S-mode interrupts
	core.csrs[csrStvec] = 0x80000100    // S-mode trap vector
	core.pc = 0x80000000

	// Trigger MTIP (which propagates to STIP when delegated)
	core.SetInterrupt(1<<7, true)

	// Execute one step. It should take the trap.
	// Since there's no instruction at 0x80000000, we'll just check if PC changed to stvec
	// We need something in memory to fetch, or Step will fail.
	// But CheckInterrupts is called BEFORE Fetch in Step.

	// Mock some memory for Fetch even if we hope it traps before
	ram := &devices.RAMDevice{}
	ram.Initialize(0x80000000, 0x1000)
	bus.AddDevice(ram)

	// Put a NOP instruction (0x00000013) at 0x80000000
	ram.Write(0x80000000, 0x13)
	ram.Write(0x80000001, 0x00)
	ram.Write(0x80000002, 0x00)
	ram.Write(0x80000003, 0x00)

	// Put a NOP at S-mode trap vector 0x80000100
	ram.Write(0x80000100, 0x13)
	ram.Write(0x80000101, 0x00)
	ram.Write(0x80000102, 0x00)
	ram.Write(0x80000103, 0x00)

	err := Step(core)
	if err != nil {
		t.Logf("Step returned error (expected if fetch failed): %v", err)
	}

	// After Step, PC should be at vector + 4 because it executed the NOP at vector
	if core.pc != 0x80000104 {
		t.Errorf("Expected PC to be 0x80000104 (after S-mode trap vector NOP), got 0x%08X", core.pc)
	}
	if core.mode != ModeSupervisor {
		t.Errorf("Expected mode to be Supervisor, got %d", core.mode)
	}
	if core.csrs[csrScause] != InterruptSTimer {
		t.Errorf("Expected scause to be %X, got %X", InterruptSTimer, core.csrs[csrScause])
	}
}

func TestMachineInterruptNoDelegation(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// Setup: Supervisor mode, STIP NOT delegated
	core.mode = ModeSupervisor
	core.csrs[csrMideleg] = 0    // NOT delegated
	core.csrs[csrMie] = (1 << 5) // Enable STIP in mie
	// M-mode interrupts are enabled because current mode < Machine
	core.csrs[csrMtvec] = 0x80000200 // M-mode trap vector
	core.pc = 0x80000000

	// Trigger STIP
	core.SetInterrupt(1<<5, true)

	ram := &devices.RAMDevice{}
	ram.Initialize(0x80000000, 0x1000)
	bus.AddDevice(ram)

	// Put a NOP instruction (0x00000013) at 0x80000000
	ram.Write(0x80000000, 0x13)
	ram.Write(0x80000001, 0x00)
	ram.Write(0x80000002, 0x00)
	ram.Write(0x80000003, 0x00)

	// Put a NOP at M-mode trap vector 0x80000200
	ram.Write(0x80000200, 0x13)
	ram.Write(0x80000201, 0x00)
	ram.Write(0x80000202, 0x00)
	ram.Write(0x80000203, 0x00)

	Step(core)

	// After Step, PC should be at vector + 4 because it executed the NOP at vector
	if core.pc != 0x80000204 {
		t.Errorf("Expected PC to be 0x80000204 (after M-mode trap vector NOP), got 0x%08X", core.pc)
	}
	if core.mode != ModeMachine {
		t.Errorf("Expected mode to be Machine, got %d", core.mode)
	}
}

func TestSieSipShadowing(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	core.csrs[csrMideleg] = (1 << 5) | (1 << 1) // Delegate STIP and SSIP

	// Write to sie
	err := core.WriteCSR(csrSie, (1 << 5))
	if err != nil {
		t.Fatalf("WriteCSR sie failed: %v", err)
	}

	// Check mie
	if (core.csrs[csrMie] & (1 << 5)) == 0 {
		t.Error("Expected STIP to be set in mie after writing to sie")
	}

	// Write to sip (SSIP)
	err = core.WriteCSR(csrSip, (1 << 1))
	if err != nil {
		t.Fatalf("WriteCSR sip failed: %v", err)
	}

	// Check mip
	if (core.csrs[csrMip] & (1 << 1)) == 0 {
		t.Error("Expected SSIP to be set in mip after writing to sip")
	}

	// Read from sie
	val, err := core.ReadCSR(csrSie)
	if err != nil {
		t.Fatalf("ReadCSR sie failed: %v", err)
	}
	if val != (1 << 5) {
		t.Errorf("Expected sie to be %x, got %x", (1 << 5), val)
	}

	// Read from sip
	val, err = core.ReadCSR(csrSip)
	if err != nil {
		t.Fatalf("ReadCSR sip failed: %v", err)
	}
	if val != (1 << 1) {
		t.Errorf("Expected sip to be %x, got %x", (1 << 1), val)
	}
}
