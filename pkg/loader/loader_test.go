package loader

import (
	"testing"

	"github.com/Keisim/go-riscv-emu/pkg/system"
)

func TestLoadELFToSystem(t *testing.T) {
	sys := system.NewSystem(false)

	err := LoadELFToSystem("../../misc/c/empty_main.o", sys)
	if err != nil {
		t.Fatalf("Failed to load ELF: %v", err)
	}

	device := sys.Bus().FindDevice(0x80000000)
	if device == nil {
		t.Fatalf("No device found at RAM base address after loading ELF")
	}

	byte1, _ := device.Read(0x80000000)
	byte2, _ := device.Read(0x80000001)
	byte3, _ := device.Read(0x80000002)
	byte4, _ := device.Read(0x80000003)

	first_op := uint32(byte1) | (uint32(byte2) << 8) |
		(uint32(byte3) << 16) | (uint32(byte4) << 24)

	expected_first_op := uint32(0x00000513) // ADDI x10, x0, 0
	if first_op != expected_first_op {
		t.Errorf("Expected first instruction %X, got %X",
			expected_first_op, first_op)
	}

	byte1, _ = device.Read(0x80000004)
	byte2, _ = device.Read(0x80000005)
	byte3, _ = device.Read(0x80000006)
	byte4, _ = device.Read(0x80000007)

	second_op := uint32(byte1) | (uint32(byte2) << 8) |
		(uint32(byte3) << 16) | (uint32(byte4) << 24)

	expected_second_op := uint32(0x00008067) // JALR x0, 0(x0)
	if second_op != expected_second_op {
		t.Errorf("Expected second instruction %X, got %X",
			expected_second_op, second_op)
	}

	expected_pc := uint32(0x80000000)
	if sys.Core().GetPc() != expected_pc {
		t.Errorf("Expected PC %X, got %X", expected_pc, sys.Core().GetPc())
	}
}
