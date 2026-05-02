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

	// User-mode ECALL is not SBI; it should trap to M-mode by default.
	if core.pc != 0x0 {
		t.Errorf("Expected PC to be 0x0 (mtvec), got %X", core.pc)
	}

	if core.mode != ModeMachine {
		t.Errorf("Expected mode to be Machine after trap, got %d", core.mode)
	}

	if core.csrs[csrMepc] != 0x1000 {
		t.Errorf("Expected mepc to be 0x1000, got %X", core.csrs[csrMepc])
	}

	if core.csrs[csrMcause] != ExceptionEnvironmentCallFromUMode {
		t.Errorf("Expected mcause to be %d, got %d", ExceptionEnvironmentCallFromUMode, core.csrs[csrMcause])
	}
}

func TestSBISetTimer(t *testing.T) {
	bus := &devices.Bus{}
	timer := &devices.TimerDevice{}
	timer.Initialize(0x02000000, 0x10000)
	bus.AddDevice(timer)

	core := NewCore(bus)
	core.mode = ModeSupervisor
	core.SetRegister(17, 0)          // a7 = sbiSetTimer
	core.SetRegister(10, 0x12345678) // a0 = low
	core.SetRegister(11, 0x87654321) // a1 = high
	core.pc = 0x1000

	// Set STIP first to see if it gets cleared
	core.SetInterrupt(1<<5, true)

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

	// Check if STIP is cleared
	if (core.csrs[csrMip] & (1 << 5)) != 0 {
		t.Errorf("Expected STIP to be cleared")
	}

	// Check if mtimecmp is set correctly
	var val uint64
	for i := uint32(0); i < 8; i++ {
		b, _ := bus.Read(0x02004000 + i)
		val |= uint64(b) << (i * 8)
	}
	expected := uint64(0x8765432112345678)
	if val != expected {
		t.Errorf("Expected mtimecmp to be %X, got %X", expected, val)
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
