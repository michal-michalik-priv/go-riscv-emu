package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/michal-michalik-priv/go-riscv-emu/pkg/loader"
	"github.com/michal-michalik-priv/go-riscv-emu/pkg/system"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	elfPath := flag.String("elf", "", "Path to the ELF file to load")
	kernelPath := flag.String("kernel", "", "Path to the raw binary kernel file to load")
	kernelAddr := flag.Uint("kernel-addr", 0x80000000, "Address to load the raw binary kernel at")
	steps := flag.Int("steps", 0, "Number of steps to execute (0 for infinite, default)")
	dummyTTY := flag.Bool("dummy-tty", false, "Enable Dummy TTY device")
	bootloader := flag.Bool("bootloader", false, "Enable bootloader adapter to boot into S-mode")
	flag.Parse()

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("Starting RISC-V RV32I Emulator")
	system := system.NewSystem(*dummyTTY)

	var err error
	if *kernelPath != "" {
		slog.Info("Initializing system and loading kernel binary", "path", *kernelPath, "addr", fmt.Sprintf("0x%X", *kernelAddr))
		err = loader.LoadBinaryToSystem(*kernelPath, uint32(*kernelAddr), system)
	} else {
		path := *elfPath
		if path == "" {
			path = "misc/c/empty_main.o"
		}
		slog.Info("Initializing system and loading ELF file", "path", path)
		err = loader.LoadELFToSystem(path, system)
	}

	if err != nil {
		slog.Error("Failed to load executable:", "error", err)
		return
	}

	if *bootloader {
		system.Bootloader(uint32(*kernelAddr))
	}

	slog.Info("Emulator initialized. Starting execution...")

	if *steps == 0 {
		for {
			system.Step()
		}
	} else {
		for i := 0; i < *steps; i++ {
			system.Step()
		}
	}
}
