package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestSTIPPropagation(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// 1. Delegate Timer Interrupts to S-mode
	core.csrs[csrMideleg] |= (1 << 5) // Delegate STimer

	// 2. Set MTIP (Machine Timer Interrupt Pending)
	core.SetInterrupt(1<<7, true)

	// 3. Trigger CheckInterrupts to propagate MTIP to STIP
	core.CheckInterrupts()

	// 4. Verify STIP is set in mip
	if (core.csrs[csrMip] & (1 << 5)) == 0 {
		t.Error("STIP was not propagated from MTIP when STimer is delegated")
	}

	// 5. Clear MTIP
	core.SetInterrupt(1<<7, false)

	// 6. Trigger CheckInterrupts again
	core.CheckInterrupts()

	// 7. Verify STIP is cleared
	if (core.csrs[csrMip] & (1 << 5)) != 0 {
		t.Error("STIP was not cleared when MTIP was cleared")
	}
}

func TestSTIPNoPropagationWithoutDelegation(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// 1. Ensure STimer is NOT delegated
	core.csrs[csrMideleg] &^= (1 << 5)

	// 2. Set MTIP
	core.SetInterrupt(1<<7, true)

	// 3. Trigger CheckInterrupts
	core.CheckInterrupts()

	// 4. Verify STIP is NOT set
	if (core.csrs[csrMip] & (1 << 5)) != 0 {
		t.Error("STIP was propagated from MTIP even though STimer is NOT delegated")
	}
}
