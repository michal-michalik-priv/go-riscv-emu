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

func TestParseIType(t *testing.T) {
	instruction := uint32(0x04800713) // Example instruction
	parsed := parseIType(instruction)

	if parsed.rd != 14 {
		t.Errorf("Expected rd to be 14, got %d", parsed.rd)
	}
	if parsed.rs1 != 0 {
		t.Errorf("Expected rs1 to be 0, got %d", parsed.rs1)
	}
	if parsed.imm != 72 {
		t.Errorf("Expected imm to be 72, got %d", parsed.imm)
	}
}

func TestParseBType(t *testing.T) {
	instruction := uint32(0xfed79ae3)
	parsed := parseBType(instruction)

	if parsed.rs1 != 15 {
		t.Errorf("Expected rs1 to be 15, got %d", parsed.rs1)
	}
	if parsed.rs2 != 13 {
		t.Errorf("Expected rs2 to be 13, got %d", parsed.rs2)
	}
	if parsed.imm != -12 {
		t.Errorf("Expected imm to be -12, got %d", parsed.imm)
	}
}

func TestParseUType(t *testing.T) {
	instruction := uint32(0x800006b7)
	parsed := parseUType(instruction)

	if parsed.rd != 13 {
		t.Errorf("Expected rd to be 13, got %d", parsed.rd)
	}
	if parsed.imm != 0x80000 {
		t.Errorf("Expected imm to be 0x80000, got %d", parsed.imm)
	}
}

func TestParseSType(t *testing.T) {
	instruction := uint32(0x00e60023)
	parsed := parseSType(instruction)

	if parsed.rs1 != 12 {
		t.Errorf("Expected rs1 to be 12, got %d", parsed.rs1)
	}
	if parsed.rs2 != 14 {
		t.Errorf("Expected rs2 to be 14, got %d", parsed.rs2)
	}
	if parsed.imm != 0 {
		t.Errorf("Expected imm to be 0, got %d", parsed.imm)
	}
}

func TestParseJType(t *testing.T) {
	instruction := uint32(0x0000006f)
	parsed := parseJType(instruction)

	if parsed.rd != 0 {
		t.Errorf("Expected rd to be 0, got %d", parsed.rd)
	}
	if parsed.imm != 0 {
		t.Errorf("Expected imm to be 0, got %d", parsed.imm)
	}
}
