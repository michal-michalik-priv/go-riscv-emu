# Go RISC-V Emulator

This project is a RISC-V emulator written in Go, created as a learning exercise to deepen understanding of both the Go programming language and the RISC-V architecture.

## Project Status

This emulator is still under active development, but it has progressed well beyond a basic RV32I subset.

Current implemented areas include:

- Core ISA coverage across a broad set of RV32I instructions
- RV32M (mul/div) support
- RV32A (atomic) support
- CSR and system instruction support (including key trap/counter CSRs)
- Privilege modes: Machine, Supervisor, and User
- Trap handling and interrupt delegation flow
- Timer and cycle/time counter behavior
- SBI support sufficient for early supervisor-mode software flows
- Bootloader path to enter S-mode and run raw kernel binaries

Known limitations:

- The emulator is not feature-complete and several architectural details/devices are still evolving.
- Linux support is work in progress (especially around full platform/device fidelity).

## How to Run

To run the emulator, you will need to have Go installed on your system.

1. **Build the emulator:**
   ```bash
   go build -o go-riscv-emu cmd/emulator/main.go
   ```

2. **Run an example program:**
   (Assuming you have a RISC-V ELF binary, e.g., `misc/c/terminal_mmio_write.o`)
   ```bash
   ./go-riscv-emu -dummy-tty -elf misc/c/terminal_mmio_write.o
   ```
   Replace `misc/c/terminal_mmio_write.o` with the path to your RISC-V ELF binary.

3. **Run a Linux kernel image (S-mode boot path):**
   ```bash
   ./go-riscv-emu -bootloader -kernel images/Image -kernel-addr 0x80000000
   ```

   Notes:

   - `-bootloader` enables the boot adapter path into Supervisor mode.
   - `-kernel` loads a raw kernel binary image.
   - `-kernel-addr` controls where the raw image is mapped in memory.

   All possible options:
   ```
   Usage of ./go-riscv-emu:
    -bootloader
            Enable bootloader adapter to boot into S-mode
    -debug
            Enable debug logging
    -dummy-tty
            Enable Dummy TTY device
    -elf string
            Path to the ELF file to load
    -kernel string
            Path to the raw binary kernel file to load
    -kernel-addr uint
            Address to load the raw binary kernel at (default 2147483648)
    -steps int
            Number of steps to execute (0 for infinite, default)
   ```

## Testing

Run the unit tests with:

```bash
go test ./...
```

## Author

Michał Michalik (<michal.michalik.priv@gmail.com>)