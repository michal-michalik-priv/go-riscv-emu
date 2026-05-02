package cpu

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/devices"
)

const (
	sbiSetTimer                  = 0
	sbiConsolePutchar            = 1
	sbiLegacyConsoleGetchar      = 2
	sbiLegacyClearIPI            = 3
	sbiLegacySendIPI             = 4
	sbiLegacyRemoteFenceI        = 5
	sbiLegacyRemoteSFenceVMA     = 6
	sbiLegacyRemoteSFenceVMAASID = 7
	sbiLegacyShutdown            = 8
	sbiExtHSM                    = 0x10
	SBI_SUCCESS                  = 0
	SBI_ERR_NOT_SUPPORTED        = -2
)

const clintBase = 0x02000000

var stdinChan chan byte

// init starts asynchronous stdin buffering used by legacy SBI getchar.
func init() {
	stdinChan = make(chan byte, 256)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				stdinChan <- buf[0]
			}
		}
	}()
}

// handleSBI handles SBI calls from Supervisor mode.
// It returns true if the call was handled, false otherwise.
func handleSBI(core *Core) bool {
	if core.mode != ModeSupervisor {
		// SBI calls are only valid in Supervisor mode.
		// User mode ECALLs are Linux syscalls — let them trap to the kernel.
		// Machine mode ECALLs should also be handled normally.
		return false
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
		if char == '\n' {
			slog.Debug("SBI putchar: newline")
		}
		fmt.Print(string(char))
		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.SetRegister(11, 0)
		core.pc += 4
		return true
	case sbiLegacyConsoleGetchar: // sbi_console_getchar (legacy v0.1)
		select {
		case ch := <-stdinChan:
			core.SetRegister(10, uint32(ch))
		default:
			core.SetRegister(10, 0xFFFFFFFF) // no input available
		}
		core.pc += 4
		return true
	case sbiLegacyClearIPI: // sbi_clear_ipi (legacy v0.1)
		// Single-hart: clear local supervisor software interrupt pending (SSIP)
		core.SetInterrupt(1<<1, false)
		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.pc += 4
		return true
	case sbiLegacySendIPI: // sbi_send_ipi (legacy v0.1)
		// Single-hart: emulate IPI as local SSIP raise
		core.SetInterrupt(1<<1, true)
		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.pc += 4
		return true
	case sbiLegacyRemoteFenceI, sbiLegacyRemoteSFenceVMA, sbiLegacyRemoteSFenceVMAASID: // sbi_remote_fence_i, sbi_remote_sfence_vma, sbi_remote_sfence_vma_asid
		// No-op on single hart, no MMU
		core.SetRegister(10, uint32(SBI_SUCCESS))
		core.pc += 4
		return true
	case sbiLegacyShutdown: // sbi_shutdown (legacy v0.1)
		slog.Info("SBI shutdown requested")
		fmt.Printf("\n%s\n", core.DumpState())
		os.Exit(0)
		return true
	case sbiExtHSM: // RISC-V Hart State Management (HSM) or other extensions
		// Just return success for now
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
