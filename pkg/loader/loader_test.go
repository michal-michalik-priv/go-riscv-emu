package loader

import (
	"os"
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/system"
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

func TestLoadBinaryToSystem(t *testing.T) {
	sys := system.NewSystem(false)
	tempFile, err := os.CreateTemp("", "test_bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	data := []byte{0x13, 0x05, 0x00, 0x00} // ADDI x10, x0, 0
	if _, err := tempFile.Write(data); err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	loadAddr := uint32(0x80000000)
	err = LoadBinaryToSystem(tempFile.Name(), loadAddr, sys)
	if err != nil {
		t.Fatalf("Failed to load binary: %v", err)
	}

	for i, b := range data {
		readB, err := sys.Bus().Read(loadAddr + uint32(i))
		if err != nil {
			t.Errorf("Error reading at %X: %v", loadAddr+uint32(i), err)
		}
		if readB != b {
			t.Errorf("Expected %X at %X, got %X", b, loadAddr+uint32(i), readB)
		}
	}

	if sys.Core().GetPc() != loadAddr {
		t.Errorf("Expected PC %X, got %X", loadAddr, sys.Core().GetPc())
	}
}
