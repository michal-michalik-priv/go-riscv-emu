# Go RISC-V Emulator

This project is a RISC-V emulator written in Go, created as a learning exercise to deepen understanding of both the Go programming language and the RISC-V architecture.

## Project Status

This emulator is currently incomplete and is under active development. It presently supports only a subset of the RV32I instruction set. Future extensions will include broader instruction set support, additional device emulation, and improved debugging capabilities.

## How to Run

To run the emulator, you will need to have Go installed on your system.

1. **Build the emulator:**
   ```bash
   go build -o go-riscv-emu cmd/emulator/main.go
   ```

2. **Run an example program:**
   (Assuming you have a RISC-V ELF binary, e.g., `misc/c/terminal_mmio_write.o`)
   ```bash
   ./go-riscv-emu misc/c/terminal_mmio_write.o
   ```
   Replace `misc/c/terminal_mmio_write.o` with the path to your RISC-V ELF binary.

   All possible options:
   ```
   Usage of ./go-riscv-emu:
    -debug
            Enable debug logging
    -dummy-tty
            Enable Dummy TTY device
    -elf string
            Path to the ELF file to load (default "misc/c/empty_main.o")
    -steps int
            Number of steps to execute (0 for infinite, default)
   ```

## Author

Michał Michalik (<michal.michalik.priv@gmail.com>)