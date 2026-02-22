package cpu

import (
	"fmt"
	"log/slog"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

const (
	sbiSetTimer           = 0
	sbiConsolePutchar     = 1
	SBI_SUCCESS           = 0
	SBI_ERR_NOT_SUPPORTED = -2
)

const clintBase = 0x02000000

// handleSBI handles SBI calls from Supervisor mode.
// It returns true if the call was handled, false otherwise.
func handleSBI(core *Core) bool {
	if core.mode == ModeMachine {
		slog.Warn("SBI call from invalid mode (Machine)")
		return false
	}

	if core.mode != ModeSupervisor {
		slog.Warn(fmt.Sprintf("SBI call from invalid mode: %v", core.mode))
		errCode := int32(SBI_ERR_NOT_SUPPORTED)
		core.SetRegister(10, uint32(errCode))
		core.SetRegister(11, 0)
		core.pc += 4
		return true
	}

	eid := core.GetRegister(17) // a7
	fid := core.GetRegister(16) // a6

	switch eid {
	case sbiSetTimer:
		stimeValueLow := core.GetRegister(10)  // a0
		stimeValueHigh := core.GetRegister(11) // a1

		// Write to mtimecmp (64-bit)
		addr := uint32(clintBase + devices.MTIMECMP_OFFSET)
		for i := uint32(0); i < 4; i++ {
			_ = core.bus.Write(addr+i, byte((stimeValueLow>>(i*8))&0xFF))
		}
		for i := uint32(0); i < 4; i++ {
			_ = core.bus.Write(addr+4+i, byte((stimeValueHigh>>(i*8))&0xFF))
		}

		// Clear Supervisor Timer Interrupt Pending bit (STIP = 1 << 5)
		core.SetInterrupt(1<<5, false)

		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.SetRegister(11, 0)
		core.pc += 4
		return true
	case sbiConsolePutchar:
		char := byte(core.GetRegister(10)) // a0
		fmt.Print(string(char))
		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.SetRegister(11, 0)
		core.pc += 4
		return true
	default:
		slog.Warn(fmt.Sprintf("Unsupported SBI call EID: 0x%x, FID: 0x%x", eid, fid))
		errCode := int32(SBI_ERR_NOT_SUPPORTED)
		core.SetRegister(10, uint32(errCode))
		core.SetRegister(11, 0)
		core.pc += 4
		return true
	}
}
