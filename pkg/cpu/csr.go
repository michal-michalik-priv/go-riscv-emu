package cpu

import (
	"fmt"
)

// CSR addresses (subset)
const (
	// User CSRs
	csrUstatus = 0x000

	// Supervisor CSRs
	csrSstatus = 0x100
	csrStvec   = 0x105
	csrSepc    = 0x141
	csrScause  = 0x142
	csrStval   = 0x143

	// Machine CSRs
	csrMstatus = 0x300
	csrMisa    = 0x301
	csrMedeleg = 0x302
	csrMideleg = 0x303
	csrMie     = 0x304
	csrMtvec   = 0x305
	csrMepc    = 0x341
	csrMcause  = 0x342
	csrMtval   = 0x343
	csrMip     = 0x344

	// Machine Information Registers
	csrMvendorid = 0xF11
	csrMarchid   = 0xF12
	csrMimpid    = 0xF13
	csrMhartid   = 0xF14
)

// mstatus/sstatus bit masks (simplified)
const (
	mstatusSIE  = 1 << 1
	mstatusMIE  = 1 << 3
	mstatusSPIE = 1 << 5
	mstatusMPIE = 1 << 7
	mstatusSPP  = 1 << 8
	mstatusMPP  = 3 << 11
	mstatusTW   = 1 << 21
)

// CanAccessCSR checks if the current mode has access to the given CSR address.
func (c *Core) CanAccessCSR(address uint32, write bool) bool {
	priv := (address >> 8) & 0x3
	readOnly := ((address >> 10) & 0x3) == 0x3

	if write && readOnly {
		return false
	}

	return c.mode >= priv
}

// ReadCSR reads a CSR value after checking privilege.
func (c *Core) ReadCSR(address uint32) (uint32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.CanAccessCSR(address, false) {
		return 0, fmt.Errorf("CSR read access denied: %03X", address)
	}

	// Some CSRs are shadows of others or have special logic
	switch address {
	case csrSstatus:
		return c.csrs[csrMstatus] & 0x800DE122, nil // Simplified mask for sstatus
	default:
		return c.csrs[address&0xFFF], nil
	}
}

// WriteCSR writes a CSR value after checking privilege and handling side effects.
func (c *Core) WriteCSR(address uint32, value uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.CanAccessCSR(address, true) {
		return fmt.Errorf("CSR write access denied: %03X", address)
	}

	address = address & 0xFFF

	switch address {
	case csrSstatus:
		// Write to sstatus actually writes to mstatus with a mask
		mask := uint32(0x800DE122)
		c.csrs[csrMstatus] = (c.csrs[csrMstatus] & ^mask) | (value & mask)
	case csrMstatus:
		c.csrs[csrMstatus] = value
	case csrMisa, csrMvendorid, csrMarchid, csrMimpid, csrMhartid:
		// Read-only or WARL with no writable bits in this implementation
		return nil
	default:
		c.csrs[address] = value
	}

	return nil
}
