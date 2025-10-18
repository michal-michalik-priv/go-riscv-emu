package cpu

import (
	"fmt"
	"log/slog"

	utils "github.com/Keisim/go-riscv-emu/pkg/utils"
)

// RV32I Instruction Opcodes
const (
	OpcodeAddi = 0b0010011
	OpcodeJalr = 0b1100111
)

// RV32I Funct3 for I-type instructions
const (
	ITypeFunc3Addi = 0b000
	ITypeFunc3Jalr = 0b000
)

// ITypeInstruction represents a parsed I-type instruction
type ITypeInstruction struct {
	rd  uint32 // Destination register
	rs1 uint32 // Source register 1
	imm int32  // Immediate value
}

// ParseIType parses a 32-bit I-type instruction and returns an
// ITypeInstruction struct.
func ParseIType(instruction uint32) ITypeInstruction {
	rd := utils.BitsSlice(instruction, 7, 12)
	rs1 := utils.BitsSlice(instruction, 15, 20)
	imm12 := utils.BitsSlice(instruction, 20, 32)
	imm := utils.SignExtend(imm12, 12)

	return ITypeInstruction{
		rd:  rd,
		rs1: rs1,
		imm: imm,
	}
}

// Addi executes the ADDI instruction on the given core.
func Addi(core *Core, instr ITypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ADDI instruction: %+v\n", instr))
	core.x[instr.rd] = core.x[instr.rs1] + uint32(instr.imm)
	core.pc += 4
	return nil
}

// Jarl executes the JALR instruction on the given core.
func Jarl(core *Core, instr ITypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing JALR instruction: %+v\n", instr))
	targetAddress := (core.x[instr.rs1] + uint32(instr.imm)) &^ 1
	core.x[instr.rd] = core.pc + 4
	core.pc = targetAddress
	return nil
}

// Parse parses a 32-bit instruction word and returns the corresponding
// instruction struct based on the opcode and funct3 fields.
func Execute(core *Core, instruction uint32) error {
	opcode := utils.BitsSlice(instruction, 0, 7)
	func3 := utils.BitsSlice(instruction, 12, 15)

	switch {
	case opcode == OpcodeAddi && func3 == ITypeFunc3Addi:
		return Addi(core, ParseIType(instruction))
	case opcode == OpcodeJalr && func3 == ITypeFunc3Jalr:
		return Jarl(core, ParseIType(instruction))

	default:
		return fmt.Errorf("unsupported instruction, %032b", instruction)
	}
}

// Step fetches and executes the next instruction for the given core.
func Step(core *Core) error {
	instruction := core.Fetch()
	err := Execute(core, instruction)
	if err != nil {
		return err
	}
	return nil
}
