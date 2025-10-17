package system

import (
	"github.com/Keisim/go-riscv-emu/pkg/cpu"
	"github.com/Keisim/go-riscv-emu/pkg/devices"
)

const (
	RAMOffset = 0x80000000
)

type System struct {
	core        *cpu.Core
	mmioDevices devices.MMIODevices
}

func NewSystem() *System {
	mmioDevices := devices.MMIODevices{}
	ramDevice := devices.RAMDevice{}
	ramDevice.Initialize(RAMOffset, 0x10000000) // 256 MB RAM
	mmioDevices.AddDevice(&ramDevice)

	system := System{
		core:        cpu.NewCore(),
		mmioDevices: mmioDevices,
	}

	return &system
}
