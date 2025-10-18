package loader

import (
	"debug/elf"
	"fmt"
	"log/slog"

	"github.com/Keisim/go-riscv-emu/pkg/system"
)

// LoadELFToSystem loads an ELF file from the specified file path into the
// provided system. It maps the ELF segments into the system's memory-mapped
// devices and sets the CPU's program counter to the ELF entry point.
func LoadELFToSystem(filePath string, sys *system.System) error {
	f, err := elf.Open(filePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening ELF file: %v", err))
	}
	defer f.Close()

	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD {
			continue
		}
		slog.Debug(fmt.Sprintf("Loading segment at 0x%X (memsz: %d, filesz: %d, offset: %d)\n",
			prog.Vaddr, prog.Memsz, prog.Filesz, prog.Off))

		device := sys.Bus().FindDevice(uint32(prog.Vaddr))
		if device == nil {
			return fmt.Errorf(
				"no device found for segment at address 0X%X", prog.Vaddr)
		}

		segmentData := make([]byte, prog.Filesz)
		_, err := prog.ReadAt(segmentData, 0)
		if err != nil {
			return fmt.Errorf(
				"error reading segment data at offset 0X%X: %v",
				prog.Off, err)
		}

		for i := uint32(0); i < uint32(prog.Filesz); i++ {
			err := device.Write(uint32(prog.Vaddr)+i, segmentData[i])
			if err != nil {
				return fmt.Errorf(
					"error writing to device at address 0X%X: %v",
					uint32(prog.Vaddr)+i, err)
			}
		}
	}

	sys.Core().SetPc(uint32(f.Entry))

	return nil
}
