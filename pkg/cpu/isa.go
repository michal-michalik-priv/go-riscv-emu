package cpu

import (
	"fmt"
	"log/slog"
	"os"

	utils "github.com/michal-michalik-priv/go-riscv-emu/pkg/utils"
)

// RV32I Instruction opcodes
const (
	opcodeAddi    = 0b0010011
	opcodeJalr    = 0b1100111
	opcodeLui     = 0b0110111
	opcodeAuipc   = 0b0010111
	opcodeSb      = 0b0100011
	opcodeJal     = 0b1101111
	opcodeLb      = 0b0000011
	opcodeLbu     = 0b0000011
	opcodeBne     = 0b1100011
	opcodeOp      = 0b0110011
	opcodeMiscMem = 0b0001111
	opcodeSystem  = 0b1110011
	opcodeAtomic  = 0b0101111
)

// RV32I Funct3 for all instructions
const (
	iTypeFunc3Addi     = 0b000
	iTypeFunc3Slli     = 0b001
	iTypeFunc3Slti     = 0b010
	iTypeFunc3Sltiu    = 0b011
	iTypeFunc3SrliSrai = 0b101
	iTypeFunc3Jalr     = 0b000
	sTypeFunc3Sb       = 0b000
	sTypeFunc3Sh       = 0b001
	sTypeFunc3Sw       = 0b010
	iTypeFunc3Lb       = 0b000
	iTypeFunc3Lh       = 0b001
	iTypeFunc3Lw       = 0b010
	iTypeFunc3Lbu      = 0b100
	iTypeFunc3Lhu      = 0b101
	iTypeFunc3Xori     = 0b100
	iTypeFunc3Ori      = 0b110
	iTypeFunc3Andi     = 0b111
	bTypeFunc3Beq      = 0b000
	bTypeFunc3Bne      = 0b001
	bTypeFunc3Blt      = 0b100
	bTypeFunc3Bge      = 0b101
	bTypeFunc3Bltu     = 0b110
	bTypeFunc3Bgeu     = 0b111
	rTypeFunc3Add      = 0b000
	rTypeFunc3Sll      = 0b001
	rTypeFunc3Slt      = 0b010
	rTypeFunc3Sltu     = 0b011
	rTypeFunc3Xor      = 0b100
	rTypeFunc3SrlSra   = 0b101
	rTypeFunc3Or       = 0b110
	rTypeFunc3And      = 0b111
	rTypeFunc3Mul      = 0b000
	rTypeFunc3Mulh     = 0b001
	rTypeFunc3Mulhsu   = 0b010
	rTypeFunc3Mulhu    = 0b011
	rTypeFunc3Div      = 0b100
	rTypeFunc3Divu     = 0b101
	rTypeFunc3Rem      = 0b110
	rTypeFunc3Remu     = 0b111
	rTypeFunc3AtomicW  = 0b010
	iTypeFunc3Fence    = 0b000
	iTypeFunc3FenceI   = 0b001
	iTypeFunc3Csrrw    = 0b001
	iTypeFunc3Csrrs    = 0b010
	iTypeFunc3Csrrc    = 0b011
	iTypeFunc3Csrrwi   = 0b101
	iTypeFunc3Csrrsi   = 0b110
	iTypeFunc3Csrrci   = 0b111
)

// RV32I Privileged instruction immediate selectors
const (
	privImmEcall = 0x000
	privImmSret  = 0x102
	privImmMret  = 0x302
)

// RV32I Funct7 for all instructions
const (
	rTypeFunc7Add  = 0b0000000
	rTypeFunc7Sub  = 0b0100000
	rTypeFunc7Sll  = 0b0000000
	rTypeFunc7Slt  = 0b0000000
	rTypeFunc7Sltu = 0b0000000
	rTypeFunc7Xor  = 0b0000000
	rTypeFunc7Srl  = 0b0000000
	rTypeFunc7Sra  = 0b0100000
	rTypeFunc7Or   = 0b0000000
	rTypeFunc7And  = 0b0000000
	rTypeFunc7M    = 0b0000001
)

// RV32A Funct5 for Atomic instructions (top 5 bits of funct7)
const (
	amoFunc5Add  = 0b00000
	amoFunc5Swap = 0b00001
	amoFunc5Lr   = 0b00010
	amoFunc5Sc   = 0b00011
	amoFunc5Xor  = 0b00100
	amoFunc5Or   = 0b01000
	amoFunc5And  = 0b01100
	amoFunc5Min  = 0b10000
	amoFunc5Max  = 0b10100
	amoFunc5Minu = 0b11000
	amoFunc5Maxu = 0b11100
)

// RISC-V Special instruction encodings
const (
	instrUnimp0      = 0x00000000
	instrUnimpPseudo = 0xC0001073
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

// sll executes the SLL instruction on the given core.
func sll(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLL instruction: %+v\n", instr))
	shamt := core.GetRegister(int(instr.rs2)) & 0x1F
	val := core.GetRegister(int(instr.rs1)) << shamt
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// slt executes the SLT instruction on the given core.
func slt(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLT instruction: %+v\n", instr))
	if int32(core.GetRegister(int(instr.rs1))) < int32(core.GetRegister(int(instr.rs2))) {
		core.SetRegister(int(instr.rd), 1)
	} else {
		core.SetRegister(int(instr.rd), 0)
	}
	core.pc += 4
	return nil
}

// sltu executes the SLTU instruction on the given core.
func sltu(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLTU instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) < core.GetRegister(int(instr.rs2)) {
		core.SetRegister(int(instr.rd), 1)
	} else {
		core.SetRegister(int(instr.rd), 0)
	}
	core.pc += 4
	return nil
}

// sub executes the SUB instruction on the given core.
func sub(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SUB instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) - core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// xor executes the XOR instruction on the given core.
func xor(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing XOR instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) ^ core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// or executes the OR instruction on the given core.
func or(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing OR instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) | core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// and executes the AND instruction on the given core.
func and(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing AND instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) & core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// fence executes the FENCE instruction on the given core.
func fence(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing FENCE instruction: %+v\n", instr))
	// TODO: In case multicore support would be added, we need to be
	// more smart than simply advancing the PC...
	core.pc += 4
	return nil
}

// fenceI executes the FENCE.I instruction on the given core.
func fenceI(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing FENCE.I instruction: %+v\n", instr))
	// TODO: In case instruction cache would be added, we need to
	// flush it here.
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

// slti executes the SLTI instruction on the given core.
func slti(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLTI instruction: %+v\n", instr))
	if int32(core.GetRegister(int(instr.rs1))) < instr.imm {
		core.SetRegister(int(instr.rd), 1)
	} else {
		core.SetRegister(int(instr.rd), 0)
	}
	core.pc += 4
	return nil
}

// sltiu executes the SLTIU instruction on the given core.
func sltiu(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SLTIU instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) < uint32(instr.imm) {
		core.SetRegister(int(instr.rd), 1)
	} else {
		core.SetRegister(int(instr.rd), 0)
	}
	core.pc += 4
	return nil
}

// xori executes the XORI instruction on the given core.
func xori(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing XORI instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) ^ uint32(instr.imm)
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// ori executes the ORI instruction on the given core.
func ori(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ORI instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) | uint32(instr.imm)
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// andi executes the ANDI instruction on the given core.
func andi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing ANDI instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) & uint32(instr.imm)
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

// sh executes the SH instruction on the given core.
func sh(core *Core, instr sTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SH instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)
	value := core.GetRegister(int(instr.rs2))

	for i := uint32(0); i < 2; i++ {
		err := core.bus.Write(address+i, byte((value>>(i*8))&0xFF))
		if err != nil {
			return fmt.Errorf("SH failed at offset %d: %v", i, err)
		}
	}

	core.pc += 4
	return nil
}

// sw executes the SW instruction on the given core.
func sw(core *Core, instr sTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SW instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)
	value := core.GetRegister(int(instr.rs2))

	for i := uint32(0); i < 4; i++ {
		err := core.bus.Write(address+i, byte((value>>(i*8))&0xFF))
		if err != nil {
			return fmt.Errorf("SW failed at offset %d: %v", i, err)
		}
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

// lh executes the LH instruction on the given core.
func lh(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LH instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)

	var value uint32
	for i := uint32(0); i < 2; i++ {
		b, err := core.bus.Read(address + i)
		if err != nil {
			return fmt.Errorf("LH failed at offset %d: %v", i, err)
		}
		value |= uint32(b) << (i * 8)
	}

	core.SetRegister(int(instr.rd), uint32(utils.SignExtend(value, 16)))
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

// lhu executes the LHU instruction on the given core.
func lhu(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LHU instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)

	var value uint32
	for i := uint32(0); i < 2; i++ {
		b, err := core.bus.Read(address + i)
		if err != nil {
			return fmt.Errorf("LHU failed at offset %d: %v", i, err)
		}
		value |= uint32(b) << (i * 8)
	}

	core.SetRegister(int(instr.rd), value)
	core.pc += 4
	return nil
}

// lw executes the LW instruction on the given core.
func lw(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LW instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1)) + uint32(instr.imm)

	var value uint32
	for i := uint32(0); i < 4; i++ {
		b, err := core.bus.Read(address + i)
		if err != nil {
			return fmt.Errorf("LW failed at offset %d: %v", i, err)
		}
		value |= uint32(b) << (i * 8)
	}

	core.SetRegister(int(instr.rd), value)
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

// bge executes the BGE instruction on the given core.
func bge(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BGE instruction: %+v\n", instr))
	if int32(core.GetRegister(int(instr.rs1))) >= int32(core.GetRegister(int(instr.rs2))) {
		core.pc = core.pc + uint32(instr.imm)
	} else {
		core.pc += 4
	}
	return nil
}

// bgeu executes the BGEU instruction on the given core.
func bgeu(core *Core, instr bTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing BGEU instruction: %+v\n", instr))
	if core.GetRegister(int(instr.rs1)) >= core.GetRegister(int(instr.rs2)) {
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

// srai executes the SRAI instruction on the given core.
func srai(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SRAI instruction: %+v\n", instr))
	shamt := uint32(instr.imm) & 0x1F
	val := int32(core.GetRegister(int(instr.rs1))) >> shamt
	core.SetRegister(int(instr.rd), uint32(val))
	core.pc += 4
	return nil
}

// srl executes the SRL instruction on the given core.
func srl(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SRL instruction: %+v\n", instr))
	shamt := core.GetRegister(int(instr.rs2)) & 0x1F
	val := core.GetRegister(int(instr.rs1)) >> shamt
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// sra executes the SRA instruction on the given core.
func sra(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SRA instruction: %+v\n", instr))
	shamt := core.GetRegister(int(instr.rs2)) & 0x1F
	val := int32(core.GetRegister(int(instr.rs1))) >> shamt
	core.SetRegister(int(instr.rd), uint32(val))
	core.pc += 4
	return nil
}

// csrrw executes the CSRRW instruction on the given core.
func csrrw(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRW instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	err = core.WriteCSR(csrAddr, core.GetRegister(int(instr.rs1)))
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// csrrwi executes the CSRRWI instruction on the given core.
func csrrwi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRWI instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address
	uimm := uint32(instr.rs1)

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	err = core.WriteCSR(csrAddr, uimm)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// mret executes the MRET instruction on the given core.
func mret(core *Core) error {
	slog.Debug("Executing MRET instruction")
	if core.mode != ModeMachine {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	// Restore mode: MPP
	mstatus := core.csrs[csrMstatus]
	core.mode = (mstatus & mstatusMPP) >> 11

	// Restore interrupt enable: MIE = MPIE
	if (mstatus & mstatusMPIE) != 0 {
		core.csrs[csrMstatus] |= mstatusMIE
	} else {
		core.csrs[csrMstatus] &^= mstatusMIE
	}

	// MPIE = 1, MPP = User
	core.csrs[csrMstatus] |= mstatusMPIE
	core.csrs[csrMstatus] &^= mstatusMPP

	core.pc = core.csrs[csrMepc]
	return nil
}

// sret executes the SRET instruction on the given core.
func sret(core *Core) error {
	slog.Debug("Executing SRET instruction")
	if core.mode < ModeSupervisor {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	// Restore mode: SPP
	mstatus := core.csrs[csrMstatus]
	if (mstatus & mstatusSPP) != 0 {
		core.mode = ModeSupervisor
	} else {
		core.mode = ModeUser
	}

	// Restore interrupt enable: SIE = SPIE
	if (mstatus & mstatusSPIE) != 0 {
		core.csrs[csrMstatus] |= mstatusSIE
	} else {
		core.csrs[csrMstatus] &^= mstatusSIE
	}

	// SPIE = 1, SPP = User
	core.csrs[csrMstatus] |= mstatusSPIE
	core.csrs[csrMstatus] &^= mstatusSPP

	core.pc = core.csrs[csrSepc]
	return nil
}

// ecall executes the ECALL instruction on the given core.
func ecall(core *Core) error {
	slog.Debug("Executing ECALL instruction")

	// Check if this is a test exit (riscv-tests convention)
	a10 := core.GetRegister(10)
	a17 := core.GetRegister(17)
	sysc := uint32(93)
	if a10 == 0 && a17 == sysc {
		slog.Info("TESTS PASSED")
		os.Exit(0)
	} else if a10 != 0 && a17 == sysc {
		slog.Info(fmt.Sprintf("TESTS FAILED (no. %+v)", a10))
		os.Exit(1)
	}

	cause := uint32(0)
	switch core.mode {
	case ModeUser:
		cause = ExceptionEnvironmentCallFromUMode
	case ModeSupervisor:
		cause = ExceptionEnvironmentCallFromSMode
	case ModeMachine:
		cause = ExceptionEnvironmentCallFromMMode
	}

	core.Trap(cause, 0)
	return nil
}

// unimp executes the UNIMP pseudo-instruction on the given core.
func unimp(core *Core, instruction uint32) error {
	slog.Debug(fmt.Sprintf("Executing UNIMP instruction: %08X", instruction))
	core.Trap(ExceptionIllegalInstruction, instruction)
	return nil
}

// csrrs executes the CSRRS instruction on the given core.
func csrrs(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRS instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if instr.rs1 != 0 {
		err = core.WriteCSR(csrAddr, oldValue|core.GetRegister(int(instr.rs1)))
		if err != nil {
			core.Trap(ExceptionIllegalInstruction, 0)
			return nil
		}
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// csrrc executes the CSRRC instruction on the given core.
func csrrc(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRC instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if instr.rs1 != 0 {
		err = core.WriteCSR(csrAddr, oldValue&^core.GetRegister(int(instr.rs1)))
		if err != nil {
			core.Trap(ExceptionIllegalInstruction, 0)
			return nil
		}
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// csrrsi executes the CSRRSI instruction on the given core.
func csrrsi(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRSI instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address
	uimm := uint32(instr.rs1)

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if uimm != 0 {
		err = core.WriteCSR(csrAddr, oldValue|uimm)
		if err != nil {
			core.Trap(ExceptionIllegalInstruction, 0)
			return nil
		}
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// csrrci executes the CSRRCI instruction on the given core.
func csrrci(core *Core, instr iTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing CSRRCI instruction: %+v\n", instr))
	csrAddr := uint32(instr.imm) & 0xFFF // 12-bit CSR address
	uimm := uint32(instr.rs1)

	oldValue, err := core.ReadCSR(csrAddr)
	if err != nil {
		core.Trap(ExceptionIllegalInstruction, 0)
		return nil
	}

	if uimm != 0 {
		err = core.WriteCSR(csrAddr, oldValue&^uimm)
		if err != nil {
			core.Trap(ExceptionIllegalInstruction, 0)
			return nil
		}
	}

	if instr.rd != 0 {
		core.SetRegister(int(instr.rd), oldValue)
	}

	core.pc += 4
	return nil
}

// mul executes the MUL instruction on the given core.
func mul(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing MUL instruction: %+v\n", instr))
	val := core.GetRegister(int(instr.rs1)) * core.GetRegister(int(instr.rs2))
	core.SetRegister(int(instr.rd), val)
	core.pc += 4
	return nil
}

// mulh executes the MULH instruction on the given core.
func mulh(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing MULH instruction: %+v\n", instr))
	res := int64(int32(core.GetRegister(int(instr.rs1)))) * int64(int32(core.GetRegister(int(instr.rs2))))
	core.SetRegister(int(instr.rd), uint32(res>>32))
	core.pc += 4
	return nil
}

// mulhsu executes the MULHSU instruction on the given core.
func mulhsu(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing MULHSU instruction: %+v\n", instr))
	res := int64(int32(core.GetRegister(int(instr.rs1)))) * int64(uint64(core.GetRegister(int(instr.rs2))))
	core.SetRegister(int(instr.rd), uint32(res>>32))
	core.pc += 4
	return nil
}

// mulhu executes the MULHU instruction on the given core.
func mulhu(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing MULHU instruction: %+v\n", instr))
	res := uint64(core.GetRegister(int(instr.rs1))) * uint64(core.GetRegister(int(instr.rs2)))
	core.SetRegister(int(instr.rd), uint32(res>>32))
	core.pc += 4
	return nil
}

// div executes the DIV instruction on the given core.
func div(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing DIV instruction: %+v\n", instr))
	dividend := int32(core.GetRegister(int(instr.rs1)))
	divisor := int32(core.GetRegister(int(instr.rs2)))

	if divisor == 0 {
		core.SetRegister(int(instr.rd), 0xFFFFFFFF)
	} else if dividend == -2147483648 && divisor == -1 {
		core.SetRegister(int(instr.rd), uint32(dividend))
	} else {
		core.SetRegister(int(instr.rd), uint32(dividend/divisor))
	}
	core.pc += 4
	return nil
}

// divu executes the DIVU instruction on the given core.
func divu(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing DIVU instruction: %+v\n", instr))
	dividend := core.GetRegister(int(instr.rs1))
	divisor := core.GetRegister(int(instr.rs2))

	if divisor == 0 {
		core.SetRegister(int(instr.rd), 0xFFFFFFFF)
	} else {
		core.SetRegister(int(instr.rd), dividend/divisor)
	}
	core.pc += 4
	return nil
}

// rem executes the REM instruction on the given core.
func rem(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing REM instruction: %+v\n", instr))
	dividend := int32(core.GetRegister(int(instr.rs1)))
	divisor := int32(core.GetRegister(int(instr.rs2)))

	if divisor == 0 {
		core.SetRegister(int(instr.rd), uint32(dividend))
	} else if dividend == -2147483648 && divisor == -1 {
		core.SetRegister(int(instr.rd), 0)
	} else {
		core.SetRegister(int(instr.rd), uint32(dividend%divisor))
	}
	core.pc += 4
	return nil
}

// remu executes the REMU instruction on the given core.
func remu(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing REMU instruction: %+v\n", instr))
	dividend := core.GetRegister(int(instr.rs1))
	divisor := core.GetRegister(int(instr.rs2))

	if divisor == 0 {
		core.SetRegister(int(instr.rd), dividend)
	} else {
		core.SetRegister(int(instr.rd), dividend%divisor)
	}
	core.pc += 4
	return nil
}

// lr_w executes the LR.W instruction on the given core.
func lr_w(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing LR.W instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1))

	var value uint32
	for i := uint32(0); i < 4; i++ {
		b, err := core.bus.Read(address + i)
		if err != nil {
			return fmt.Errorf("LR.W failed at offset %d: %v", i, err)
		}
		value |= uint32(b) << (i * 8)
	}

	core.loadReservationAddr = address
	core.loadReservationValid = true

	core.SetRegister(int(instr.rd), value)
	core.pc += 4
	return nil
}

// sc_w executes the SC.W instruction on the given core.
func sc_w(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing SC.W instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1))

	if core.loadReservationValid && core.loadReservationAddr == address {
		value := core.GetRegister(int(instr.rs2))
		for i := uint32(0); i < 4; i++ {
			err := core.bus.Write(address+i, byte((value>>(i*8))&0xFF))
			if err != nil {
				return fmt.Errorf("SC.W failed at offset %d: %v", i, err)
			}
		}
		core.SetRegister(int(instr.rd), 0)
	} else {
		core.SetRegister(int(instr.rd), 1)
	}

	core.loadReservationValid = false
	core.pc += 4
	return nil
}

// amo executes the AMO instructions on the given core.
func amo(core *Core, instr rTypeInstruction) error {
	slog.Debug(fmt.Sprintf("Executing AMO instruction: %+v\n", instr))
	address := core.GetRegister(int(instr.rs1))
	funct5 := instr.func7 >> 2

	var oldValue uint32
	for i := uint32(0); i < 4; i++ {
		b, err := core.bus.Read(address + i)
		if err != nil {
			return fmt.Errorf("AMO failed at offset %d: %v", i, err)
		}
		oldValue |= uint32(b) << (i * 8)
	}

	rs2Val := core.GetRegister(int(instr.rs2))
	var newValue uint32

	switch funct5 {
	case amoFunc5Add:
		newValue = oldValue + rs2Val
	case amoFunc5Swap:
		newValue = rs2Val
	case amoFunc5Xor:
		newValue = oldValue ^ rs2Val
	case amoFunc5Or:
		newValue = oldValue | rs2Val
	case amoFunc5And:
		newValue = oldValue & rs2Val
	case amoFunc5Min:
		if int32(oldValue) < int32(rs2Val) {
			newValue = oldValue
		} else {
			newValue = rs2Val
		}
	case amoFunc5Max:
		if int32(oldValue) > int32(rs2Val) {
			newValue = oldValue
		} else {
			newValue = rs2Val
		}
	case amoFunc5Minu:
		if oldValue < rs2Val {
			newValue = oldValue
		} else {
			newValue = rs2Val
		}
	case amoFunc5Maxu:
		if oldValue > rs2Val {
			newValue = oldValue
		} else {
			newValue = rs2Val
		}
	default:
		return fmt.Errorf("unsupported AMO operation: %05b", funct5)
	}

	for i := uint32(0); i < 4; i++ {
		err := core.bus.Write(address+i, byte((newValue>>(i*8))&0xFF))
		if err != nil {
			return fmt.Errorf("AMO failed write at offset %d: %v", i, err)
		}
	}

	core.SetRegister(int(instr.rd), oldValue)
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
	case opcode == opcodeAddi && func3 == iTypeFunc3Slti:
		return slti(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Sltiu:
		return sltiu(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Xori:
		return xori(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Ori:
		return ori(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Andi:
		return andi(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3Slli:
		return slli(core, parseIType(instruction))
	case opcode == opcodeAddi && func3 == iTypeFunc3SrliSrai:
		instr := parseIType(instruction)
		// Check imm[11:5] to distinguish between SRLI and SRAI
		if (instr.imm >> 5) == 0 {
			return srli(core, instr)
		} else if (instr.imm >> 5) == 0b0100000 {
			return srai(core, instr)
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
	case opcode == opcodeSb && func3 == sTypeFunc3Sh:
		return sh(core, parseSType(instruction))
	case opcode == opcodeSb && func3 == sTypeFunc3Sw:
		return sw(core, parseSType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lb:
		return lb(core, parseIType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lh:
		return lh(core, parseIType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lw:
		return lw(core, parseIType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lbu:
		return lbu(core, parseIType(instruction))
	case opcode == opcodeLb && func3 == iTypeFunc3Lhu:
		return lhu(core, parseIType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Beq:
		return beq(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bne:
		return bne(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Blt:
		return blt(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bge:
		return bge(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bltu:
		return bltu(core, parseBType(instruction))
	case opcode == opcodeBne && func3 == bTypeFunc3Bgeu:
		return bgeu(core, parseBType(instruction))
	case opcode == opcodeOp && func3 == rTypeFunc3Add:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Add {
			return add(core, instr)
		} else if instr.func7 == rTypeFunc7Sub {
			return sub(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return mul(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3Sll:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Sll {
			return sll(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return mulh(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3Slt:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Slt {
			return slt(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return mulhsu(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3Sltu:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Sltu {
			return sltu(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return mulhu(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3Xor:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Xor {
			return xor(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return div(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3Or:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Or {
			return or(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return rem(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3And:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7And {
			return and(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return remu(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeOp && func3 == rTypeFunc3SrlSra:
		instr := parseRType(instruction)
		if instr.func7 == rTypeFunc7Srl {
			return srl(core, instr)
		} else if instr.func7 == rTypeFunc7Sra {
			return sra(core, instr)
		} else if instr.func7 == rTypeFunc7M {
			return divu(core, instr)
		}
		return fmt.Errorf("unsupported R-type instruction, %032b", instruction)
	case opcode == opcodeAtomic && func3 == rTypeFunc3AtomicW:
		instr := parseRType(instruction)
		funct5 := instr.func7 >> 2
		switch funct5 {
		case amoFunc5Lr:
			return lr_w(core, instr)
		case amoFunc5Sc:
			return sc_w(core, instr)
		default:
			return amo(core, instr)
		}
	case opcode == opcodeMiscMem && func3 == iTypeFunc3Fence:
		return fence(core, parseIType(instruction))
	case opcode == opcodeMiscMem && func3 == iTypeFunc3FenceI:
		return fenceI(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == 0:
		instr := parseIType(instruction)
		switch instr.imm {
		case privImmEcall:
			return ecall(core)
		case privImmMret:
			return mret(core)
		case privImmSret:
			return sret(core)
		default:
			return fmt.Errorf("unsupported system instruction, %032b", instruction)
		}
	case instruction == instrUnimp0 || instruction == instrUnimpPseudo:
		return unimp(core, instruction)
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrw:
		return csrrw(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrwi:
		return csrrwi(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrsi:
		return csrrsi(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrci:
		return csrrci(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrs:
		return csrrs(core, parseIType(instruction))
	case opcode == opcodeSystem && func3 == iTypeFunc3Csrrc:
		return csrrc(core, parseIType(instruction))

	default:
		return fmt.Errorf("unsupported instruction, %032b", instruction)
	}
}

// Step fetches and executes the next instruction for the given core.
func Step(core *Core) error {
	core.CheckInterrupts()
	instruction := core.Fetch()
	err := execute(core, instruction)
	if err != nil {
		return err
	}
	return nil
}
