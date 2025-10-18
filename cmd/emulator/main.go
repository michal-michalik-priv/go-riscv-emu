package main

import (
	"flag"
	"log/slog"

	"github.com/Keisim/go-riscv-emu/pkg/loader"
	"github.com/Keisim/go-riscv-emu/pkg/system"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	elfPath := flag.String("elf", "misc/c/empty_main.o", "Path to the ELF file to load")
	steps := flag.Int("steps", 2, "Number of steps to execute")
	flag.Parse()

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("Starting RISC-V RV32I Emulator")
	slog.Info("Initializing system and loading ELF file", "path", *elfPath)
	system := system.NewSystem()
	err := loader.LoadELFToSystem(*elfPath, system)
	if err != nil {
		slog.Error("Failed to load ELF file:", "error", err)
		return
	}
	slog.Info("Emulator initialized with ELF file. Starting execution...")

	for i := 0; i < *steps; i++ {
		system.Step()
	}
}
