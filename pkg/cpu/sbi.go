package cpu

import (
	"fmt"
	"log/slog"
)

const (
	sbiConsolePutchar     = 1
	SBI_SUCCESS           = 0
	SBI_ERR_NOT_SUPPORTED = -2
)

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
