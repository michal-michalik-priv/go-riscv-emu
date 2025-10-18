package cpu

import (
	"fmt"
	"log/slog"

	utils "github.com/Keisim/go-riscv-emu/pkg/utils"
)

// RV32I Instruction opcodes
const (
	opcodeAddi = 0b0010011
	opcodeJalr = 0b1100111
	opcodeLui  = 0b0110111
	opcodeSb   = 0b0100011
	opcodeJal  = 0b1101111
	opcodeLb   = 0b0000011
	opcodeLbu  = 0b0000011
	opcodeBne  = 0b1100011
)

// RV32I Funct3 for all instructions
const (
	iTypeFunc3Addi = 0b000
	iTypeFunc3Jalr = 0b000
	sTypeFunc3Sb   = 0b000
	iTypeFunc3Lb   = 0b000
	iTypeFunc3Lbu  = 0b100
	bTypeFunc3Bne  = 0b001
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

// addi executes the ADDI instruction on the given core.
func addi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ADDI instruction: %+v\n", instr))
	core.x[instr.rd] = core.x[instr.rs1] + uint32(instr.imm)
	core.pc += 4
	return nil
}

// jarl executes the JALR instruction on the given core.
func jarl(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing JALR instruction: %+v\n", instr))
	targetAddress := (core.x[instr.rs1] + uint32(instr.imm)) &^ 1
	core.x[instr.rd] = core.pc + 4
	core.pc = targetAddress
	return nil
}

// lui executes the LUI instruction on the given core.
func lui(core *Core, instr uTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LUI instruction: %+v\n", instr))
	core.x[instr.rd] = uint32(instr.imm) << 12
	core.pc += 4
	return nil
}

// sb executes the SB instruction on the given core.
func sb(core *Core, instr sTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SB instruction: %+v\n", instr))
	address := core.x[instr.rs1] + uint32(instr.imm)
	value := byte(core.x[instr.rs2] & 0xFF)

	device := core.bus.FindDevice(address)
	if device == nil {
		return fmt.Errorf("SB failed: no device found at address 0x%X", address)
	}
	err := device.Write(address, value)
	if err != nil {
		return fmt.Errorf("SB failed: %v", err)
	}
	core.pc += 4
	return nil
}

// jal executes the JAL instruction on the given core.
func jal(core *Core, instr jTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing JAL instruction: %+v\n", instr))
	core.x[instr.rd] = core.pc + 4
	core.pc = core.pc + uint32(instr.imm)
	return nil
}

func lb(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LB instruction: %+v\n", instr))
	address := core.x[instr.rs1] + uint32(instr.imm)

	device := core.bus.FindDevice(address)
	if device == nil {
		return fmt.Errorf("LB failed: no device found at address 0x%X", address)
	}
	value, err := device.Read(address)
	if err != nil {
		return fmt.Errorf("LB failed: %v", err)
	}

	core.x[instr.rd] = uint32(utils.SignExtend(uint32(value), 8))
	core.pc += 4
	return nil
}

func lbu(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LBU instruction: %+v\n", instr))
	address := core.x[instr.rs1] + uint32(instr.imm)

	device := core.bus.FindDevice(address)
	if device == nil {
		return fmt.Errorf("LBU failed: no device found at address 0x%X", address)
	}
	value, err := device.Read(address)
	if err != nil {
		return fmt.Errorf("LBU failed: %v", err)
	}

	core.x[instr.rd] = uint32(value)
	core.pc += 4
	return nil
}

func bne(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BNE instruction: %+v\n", instr))
	if core.x[instr.rs1] != core.x[instr.rs2] {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
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
	case opcode == opcodeJalr && func3 == iTypeFunc3Jalr:
		return jarl(core, parseIType(instruction))
	case opcode == opcodeLui: // TODO: We might check that before slicing func3
		return lui(core, parseUType(instruction))
	case opcode == opcodeJal:
		return jal(core, parseJType(instruction))
	case opcode == opcodeSb && func3 == sTypeFunc3Sb:
		return sb(core, parseSType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lb:
		return lb(core, parseIType(instruction))
	case opcode == opcodeLbu && func3 == iTypeFunc3Lbu:
		return lbu(core, parseIType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bne:
		return bne(core, parseBType(instruction))

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
