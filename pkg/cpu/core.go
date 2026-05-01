package cpu

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

// Privilege modes
const (
	ModeUser       uint32 = 0
	ModeSupervisor uint32 = 1
	ModeMachine    uint32 = 3
)

// Core represents a RISC-V CPU core.
type Core struct {
	bus                  *devices.Bus
	pc                   uint32
	x                    [32]uint32
	csrs                 [4096]uint32
	mode                 uint32
	wfi                  bool
	mu                   sync.Mutex
	instretSuppressed    bool // Suppress instret increment for next instruction
	loadReservationAddr  uint32
	loadReservationValid bool
}

// NewCore creates and initializes a new CPU core with the given bus.
func NewCore(bus *devices.Bus) *Core {
	core := &Core{
		pc:   0,
		bus:  bus,
		x:    [32]uint32{},
		mode: ModeMachine,

		loadReservationAddr:  0,
		loadReservationValid: false,
		wfi:                  false,
		mu:                   sync.Mutex{},
	}

	// Initialize MISA: RV32IMASU
	// Bits 31:30 = 1 (RV32)
	// Bit 0 = A (Atomic Extension)
	// Bit 8 = I (Base Integer)
	// Bit 12 = M (Multiply/Divide)
	// Bit 18 = S (Supervisor Mode)
	// Bit 20 = U (User Mode)
	core.csrs[csrMisa] = (1 << 30) | (1 << 0) | (1 << 8) | (1 << 12) | (1 << 18) | (1 << 20)

	return core
}

// SetPc sets the program counter to the specified value.
func (c *Core) SetPc(value uint32) {
	c.pc = value
}

// GetPc returns the current value of the program counter.
func (c *Core) GetPc() uint32 {
	return c.pc
}

// SetRegister sets the value of a specific general-purpose register.
func (c *Core) SetRegister(index int, value uint32) {
	if index > 0 && index < 32 {
		c.x[index] = value
	}
}

// GetRegister returns the value of a specific general-purpose register.
func (c *Core) GetRegister(index int) uint32 {
	if index >= 0 && index < 32 {
		return c.x[index]
	}
	return 0
}

// Fetch retrieves the next instruction from memory at the current PC.
func (c *Core) Fetch() uint32 {
	slog.Debug(fmt.Sprintf("Fetching instruction at PC: %X", c.pc))

	byte1, err1 := c.bus.Read(c.pc)
	byte2, err2 := c.bus.Read(c.pc + 1)
	byte3, err3 := c.bus.Read(c.pc + 2)
	byte4, err4 := c.bus.Read(c.pc + 3)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		slog.Error(fmt.Sprintf("Error fetching instruction at PC %X: %v, %v, %v, %v", c.pc, err1, err2, err3, err4))
	}

	instruction := uint32(byte1) | (uint32(byte2) << 8) |
		(uint32(byte3) << 16) | (uint32(byte4) << 24)

	slog.Debug(fmt.Sprintf("Fetched instruction: %08X at PC %X", instruction, c.pc))

	return instruction
}

// SetInterrupt sets or clears an interrupt bit in the mip CSR.
func (c *Core) SetInterrupt(bit uint32, active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if active {
		c.csrs[csrMip] |= bit
	} else {
		c.csrs[csrMip] &^= bit
	}
}

// IncrementCounters increments the mcycle and minstret CSRs.
func (c *Core) IncrementCounters(retired bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Increment mcycle
	oldCycle := c.csrs[csrMcycle]
	c.csrs[csrMcycle]++
	if c.csrs[csrMcycle] < oldCycle {
		c.csrs[csrMcycleH]++
	}

	if retired && !c.instretSuppressed {
		// Increment minstret
		oldInstret := c.csrs[csrMinstret]
		c.csrs[csrMinstret]++
		if c.csrs[csrMinstret] < oldInstret {
			c.csrs[csrMinstretH]++
		}
	}

	// Clear the suppression flag after checking
	c.instretSuppressed = false
}

// Bootloader sets up the CPU state to boot into Supervisor mode.
func (c *Core) Bootloader(entryPoint uint32) {
	slog.Info("Running bootloader adapter to boot into S-mode", "entryPoint", fmt.Sprintf("0x%X", entryPoint))
	// 1. Set mstatus.MPP to Supervisor mode (1)
	c.csrs[csrMstatus] = (c.csrs[csrMstatus] & ^uint32(mstatusMPP)) | (ModeSupervisor << 11)

	// 2. Set mepc to the entry point
	c.csrs[csrMepc] = entryPoint

	// 3. Delegate interrupts and exceptions to S-mode
	c.csrs[csrMideleg] = 0xFFFF
	c.csrs[csrMedeleg] = 0xFFFF

	// 4. Allow S-mode access to counters/timers (CY=0, TM=1, IR=2)
	c.csrs[csrMcounteren] = 0x7

	// 5. Perform MRET to transition to S-mode
	mret(c)
}

// GetRegisters returns all general-purpose registers.
func (c *Core) GetRegisters() [32]uint32 {
	return c.x
}

// DumpState returns a string representation of the current CPU state.
func (c *Core) DumpState() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	modeStr := ""
	switch c.mode {
	case ModeUser:
		modeStr = "User"
	case ModeSupervisor:
		modeStr = "Supervisor"
	case ModeMachine:
		modeStr = "Machine"
	default:
		modeStr = fmt.Sprintf("Unknown (%d)", c.mode)
	}

	state := fmt.Sprintf("CPU State (Mode: %s):\n", modeStr)
	state += fmt.Sprintf("PC: 0x%08X\n", c.pc)

	state += "Registers:\n"
	for i := 0; i < 32; i++ {
		state += fmt.Sprintf(" x%-2d: 0x%08X", i, c.x[i])
		if (i+1)%4 == 0 {
			state += "\n"
		}
	}

	state += "Key CSRs:\n"
	state += fmt.Sprintf(" mstatus:  0x%08X  mtvec:    0x%08X  mepc:     0x%08X  mcause:   0x%08X\n",
		c.csrs[csrMstatus], c.csrs[csrMtvec], c.csrs[csrMepc], c.csrs[csrMcause])
	state += fmt.Sprintf(" mie:      0x%08X  mip:      0x%08X  mideleg:  0x%08X  medeleg:  0x%08X\n",
		c.csrs[csrMie], c.csrs[csrMip], c.csrs[csrMideleg], c.csrs[csrMedeleg])
	state += fmt.Sprintf(" sstatus:  0x%08X  stvec:    0x%08X  sepc:     0x%08X  scause:   0x%08X\n",
		c.csrs[csrSstatus], c.csrs[csrStvec], c.csrs[csrSepc], c.csrs[csrScause])
	state += fmt.Sprintf(" sie:      0x%08X  sip:      0x%08X\n",
		c.csrs[csrSie], c.csrs[csrSip])

	return state
}
