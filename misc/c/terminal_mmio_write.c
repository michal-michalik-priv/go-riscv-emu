// A minimal C program with a simple memory write in the main function. Compile with:
//   riscv64-unknown-elf-gcc terminal_mmio_write.c -o terminal_mmio_write.o -march=rv32i -mabi=ilp32 -nostdlib -ffreestanding -Wl,-T,linker.ld -Wall -Wextra -Werror -O2
// After compilation you can check the output with:
//   riscv64-unknown-elf-objdump -d -M no-aliases terminal_mmio_write.o
int main() {
    char *p = (char *)0x1000;
    *p = 'A';
    return 0;
}