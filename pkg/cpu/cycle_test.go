package cpu

import (
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

func TestCycleCSR(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// In Machine mode, we should have access to mcycle and cycle
	val, err := core.ReadCSR(csrMcycle)
	if err != nil {
		t.Fatalf("Failed to read mcycle: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected initial mcycle 0, got %d", val)
	}

	// Increment counters
	core.IncrementCounters(true)

	val, err = core.ReadCSR(csrMcycle)
	if err != nil {
		t.Fatalf("Failed to read mcycle: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected mcycle 1, got %d", val)
	}

	// Read via cycle (shadow of mcycle)
	val, err = core.ReadCSR(csrCycle)
	if err != nil {
		t.Fatalf("Failed to read cycle: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected cycle 1, got %d", val)
	}

	// Test 64-bit increment (mcycleH/cycleH)
	core.csrs[csrMcycle] = 0xFFFFFFFF
	core.IncrementCounters(true)

	val, err = core.ReadCSR(csrMcycle)
	if err != nil {
		t.Fatalf("Failed to read mcycle: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected mcycle 0 after overflow, got %d", val)
	}

	val, err = core.ReadCSR(csrMcycleH)
	if err != nil {
		t.Fatalf("Failed to read mcycleh: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected mcycleh 1, got %d", val)
	}

	val, err = core.ReadCSR(csrCycleH)
	if err != nil {
		t.Fatalf("Failed to read cycleh: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected cycleh 1, got %d", val)
	}
}

func TestCycleAccessControl(t *testing.T) {
	bus := &devices.Bus{}
	core := NewCore(bus)

	// Initial state: Machine mode, access allowed
	if !core.CanAccessCSR(csrCycle, false) {
		t.Error("Machine mode should access cycle")
	}

	// Switch to User mode
	core.mode = ModeUser

	// Default: mcounteren/scounteren are 0, so access should be denied
	if core.CanAccessCSR(csrCycle, false) {
		t.Error("User mode should NOT access cycle by default")
	}

	// Enable in mcounteren
	core.csrs[csrMcounteren] |= 1 // Bit 0 is CY
	if core.CanAccessCSR(csrCycle, false) {
		t.Error("User mode should NOT access cycle if only mcounteren is set")
	}

	// Enable in scounteren
	core.csrs[csrScounteren] |= 1 // Bit 0 is CY
	if !core.CanAccessCSR(csrCycle, false) {
		t.Error("User mode should access cycle if both mcounteren and scounteren are set")
	}

	// Read should work now
	core.csrs[csrMcycle] = 42
	val, err := core.ReadCSR(csrCycle)
	if err != nil {
		t.Fatalf("ReadCSR cycle failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Supervisor mode
	core.mode = ModeSupervisor
	core.csrs[csrScounteren] = 0 // Disable in scounteren
	if !core.CanAccessCSR(csrCycle, false) {
		t.Error("Supervisor mode should access cycle if mcounteren is set")
	}
}

func TestTimeCSR(t *testing.T) {
	bus := &devices.Bus{}
	timer := &devices.TimerDevice{}
	timer.Initialize(0x02000000, 0x10000)
	bus.AddDevice(timer)
	core := NewCore(bus)

	// In Machine mode, we should have access to time
	if !core.CanAccessCSR(csrTime, false) {
		t.Error("Machine mode should access time")
	}

	// Set mtime in timer device
	// mtime is at TimerOffset + 0xBFF8 = 0x0200BFF8
	timer.Write(0x0200BFF8, 0x11)
	timer.Write(0x0200BFF9, 0x22)
	timer.Write(0x0200BFFA, 0x33)
	timer.Write(0x0200BFFB, 0x44)
	timer.Write(0x0200BFFC, 0x55)
	timer.Write(0x0200BFFD, 0x66)
	timer.Write(0x0200BFFE, 0x77)
	timer.Write(0x0200BFFF, 0x88)

	val, err := core.ReadCSR(csrTime)
	if err != nil {
		t.Fatalf("Failed to read time: %v", err)
	}
	if val != 0x44332211 {
		t.Errorf("Expected time 0x44332211, got %X", val)
	}

	val, err = core.ReadCSR(csrTimeH)
	if err != nil {
		t.Fatalf("Failed to read timeh: %v", err)
	}
	if val != 0x88776655 {
		t.Errorf("Expected timeh 0x88776655, got %X", val)
	}

	// Test access control for time
	core.mode = ModeUser
	core.csrs[csrMcounteren] = 0
	core.csrs[csrScounteren] = 0
	if core.CanAccessCSR(csrTime, false) {
		t.Error("User mode should NOT access time by default")
	}

	core.csrs[csrMcounteren] |= 2 // Bit 1 is TM
	core.csrs[csrScounteren] |= 2 // Bit 1 is TM
	if !core.CanAccessCSR(csrTime, false) {
		t.Error("User mode should access time when enabled")
	}
}
