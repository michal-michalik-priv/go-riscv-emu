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

// Core represents the CPU core with its registers and program counter.
type Core struct {
	pc   uint32
	x    [32]uint32
	csrs [4096]uint32
	bus  *devices.Bus
	mode uint32

	loadReservationAddr  uint32
	loadReservationValid bool
	wfi                  bool
	mu                   sync.Mutex
}

// NewCore creates and initializes a new CPU core with the given bus.
func NewCore(bus *devices.Bus) *Core {
	core := &Core{
		pc:   0,
		bus:  bus,
		x:    [32]uint32{},
		csrs: [4096]uint32{},
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

	byte1, _ := c.bus.Read(c.pc)
	byte2, _ := c.bus.Read(c.pc + 1)
	byte3, _ := c.bus.Read(c.pc + 2)
	byte4, _ := c.bus.Read(c.pc + 3)

	instruction := uint32(byte1) | (uint32(byte2) << 8) |
		(uint32(byte3) << 16) | (uint32(byte4) << 24)

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

	if retired {
		// Increment minstret
		oldInstret := c.csrs[csrMinstret]
		c.csrs[csrMinstret]++
		if c.csrs[csrMinstret] < oldInstret {
			c.csrs[csrMinstretH]++
		}
	}
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

	// 4. Perform MRET to transition to S-mode
	mret(c)
}
