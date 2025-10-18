package cpu

import (
	"strings"
	"testing"

	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

func TestAddi(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 10 // Set register x1 to 10

	instr := iTypeInstruction{
		rd:  2, // Destination register x2
		rs1: 1, // Source register x1
		imm: 5, // Immediate value 5
	}

	err := addi(core, instr)
	if err != nil {
		t.Fatalf("addi failed: %v", err)
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

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
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

	err := execute(core, instruction)
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

	instr := iTypeInstruction{
		rd:  2,    // Destination register x2
		rs1: 1,    // Source register x1
		imm: 0x10, // Immediate offset
	}

	err := jarl(core, instr)
	if err != nil {
		t.Fatalf("jarl failed: %v", err)
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

func TestStep(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(0x1000, 0x100)
	bus.AddDevice(ramDevice)

	core := NewCore(bus)
	core.pc = 0x1000

	// Load an ADDI instruction into memory at address 0x1000
	// ADDI x2, x0, 42  ->  0x02A00093
	ramDevice.Write(0x1000, 0x93) // opcode and rd
	ramDevice.Write(0x1001, 0x00) // rs1 and funct3
	ramDevice.Write(0x1002, 0xA0) // imm[11:4]
	ramDevice.Write(0x1003, 0x02) // imm[3:0]

	err := Step(core)
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}
}

func TestLui(t *testing.T) {
	core := NewCore(&devices.Bus{})

	instr := uTypeInstruction{
		rd:  3,       // Destination register x3
		imm: 0x12345, // Immediate value
	}

	err := lui(core, instr)
	if err != nil {
		t.Fatalf("lui failed: %v", err)
	}

	expected := uint32(0x12345000) // 0x12345 << 12
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %X, got %X", expected, core.x[3])
	}
}

func TestExecute_Lui(t *testing.T) {
	core := NewCore(&devices.Bus{})

	// Encode LUI x4, 0x1ABCD
	instruction := uint32(0b00011010101111001101001000110111)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(0x1ABCD000) // 0x1ABCD << 12
	if core.x[4] != expected {
		t.Errorf("Expected x4 to be %X, got %X", expected, core.x[4])
	}
}

func TestSb(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(0x2000, 0x100)
	bus.AddDevice(ramDevice)

	core := NewCore(bus)
	core.x[1] = 0x2004 // Base address in rs1
	core.x[2] = 0xABCD // Value to store in rs2

	instr := sTypeInstruction{
		rs2: 2,    // Source register x2
		rs1: 1,    // Base register x1
		imm: 0x00, // Immediate offset
	}

	err := sb(core, instr)
	if err != nil {
		t.Fatalf("sb failed: %v", err)
	}

	// Verify that the byte at address 0x2004 is 0xCD (least significant byte of 0xABCD)
	value, err := ramDevice.Read(0x2004)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	expected := byte(0xCD)
	if value != expected {
		t.Errorf("Expected memory at 0x2004 to be %X, got %X", expected, value)
	}
}

func TestExecute_Sb(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(0x2000, 0x100)
	bus.AddDevice(ramDevice)

	core := NewCore(bus)
	core.x[1] = 0x2004 // Base address in rs1
	core.x[2] = 0xABCD // Value to store in rs2

	// Encode SB x2, 0(x1)
	// imm[11:5] = 0b0000000 (0)
	// rs2 = 0b00010 (2)
	// rs1 = 0b00001 (1)
	// funct3 = 0b000
	// imm[4:0] = 0b00000 (0)
	// opcode = 0b0100011 (SB)
	instruction := uint32(0b00000000001000001000000000100011)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify that the byte at address 0x2004 is 0xCD (least significant byte of 0xABCD)
	value, err := ramDevice.Read(0x2004)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	expected := byte(0xCD)
	if value != expected {
		t.Errorf("Expected memory at 0x2004 to be %X, got %X", expected, value)
	}
}

func TestJal(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x4000 // Set initial program counter

	instr := jTypeInstruction{
		rd:  5,     // Destination register x5
		imm: 0x200, // Immediate offset
	}

	err := jal(core, instr)
	if err != nil {
		t.Fatalf("jal failed: %v", err)
	}

	expectedPC := uint32(0x4200) // 0x4000 + 0x200
	if core.pc != expectedPC {
		t.Errorf("Expected PC to be %X, got %X", expectedPC, core.pc)
	}

	expectedReturnAddr := uint32(0x4004) // Original PC + 4
	if core.x[5] != expectedReturnAddr {
		t.Errorf("Expected x5 to be %X, got %X", expectedReturnAddr, core.x[5])
	}
}

func TestLb(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(0x4000, 0x100)
	bus.AddDevice(ramDevice)

	// Write a byte to memory at address 0x4005
	err := ramDevice.Write(0x4005, 0x7F)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	core := NewCore(bus)
	core.x[1] = 0x4000 // Base address in rs1

	instr := iTypeInstruction{
		rd:  2,    // Destination register x2
		rs1: 1,    // Source register x1
		imm: 0x05, // Immediate offset
	}

	err = lb(core, instr)
	if err != nil {
		t.Fatalf("lb failed: %v", err)
	}

	expected := uint32(0x7F)
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %X, got %X", expected, core.x[2])
	}
}

func TestLbu(t *testing.T) {
	bus := &devices.Bus{}
	ramDevice := &devices.RAMDevice{}
	ramDevice.Initialize(0x5000, 0x100)
	bus.AddDevice(ramDevice)

	// Write a byte to memory at address 0x5006
	err := ramDevice.Write(0x5006, 0xFF)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	core := NewCore(bus)
	core.x[1] = 0x5000 // Base address in rs1

	instr := iTypeInstruction{
		rd:  3,    // Destination register x3
		rs1: 1,    // Source register x1
		imm: 0x06, // Immediate offset
	}

	err = lbu(core, instr)
	if err != nil {
		t.Fatalf("lbu failed: %v", err)
	}

	expected := uint32(0xFF)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %X, got %X", expected, core.x[3])
	}
}

func TestBne(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000 // Set initial program counter
	core.x[1] = 10   // Set register x1
	core.x[2] = 20   // Set register x2

	instr := bTypeInstruction{
		rs1: 1,     // Source register x1
		rs2: 2,     // Source register x2
		imm: 0x100, // Immediate offset
	}

	err := bne(core, instr)
	if err != nil {
		t.Fatalf("bne failed: %v", err)
	}

	expectedPC := uint32(0x3100) // 0x3000 + 0x100
	if core.pc != expectedPC {
		t.Errorf("Expected PC to be %X, got %X", expectedPC, core.pc)
	}

	// Now test when registers are equal
	core.pc = 0x3000 // Reset program counter
	core.x[2] = 10   // Set register x2 equal to x1

	err = bne(core, instr)
	if err != nil {
		t.Fatalf("bne failed: %v", err)
	}

	expectedPC = uint32(0x3004) // PC should advance by 4
	if core.pc != expectedPC {
		t.Errorf("Expected PC to be %X, got %X", expectedPC, core.pc)
	}
}
