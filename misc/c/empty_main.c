// A minimal C program with a stupidly empty main function. Compile with:
//   riscv64-unknown-elf-gcc empty_main.c -o empty_main.o -march=rv32i -mabi=ilp32 -nostdlib -ffreestanding -Wl,-T,linker.ld -Wall -Wextra -Werror -O2
// After compilation you can check the output with:
//   riscv64-unknown-elf-objdump -d -M no-aliases empty_main.o
int main() {
    return 0;
}