package cpu

import (
	"fmt"
	"log/slog"
)

const (
	sbiConsolePutchar = 1
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

	switch eid {
	case sbiConsolePutchar:
		char := byte(core.GetRegister(10)) // a0
		fmt.Print(string(char))
		core.pc += 4
		return true
	default:
		return false
	}
}
