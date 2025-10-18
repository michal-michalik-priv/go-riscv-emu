// A minimal C program with a simple memory write in the main function. Compile with:
//   riscv64-unknown-elf-gcc terminal_mmio_write.c -o terminal_mmio_write.o -march=rv32i -mabi=ilp32 -nostdlib -ffreestanding -Wl,-T,linker.ld -Wall -Wextra -Werror -O2
// After compilation you can check the output with:
//   riscv64-unknown-elf-objdump -d -M no-aliases terminal_mmio_write.o
#define DUMMY_TTY_MMIO 0x10000000

int main() {
    char *tty = (char *)DUMMY_TTY_MMIO;
    const char *msg = "Hello, World!";
    for (const char *p = msg; *p != '\0'; p++) {
        *tty = *p;
    }
    while(1) {};
}