package cpu

import (
	"strings"
	"testing"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
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

func TestSlli(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 2 // Set register x1 to 2

	instr := iTypeInstruction{
		rd:  2, // Destination register x2
		rs1: 1, // Source register x1
		imm: 4, // Shift amount 4
	}

	err := slli(core, instr)
	if err != nil {
		t.Fatalf("slli failed: %v", err)
	}

	expected := uint32(32) // 2 << 4
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestExecute_Slli(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 2

	// SLLI x2, x1, 4 -> 0x00409113
	instruction := uint32(0x00409113)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(32)
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestSrli(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 32 // Set register x1 to 32

	instr := iTypeInstruction{
		rd:  2, // Destination register x2
		rs1: 1, // Source register x1
		imm: 4, // Shift amount 4
	}

	err := srli(core, instr)
	if err != nil {
		t.Fatalf("srli failed: %v", err)
	}

	expected := uint32(2) // 32 >> 4
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestExecute_Srli(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 32

	// SRLI x2, x1, 4 -> 0x0040D113
	instruction := uint32(0x0040D113)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(2)
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %d, got %d", expected, core.x[2])
	}
}

func TestAdd(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 10
	core.x[2] = 20

	instr := rTypeInstruction{
		rd:  3,
		rs1: 1,
		rs2: 2,
	}

	err := add(core, instr)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	expected := uint32(30)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %d, got %d", expected, core.x[3])
	}
}

func TestSub(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 30
	core.x[2] = 10

	instr := rTypeInstruction{
		rd:  3,
		rs1: 1,
		rs2: 2,
	}

	err := sub(core, instr)
	if err != nil {
		t.Fatalf("sub failed: %v", err)
	}

	expected := uint32(20)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %d, got %d", expected, core.x[3])
	}
}

func TestXor(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 0b1010
	core.x[2] = 0b1100

	instr := rTypeInstruction{
		rd:  3,
		rs1: 1,
		rs2: 2,
	}

	err := xor(core, instr)
	if err != nil {
		t.Fatalf("xor failed: %v", err)
	}

	expected := uint32(0b0110)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %b, got %b", expected, core.x[3])
	}
}

func TestXori(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 0b1010

	instr := iTypeInstruction{
		rd:  2,
		rs1: 1,
		imm: 0b1100,
	}

	err := xori(core, instr)
	if err != nil {
		t.Fatalf("xori failed: %v", err)
	}

	expected := uint32(0b0110)
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %b, got %b", expected, core.x[2])
	}
}

func TestFence(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x1000

	instr := iTypeInstruction{} // Fields don't matter for NOP implementation

	err := fence(core, instr)
	if err != nil {
		t.Fatalf("fence failed: %v", err)
	}

	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}
}

func TestExecute_Fence(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x1000

	// FENCE -> 0x0ff0000f
	instruction := uint32(0x0ff0000f)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.pc != 0x1004 {
		t.Errorf("Expected PC to be 0x1004, got %X", core.pc)
	}
}

func TestExecute_Add(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 10
	core.x[2] = 20

	// ADD x3, x1, x2 -> 0x002081B3
	instruction := uint32(0x002081B3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(30)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %d, got %d", expected, core.x[3])
	}
}

func TestExecute_Sub(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 30
	core.x[2] = 10

	// SUB x3, x1, x2 -> 0x402081B3
	instruction := uint32(0x402081B3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(20)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %d, got %d", expected, core.x[3])
	}
}

func TestExecute_Xor(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 0b1010
	core.x[2] = 0b1100

	// XOR x3, x1, x2 -> 0x0020C1B3
	instruction := uint32(0x0020C1B3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(0b0110)
	if core.x[3] != expected {
		t.Errorf("Expected x3 to be %b, got %b", expected, core.x[3])
	}
}

func TestExecute_Xori(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.x[1] = 0b1010

	// XORI x2, x1, 0b1100 -> 0x00C0C113
	instruction := uint32(0x00C0C113)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	expected := uint32(0b0110)
	if core.x[2] != expected {
		t.Errorf("Expected x2 to be %b, got %b", expected, core.x[2])
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

func TestAuipc(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x1000

	instr := uTypeInstruction{
		rd:  4,       // Destination register x4
		imm: 0x12345, // Immediate value
	}

	err := auipc(core, instr)
	if err != nil {
		t.Fatalf("auipc failed: %v", err)
	}

	expected := uint32(0x12345000 + 0x1000)
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

func TestBeq(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000

	// Case 1: rs1 == rs2, branch taken
	core.x[1] = 10
	core.x[2] = 10
	instr := bTypeInstruction{
		rs1: 1,
		rs2: 2,
		imm: 0x100,
	}
	err := beq(core, instr)
	if err != nil {
		t.Fatalf("beq failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}

	// Case 2: rs1 != rs2, branch not taken
	core.pc = 0x3000
	core.x[2] = 20
	err = beq(core, instr)
	if err != nil {
		t.Fatalf("beq failed: %v", err)
	}
	if core.pc != 0x3004 {
		t.Errorf("Expected PC to be 0x3004, got %X", core.pc)
	}
}

func TestExecute_Beq(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000
	core.x[1] = 10
	core.x[2] = 10

	// BEQ x1, x2, 0x100 -> 0x10208063
	instruction := uint32(0x10208063)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
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

func TestMret(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.csrs[csrMepc] = 0x4000 // mepc

	err := mret(core)
	if err != nil {
		t.Fatalf("mret failed: %v", err)
	}

	if core.pc != 0x4000 {
		t.Errorf("Expected PC to be 0x4000, got %X", core.pc)
	}
}

func TestExecute_Mret(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.csrs[csrMepc] = 0x5000 // mepc

	// MRET -> 0x30200073
	instruction := uint32(0x30200073)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.pc != 0x5000 {
		t.Errorf("Expected PC to be 0x5000, got %X", core.pc)
	}
}

func TestSret(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.csrs[csrSepc] = 0x6000 // sepc

	err := sret(core)
	if err != nil {
		t.Fatalf("sret failed: %v", err)
	}

	if core.pc != 0x6000 {
		t.Errorf("Expected PC to be 0x6000, got %X", core.pc)
	}
}

func TestExecute_Sret(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.csrs[csrSepc] = 0x7000 // sepc

	// SRET -> 0x10200073
	instruction := uint32(0x10200073)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.pc != 0x7000 {
		t.Errorf("Expected PC to be 0x7000, got %X", core.pc)
	}
}

func TestExecute_Ecall(t *testing.T) {
	core := NewCore(&devices.Bus{})

	// ECALL -> 0x00000073
	instruction := uint32(0x00000073)
	err := execute(core, instruction)
	if err == nil {
		t.Fatal("Expected error from ECALL, got nil")
	}

	if err.Error() != "ECALL triggered" {
		t.Errorf("Expected 'ECALL triggered' error, got %v", err)
	}
}

func TestExecute_Unimp(t *testing.T) {
	core := NewCore(&devices.Bus{})

	// UNIMP (0x00000000)
	instruction := uint32(0x00000000)
	err := execute(core, instruction)
	if err == nil {
		t.Fatal("Expected error from UNIMP (0x00000000), got nil")
	}
	if !strings.Contains(err.Error(), "UNIMP instruction encountered") {
		t.Errorf("Unexpected error message: %v", err)
	}

	// UNIMP (0xC0001073)
	instruction = uint32(0xC0001073)
	err = execute(core, instruction)
	if err == nil {
		t.Fatal("Expected error from UNIMP (0xC0001073), got nil")
	}
	if !strings.Contains(err.Error(), "UNIMP instruction encountered") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestExecute_Bne(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000
	core.x[1] = 10
	core.x[2] = 20

	// BNE x1, x2, 0x100
	// Opcode: 1100011 (0x63)
	// rs1: 1, rs2: 2, funct3: 001, imm: 0x100
	// 0x10209063
	instruction := uint32(0x10209063)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}
}

func TestBlt(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000

	// Case 1: rs1 < rs2 (signed), branch taken
	core.x[1] = 0xFFFFFFF6 // -10
	core.x[2] = 0x00000005 // 5
	instr := bTypeInstruction{
		rs1: 1,
		rs2: 2,
		imm: 0x100,
	}
	err := blt(core, instr)
	if err != nil {
		t.Fatalf("blt failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}

	// Case 2: rs1 >= rs2 (signed), branch not taken
	core.pc = 0x3000
	core.x[1] = 0x00000005 // 5
	core.x[2] = 0xFFFFFFF6 // -10
	err = blt(core, instr)
	if err != nil {
		t.Fatalf("blt failed: %v", err)
	}
	if core.pc != 0x3004 {
		t.Errorf("Expected PC to be 0x3004, got %X", core.pc)
	}
}

func TestExecute_Blt(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000
	core.x[1] = 0xFFFFFFF6 // -10
	core.x[2] = 0x00000005 // 5

	// BLT x1, x2, 0x100 -> 0x1020C063
	instruction := uint32(0x1020C063)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}
}

func TestBltu(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000

	// Case 1: rs1 < rs2 (unsigned), branch taken
	core.x[1] = 0x00000005 // 5
	core.x[2] = 0xFFFFFFF6 // large unsigned
	instr := bTypeInstruction{
		rs1: 1,
		rs2: 2,
		imm: 0x100,
	}
	err := bltu(core, instr)
	if err != nil {
		t.Fatalf("bltu failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}

	// Case 2: rs1 >= rs2 (unsigned), branch not taken
	core.pc = 0x3000
	core.x[1] = 0xFFFFFFF6 // large unsigned
	core.x[2] = 0x00000005 // 5
	err = bltu(core, instr)
	if err != nil {
		t.Fatalf("bltu failed: %v", err)
	}
	if core.pc != 0x3004 {
		t.Errorf("Expected PC to be 0x3004, got %X", core.pc)
	}
}

func TestExecute_Bltu(t *testing.T) {
	core := NewCore(&devices.Bus{})
	core.pc = 0x3000
	core.x[1] = 0x00000005 // 5
	core.x[2] = 0xFFFFFFF6 // large unsigned

	// BLTU x1, x2, 0x100 -> 0x1020E063
	instruction := uint32(0x1020E063)
	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if core.pc != 0x3100 {
		t.Errorf("Expected PC to be 0x3100, got %X", core.pc)
	}
}

func TestCsrrw(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMstatus)
	core.csrs[csrAddr] = 0x1
	core.x[1] = 0x8 // new value

	instr := iTypeInstruction{
		rd:  2,
		rs1: 1,
		imm: int32(csrAddr),
	}

	err := csrrw(core, instr)
	if err != nil {
		t.Fatalf("csrrw failed: %v", err)
	}

	if core.x[2] != 0x1 {
		t.Errorf("Expected x2 to be 0x1, got %X", core.x[2])
	}
	if core.csrs[csrAddr] != 0x8 {
		t.Errorf("Expected CSR %X to be 0x8, got %X", csrAddr, core.csrs[csrAddr])
	}
}

func TestCsrrw_RdZero(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMstatus)
	core.csrs[csrAddr] = 0x1
	core.x[1] = 0x8 // new value

	instr := iTypeInstruction{
		rd:  0, // x0, should not read from CSR (though in our implementation it just doesn't write to rd)
		rs1: 1,
		imm: int32(csrAddr),
	}

	err := csrrw(core, instr)
	if err != nil {
		t.Fatalf("csrrw failed: %v", err)
	}

	if core.csrs[csrAddr] != 0x8 {
		t.Errorf("Expected CSR %X to be 0x8, got %X", csrAddr, core.csrs[csrAddr])
	}
}

func TestExecute_Csrrw(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMcause)
	core.csrs[csrAddr] = 0x5
	core.x[10] = 0x9

	// CSRRW x11, mcause, x10
	// opcode: 1110011 (73)
	// rd: 11 (0x0B)
	// funct3: 001 (1)
	// rs1: 10 (0x0A)
	// csr: 0x342
	// 0x342 0A 1 0B 73 -> 0x342515F3
	instruction := uint32(0x342515F3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.x[11] != 0x5 {
		t.Errorf("Expected x11 to be 0x5, got %X", core.x[11])
	}
	if core.csrs[csrAddr] != 0x9 {
		t.Errorf("Expected CSR 0x342 to be 0x9, got %X", core.csrs[csrAddr])
	}
}

func TestCsrrwi(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMstatus)
	core.csrs[csrAddr] = 0x1

	instr := iTypeInstruction{
		rd:  2,
		rs1: 0x1F, // uimm
		imm: int32(csrAddr),
	}

	err := csrrwi(core, instr)
	if err != nil {
		t.Fatalf("csrrwi failed: %v", err)
	}

	if core.x[2] != 0x1 {
		t.Errorf("Expected x2 to be 0x1, got %X", core.x[2])
	}
	if core.csrs[csrAddr] != 0x1F {
		t.Errorf("Expected CSR 0x300 to be 0x1F, got %X", core.csrs[csrAddr])
	}
}

func TestExecute_Csrrwi(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMcause)
	core.csrs[csrAddr] = 0x5

	// CSRRWI x11, mcause, 0x1A
	// opcode: 1110011 (73)
	// rd: 11 (0x0B)
	// funct3: 101 (5)
	// uimm: 11010 (0x1A)
	// csr: 0x342
	// 0x342 1A 5 0B 73 -> 0x342D55F3
	instruction := uint32(0x342D55F3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.x[11] != 0x5 {
		t.Errorf("Expected x11 to be 0x5, got %X", core.x[11])
	}
	if core.csrs[csrAddr] != 0x1A {
		t.Errorf("Expected CSR 0x342 to be 0x1A, got %X", core.csrs[csrAddr])
	}
}

func TestCsrrs(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMstatus)
	core.csrs[csrAddr] = 0x1
	core.x[1] = 0x8 // bit to set

	instr := iTypeInstruction{
		rd:  2,
		rs1: 1,
		imm: int32(csrAddr),
	}

	err := csrrs(core, instr)
	if err != nil {
		t.Fatalf("csrrs failed: %v", err)
	}

	if core.x[2] != 0x1 {
		t.Errorf("Expected x2 to be 0x1, got %X", core.x[2])
	}
	if core.csrs[csrAddr] != 0x9 { // 0x1 | 0x8
		t.Errorf("Expected CSR 0x300 to be 0x9, got %X", core.csrs[csrAddr])
	}
}

func TestCsrrs_ReadOnce(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMstatus)
	core.csrs[csrAddr] = 0x1

	instr := iTypeInstruction{
		rd:  2,
		rs1: 0, // x0, should not write to CSR
		imm: int32(csrAddr),
	}

	err := csrrs(core, instr)
	if err != nil {
		t.Fatalf("csrrs failed: %v", err)
	}

	if core.x[2] != 0x1 {
		t.Errorf("Expected x2 to be 0x1, got %X", core.x[2])
	}
	if core.csrs[csrAddr] != 0x1 { // Should not change
		t.Errorf("Expected CSR 0x300 to be 0x1, got %X", core.csrs[csrAddr])
	}
}

func TestExecute_Csrrs(t *testing.T) {
	core := NewCore(&devices.Bus{})
	csrAddr := uint32(csrMcause)
	core.csrs[csrAddr] = 0x5
	core.x[10] = 0x2

	// CSRRS x11, mcause, x10
	// opcode: 1110011 (73)
	// rd: 11 (0x0B)
	// funct3: 010 (2)
	// rs1: 10 (0x0A)
	// csr: 0x342
	// 0x342 0A 2 0B 73 -> 0x342525F3
	instruction := uint32(0x342525F3)

	err := execute(core, instruction)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if core.x[11] != 0x5 {
		t.Errorf("Expected x11 to be 0x5, got %X", core.x[11])
	}
	if core.csrs[csrAddr] != 0x7 { // 0x5 | 0x2
		t.Errorf("Expected CSR 0x342 to be 0x7, got %X", core.csrs[csrAddr])
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

func TestParseRType(t *testing.T) {
	// ADD x3, x1, x2 -> 0x002081B3
	instruction := uint32(0x002081B3)
	parsed := parseRType(instruction)

	if parsed.rd != 3 {
		t.Errorf("Expected rd to be 3, got %d", parsed.rd)
	}
	if parsed.rs1 != 1 {
		t.Errorf("Expected rs1 to be 1, got %d", parsed.rs1)
	}
	if parsed.rs2 != 2 {
		t.Errorf("Expected rs2 to be 2, got %d", parsed.rs2)
	}
	if parsed.func7 != 0 {
		t.Errorf("Expected func7 to be 0, got %d", parsed.func7)
	}
}
