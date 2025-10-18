package cpu

import (
	"strings"
	"testing"

	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

func TestAddi(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 10 // Set register x1 to 10

	instr := ITypeInstruction{
		rd:  2, // Destination register x2
		rs1: 1, // Source register x1
		imm: 5, // Immediate value 5
	}

	err := Addi(core, instr)
	if err != nil {
		t.Fatalf("Addi failed: %v", err)
	}

	expected := uint32(15) // 10 + 5
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestExecute_Addi(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 20 // Set register x1 to 20

	// Encode ADDI x2, x1, 10
	instruction := uint32(0b00000000101000001000000100010011)

	err := Execute(core, instruction)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	expected := uint32(30) // 20 + 10
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestExecute_UnsupportedInstruction(t *testing.T) {
	core := NewCore(&devices.Bus{})

	// Encode an unsupported instruction
	instruction := uint32(0xFFFFFFFF)

	err := Execute(core, instruction)
	if err == nil {
		t.Fatal("Expected error for unsupported instruction, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported instruction") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestJalr(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 0x1000 // Set register x1 to target address
	core.pc = 0x2000   // Set program counter

	instr := ITypeInstruction{
		rd:  2,    // Destination register x2
		rs1: 1,    // Source register x1
		imm: 0x10, // Immediate offset
	}

	err := Jarl(core, instr)
	if err != nil {
		t.Fatalf("Jarl failed: %v", err)
	}

	expectedPC := uint32(0x1010) // (0x1000 + 0x10) &^ 1
	if core.pc != expectedPC {
		t.Errorf("Expected PC to be %X, got %X", expectedPC, core.pc)
	}

	expectedReturnAddr := uint32(0x2004) // Original PC + 4
	if core.x[2] != expectedReturnAddr {
		t.Errorf("Expected x2 to be %X, got %X", expectedReturnAddr, core.x[2])
	}
}
