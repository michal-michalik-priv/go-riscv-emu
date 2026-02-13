package cpu

import (
	"fmt"
	"log/slog"

	utils "github.com/michal-michalik-priv/go-riscv-emu/pkg/utils"
)

// RV32I Instruction opcodes
const (
	opcodeAddi   = 0b0010011
	opcodeJalr   = 0b1100111
	opcodeLui    = 0b0110111
	opcodeAuipc  = 0b0010111
	opcodeSb     = 0b0100011
	opcodeJal    = 0b1101111
	opcodeLb     = 0b0000011
	opcodeLbu    = 0b0000011
	opcodeBne    = 0b1100011
	opcodeOp     = 0b0110011
	opcodeSystem = 0b1110011
)

// RV32I Funct3 for all instructions
const (
	iTypeFunc3Addi     = 0b000
	iTypeFunc3Slli     = 0b001
	iTypeFunc3SrliSrai = 0b101
	iTypeFunc3Jalr     = 0b000
	sTypeFunc3Sb       = 0b000
	iTypeFunc3Lb       = 0b000
	iTypeFunc3Lbu      = 0b100
	bTypeFunc3Beq      = 0b000
	bTypeFunc3Bne      = 0b001
	bTypeFunc3Blt      = 0b100
	bTypeFunc3Bltu     = 0b110
	rTypeFunc3Add      = 0b000
	iTypeFunc3Csrrw    = 0b001
	iTypeFunc3Csrrs    = 0b010
	iTypeFunc3Csrrwi   = 0b101
)

// RV32I Privileged instruction immediate selectors
const (
	privImmSret = 0x102
	privImmMret = 0x302
)

// RV32I Funct7 for all instructions
const (
	rTypeFunc7Add = 0b0000000
)

// RISC-V CSR addresses
const (
	csrSepc    = 0x141
	csrMstatus = 0x300
	csrMepc    = 0x341
	csrMcause  = 0x342
)

// iTypeInstruction represents a parsed I-type instruction
type iTypeInstruction struct {
	rd  uint32 // Destination register
	rs1 uint32 // Source register 1
	imm int32  // Immediate value
}

// uTypeInstruction represents a parsed U-type instruction
type uTypeInstruction struct {
	rd  uint32 // Destination register
	imm int32  // Immediate value
}

// sTypeInstruction represents a parsed S-type instruction
type sTypeInstruction struct {
	rs1 uint32 // Source register 1
	rs2 uint32 // Source register 2
	imm int32  // Immediate value
}

// bTypeInstruction represents a parsed B-type instruction
type bTypeInstruction struct {
	rs1 uint32 // Source register 1
	rs2 uint32 // Source register 2
	imm int32  // Immediate value
}

// jTypeInstruction represents a parsed J-type instruction
type jTypeInstruction struct {
	rd  uint32 // Destination register
	imm int32  // Immediate value
}

// rTypeInstruction represents a parsed R-type instruction
type rTypeInstruction struct {
	rd    uint32 // Destination register
	rs1   uint32 // Source register 1
	rs2   uint32 // Source register 2
	func7 uint32 // Funct7 field
}

// parseIType parses a 32-bit I-type instruction and returns an
// iTypeInstruction struct.
func parseIType(instruction uint32) iTypeInstruction {
	rd := utils.BitsSlice(instruction, 7, 12)
	rs1 := utils.BitsSlice(instruction, 15, 20)
	imm12 := utils.BitsSlice(instruction, 20, 32)
	imm := utils.SignExtend(imm12, 12)

	return iTypeInstruction{
		rd:  rd,
		rs1: rs1,
		imm: imm,
	}
}

// parseUType parses a 32-bit U-type instruction and returns a
// uTypeInstruction struct.
func parseUType(instruction uint32) uTypeInstruction {
	rd := utils.BitsSlice(instruction, 7, 12)
	imm20 := utils.BitsSlice(instruction, 12, 32)

	return uTypeInstruction{
		rd:  rd,
		imm: int32(imm20),
	}
}

// parseSType parses a 32-bit S-type instruction and returns a
// sTypeInstruction struct.
func parseSType(instruction uint32) sTypeInstruction {
	rs1 := utils.BitsSlice(instruction, 15, 20)
	rs2 := utils.BitsSlice(instruction, 20, 25)
	imm4_0 := utils.BitsSlice(instruction, 7, 12)
	imm11_5 := utils.BitsSlice(instruction, 25, 32)
	imm := utils.SignExtend((imm11_5<<5)|imm4_0, 12)

	return sTypeInstruction{
		rs1: rs1,
		rs2: rs2,
		imm: imm,
	}
}

// parseJType parses a 32-bit J-type instruction and returns a
// jTypeInstruction struct.
func parseJType(instruction uint32) jTypeInstruction {
	rd := utils.BitsSlice(instruction, 7, 12)
	imm20 := utils.BitsSlice(instruction, 31, 32)
	imm10_1 := utils.BitsSlice(instruction, 21, 31)
	imm11 := utils.BitsSlice(instruction, 20, 21)
	imm19_12 := utils.BitsSlice(instruction, 12, 20)
	imm := utils.SignExtend((imm20<<20)|(imm19_12<<12)|(imm11<<11)|(imm10_1<<1), 21)

	return jTypeInstruction{
		rd:  rd,
		imm: int32(imm),
	}
}

// parseBType parses a 32-bit B-type instruction and returns a
// bTypeInstruction struct.
func parseBType(instruction uint32) bTypeInstruction {
	rs1 := utils.BitsSlice(instruction, 15, 20)
	rs2 := utils.BitsSlice(instruction, 20, 25)
	// Extract immediate bits according to RISC-V B-type format
	imm12 := utils.BitsSlice(instruction, 31, 32)   // instruction[31] -> imm[12]
	imm11 := utils.BitsSlice(instruction, 7, 8)     // instruction[7] -> imm[11]
	imm10_5 := utils.BitsSlice(instruction, 25, 31) // instruction[30:25] -> imm[10:5]
	imm4_1 := utils.BitsSlice(instruction, 8, 12)   // instruction[11:8] -> imm[4:1]

	// Assemble the 13-bit immediate value (imm[0] is implicitly 0)
	// imm = {imm[12], imm[11], imm[10:5], imm[4:1], 0}
	assembled_imm := (imm12 << 12) | (imm11 << 11) | (imm10_5 << 5) | (imm4_1 << 1)

	// Sign extend the 13-bit immediate value
	imm := utils.SignExtend(assembled_imm, 13)

	return bTypeInstruction{
		rs1: rs1,
		rs2: rs2,
		imm: imm,
	}
}

// parseRType parses a 32-bit R-type instruction and returns an
// rTypeInstruction struct.
func parseRType(instruction uint32) rTypeInstruction {
	rd := utils.BitsSlice(instruction, 7, 12)
	rs1 := utils.BitsSlice(instruction, 15, 20)
	rs2 := utils.BitsSlice(instruction, 20, 25)
	func7 := utils.BitsSlice(instruction, 25, 32)

	return rTypeInstruction{
		rd:    rd,
		rs1:   rs1,
		rs2:   rs2,
		func7: func7,
	}
}

// add executes the ADD instruction on the given core.
func add(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ADD instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) + core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// addi executes the ADDI instruction on the given core.
func addi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ADDI instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// jarl executes the JALR instruction on the given core.
func jarl(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing JALR instruction: %+v\n", instr))
	targetAddress := (core.GetRegister(int(instr.rs1)) + uint32(instr.imm)) &^ 1
	core.SetRegister(int(instr.rd), core.pc+4)
	core.pc = targetAddress
	return nil
}

// lui executes the LUI instruction on the given core.
func lui(core *Core, instr uTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LUI instruction: %+v\n", instr))
	core.SetRegister(int(instr.rd), uint32(instr.imm)<<12)
	core.pc += 4
	return nil
}

// auipc executes the AUIPC instruction on the given core.
func auipc(core *Core, instr uTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing AUIPC instruction: %+v\n", instr))
	core.SetRegister(int(instr.rd), core.pc+(uint32(instr.imm)<<12))
	core.pc += 4
	return nil
}

// sb executes the SB instruction on the given core.
func sb(core *Core, instr sTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SB instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)
	value := byte(core.GetRegister(int(instr.rs2)) & 0xFF)

	err := core.bus.Write(address, value)
	if err != nil {
		return fmt.Errorf("SB failed: %v", err)
	}
	core.pc += 4
	return nil
}

// jal executes the JAL instruction on the given core.
func jal(core *Core, instr jTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing JAL instruction: %+v\n", instr))
	core.SetRegister(int(instr.rd), core.pc+4)
	core.pc = core.pc + uint32(instr.imm)
	return nil
}

// lb executes the LB instruction on the given core.
func lb(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LB instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)

	value, err := core.bus.Read(address)
	if err != nil {
		return fmt.Errorf("LB failed: %v", err)
	}

	core.SetRegister(int(instr.rd), uint32(utils.SignExtend(uint32(value), 8)))
	core.pc += 4
	return nil
}

// lbu executes the LBU instruction on the given core.
func lbu(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LBU instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)

	value, err := core.bus.Read(address)
	if err != nil {
		return fmt.Errorf("LBU failed: %v", err)
	}

	core.SetRegister(int(instr.rd), uint32(value))
	core.pc += 4
	return nil
}

// beq executes the BEQ instruction on the given core.
func beq(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BEQ instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) == core.GetRegister(int(instr.rs2)) {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
	return nil
}

// bne executes the BNE instruction on the given core.
func bne(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BNE instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) != core.GetRegister(int(instr.rs2)) {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
	return nil
}

// blt executes the BLT instruction on the given core.
func blt(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BLT instruction: %+v\n", instr))
	if int32(core.GetRegister(int(instr.rs1))) < int32(core.GetRegister(int(instr.rs2))) {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
	return nil
}

// bltu executes the BLTU instruction on the given core.
func bltu(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BLTU instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) < core.GetRegister(int(instr.rs2)) {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
	return nil
}

// slli executes the SLLI instruction on the given core.
func slli(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLLI instruction: %+v\n", instr))
	shamt := uint32(instr.imm) & 0x1F
	val := core.GetRegister(int(instr.rs1)) << shamt
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// srli executes the SRLI instruction on the given core.
func srli(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SRLI instruction: %+v\n", instr))
	shamt := uint32(instr.imm) & 0x1F
	val := core.GetRegister(int(instr.rs1)) >> shamt
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// csrrw executes the CSRRW instruction on the given core.
func csrrw(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRW instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address

	// If rd is not x0, read current CSR value and write it to rd
	if instr.rd != 0 {
		oldValue := core.csrs[csrAddr]
		core.SetRegister(int(instr.rd), oldValue)
	}

	// Write value from rs1 to CSR
	core.csrs[csrAddr] = core.GetRegister(int(instr.rs1))

	core.pc += 4
	return nil
}

// csrrwi executes the CSRRWI instruction on the given core.
func csrrwi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRWI instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address
	uimm := uint32(instr.rs1)

	// If rd is not x0, read current CSR value and write it to rd
	if instr.rd != 0 {
		oldValue := core.csrs[csrAddr]
		core.SetRegister(int(instr.rd), oldValue)
	}

	// Write zero-extended 5-bit immediate to CSR
	core.csrs[csrAddr] = uimm

	core.pc += 4
	return nil
}

// mret executes the MRET instruction on the given core.
func mret(core *Core) error {
	slog.Debug("Executing MRET instruction")
	core.pc = core.csrs[csrMepc]
	return nil
}

// sret executes the SRET instruction on the given core.
func sret(core *Core) error {
	slog.Debug("Executing SRET instruction")
	core.pc = core.csrs[csrSepc]
	return nil
}

// csrrs executes the CSRRS instruction on the given core.
func csrrs(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRS instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address

	// Read current CSR value
	oldValue := core.csrs[csrAddr]

	// Write old value to destination register
	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	// If rs1 is not x0, set bits in CSR
	if instr.rs1 != 0 {
		core.csrs[csrAddr] = oldValue | core.GetRegister(int(instr.rs1))
	}

	core.pc += 4
	return nil
}

// Parse parses a 32-bit instruction word and returns the corresponding
// instruction struct based on the opcode and funct3 fields.
func execute(core *Core, instruction uint32) error {
	opcode := utils.BitsSlice(instruction, 0, 7)
	func3 := utils.BitsSlice(instruction, 12, 15)

	switch {
	case opcode == opcodeAddi && func3 == iTypeFunc3Addi:
		return addi(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Slli:
		return slli(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3SrliSrai:
		instr := parseIType(instruction)
		// Check imm[11:5] to distinguish between SRLI and SRAI
		if (instr.imm >> 5) == 0 {
			return srli(core, instr)
		}
		return fmt.Errorf("unsupported instruction, %032b", instruction)
	case opcode == opcodeJalr && func3 == iTypeFunc3Jalr:
		return jarl(core, parseIType(instruction))
	case opcode == opcodeLui: // TODO: We might check that before slicing func3
		return lui(core, parseUType(instruction))
	case opcode == opcodeAuipc:
		return auipc(core, parseUType(instruction))
	case opcode == opcodeJal:
		return jal(core, parseJType(instruction))
	case opcode == opcodeSb && func3 == sTypeFunc3Sb:
		return sb(core, parseSType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lb:
		return lb(core, parseIType(instruction))
	case opcode == opcodeLbu && func3 == iTypeFunc3Lbu:
		return lbu(core, parseIType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Beq:
		return beq(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bne:
		return bne(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Blt:
		return blt(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bltu:
		return bltu(core, parseBType(instruction))
	case opcode == opcodeOp && func3 == rTypeFunc3Add:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Add {
			return add(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeSystem && func3 == 0:
		instr := parseIType(instruction)
		switch instr.imm {
		case privImmMret:
			return mret(core)
		case privImmSret:
			return sret(core)
		default:
			return fmt.Errorf("unsupported system instruction, %032b", instruction)
		}
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrw:
		return csrrw(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrwi:
		return csrrwi(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrs:
		return csrrs(core, parseIType(instruction))

	default:
		return fmt.Errorf("unsupported instruction, %032b", instruction)
	}
}

// Step fetches and executes the next instruction for the given core.
func Step(core *Core) error {
	instruction := core.Fetch()
	err := execute(core, instruction)
	if err != nil {
		return err
	}
	return nil
}
