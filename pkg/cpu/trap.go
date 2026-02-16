package cpu

import (
	"fmt"
	"log/slog"
)

// Exception codes
const (
	ExceptionInstructionAddressMisaligned = 0
	ExceptionInstructionAccessFault       = 1
	ExceptionIllegalInstruction           = 2
	ExceptionBreakpoint                   = 3
	ExceptionLoadAddressMisaligned        = 4
	ExceptionLoadAccessFault              = 5
	ExceptionStoreAddressMisaligned       = 6
	ExceptionStoreAccessFault             = 7
	ExceptionEnvironmentCallFromUMode     = 8
	ExceptionEnvironmentCallFromSMode     = 9
	ExceptionEnvironmentCallFromMMode     = 11
	ExceptionInstructionPageFault         = 12
	ExceptionLoadPageFault                = 13
	ExceptionStorePageFault               = 15
)

// Interrupt codes (top bit set)
const (
	InterruptUSoftware = (1 << 31) | 0
	InterruptSSoftware = (1 << 31) | 1
	InterruptMSoftware = (1 << 31) | 3
	InterruptUTimer    = (1 << 31) | 4
	InterruptSTimer    = (1 << 31) | 5
	InterruptMTimer    = (1 << 31) | 7
	InterruptUExternal = (1 << 31) | 8
	InterruptSExternal = (1 << 31) | 9
	InterruptMExternal = (1 << 31) | 11
)

// Trap handles an exception or interrupt.
func (c *Core) Trap(cause uint32, tval uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trap(cause, tval)
}

func (c *Core) trap(cause uint32, tval uint32) {
	slog.Debug(fmt.Sprintf("Trap: cause=%d, tval=%X, pc=%X, mode=%d", cause, tval, c.pc, c.mode))

	// Determine if we should delegate to S-mode
	delegate := false
	if c.mode < ModeMachine {
		if (c.csrs[csrMedeleg]>>cause)&1 != 0 {
			delegate = true
		}
	}

	if delegate {
		c.handleTrapS(cause, tval)
	} else {
		c.handleTrapM(cause, tval)
	}
}

func (c *Core) handleTrapM(cause uint32, tval uint32) {
	// Save current state
	c.csrs[csrMepc] = c.pc
	c.csrs[csrMtval] = tval
	c.csrs[csrMcause] = cause

	// Update mstatus
	// MPIE = MIE
	if (c.csrs[csrMstatus] & mstatusMIE) != 0 {
		c.csrs[csrMstatus] |= mstatusMPIE
	} else {
		c.csrs[csrMstatus] &^= mstatusMPIE
	}
	// MIE = 0
	c.csrs[csrMstatus] &^= mstatusMIE
	// MPP = current mode
	c.csrs[csrMstatus] = (c.csrs[csrMstatus] &^ mstatusMPP) | (c.mode << 11)

	// Change mode to Machine
	c.mode = ModeMachine

	// Jump to vector
	mtvec := c.csrs[csrMtvec]
	isInterrupt := (cause >> 31) != 0
	causeNum := cause & 0x7FFFFFFF

	if (mtvec&1) == 0 || !isInterrupt {
		// Direct or Exception
		c.pc = mtvec &^ 3
	} else {
		// Vectored Interrupt
		c.pc = (mtvec &^ 3) + (causeNum * 4)
	}
}

func (c *Core) handleTrapS(cause uint32, tval uint32) {
	// Save current state
	c.csrs[csrSepc] = c.pc
	c.csrs[csrStval] = tval
	c.csrs[csrScause] = cause

	// Update sstatus (shadowed by mstatus)
	// SPIE = SIE
	if (c.csrs[csrMstatus] & mstatusSIE) != 0 {
		c.csrs[csrMstatus] |= mstatusSPIE
	} else {
		c.csrs[csrMstatus] &^= mstatusSPIE
	}
	// SIE = 0
	c.csrs[csrMstatus] &^= mstatusSIE
	// SPP = current mode (only 1 bit for S-mode)
	if c.mode == ModeSupervisor {
		c.csrs[csrMstatus] |= mstatusSPP
	} else {
		c.csrs[csrMstatus] &^= mstatusSPP
	}

	// Change mode to Supervisor
	c.mode = ModeSupervisor

	// Jump to vector
	stvec := c.csrs[csrStvec]
	isInterrupt := (cause >> 31) != 0
	causeNum := cause & 0x7FFFFFFF

	if (stvec&1) == 0 || !isInterrupt {
		// Direct
		c.pc = stvec &^ 3
	} else {
		// Vectored
		c.pc = (stvec &^ 3) + (causeNum * 4)
	}
}

// CheckInterrupts checks for pending and enabled interrupts.
func (c *Core) CheckInterrupts() {
	c.mu.Lock()
	defer c.mu.Unlock()

	pending := c.csrs[csrMip] & c.csrs[csrMie]
	if pending == 0 {
		return
	}

	mstatus := c.csrs[csrMstatus]

	// Check for Machine-mode interrupts
	mEnabled := (c.mode < ModeMachine) || (c.mode == ModeMachine && (mstatus&mstatusMIE) != 0)
	if mEnabled {
		mPending := pending & ((1 << 3) | (1 << 7) | (1 << 11)) // MSIP, MTIP, MEIP
		if mPending != 0 {
			var cause uint32
			if (mPending & (1 << 11)) != 0 {
				cause = InterruptMExternal
			} else if (mPending & (1 << 3)) != 0 {
				cause = InterruptMSoftware
			} else {
				cause = InterruptMTimer
			}
			c.trap(cause, 0)
			return
		}
	}

	// Check for Supervisor-mode interrupts
	sEnabled := (c.mode < ModeSupervisor) || (c.mode == ModeSupervisor && (mstatus&mstatusSIE) != 0)
	if sEnabled {
		sPending := pending & ((1 << 1) | (1 << 5) | (1 << 9)) // SSIP, STIP, SEIP
		// Filter out those delegated to S-mode
		sPending &= c.csrs[csrMideleg]
		if sPending != 0 {
			var cause uint32
			if (sPending & (1 << 9)) != 0 {
				cause = InterruptSExternal
			} else if (sPending & (1 << 1)) != 0 {
				cause = InterruptSSoftware
			} else {
				cause = InterruptSTimer
			}
			c.trap(cause, 0)
			return
		}
	}
}
