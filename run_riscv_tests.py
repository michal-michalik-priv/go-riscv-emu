import subprocess, shlex
import re

rv32ui_p_tests = '''
rv32ui-p-add
rv32ui-p-addi
rv32ui-p-and
rv32ui-p-andi
rv32ui-p-auipc
rv32ui-p-beq
rv32ui-p-bge
rv32ui-p-bgeu
rv32ui-p-blt
rv32ui-p-bltu
rv32ui-p-bne
rv32ui-p-fence_i
rv32ui-p-jal
rv32ui-p-jalr
rv32ui-p-lb
rv32ui-p-lbu
rv32ui-p-ld_st
rv32ui-p-lh
rv32ui-p-lhu
rv32ui-p-lui
rv32ui-p-lw
rv32ui-p-ma_data
rv32ui-p-or
rv32ui-p-ori
rv32ui-p-sb
rv32ui-p-sh
rv32ui-p-simple
rv32ui-p-sll
rv32ui-p-slli
rv32ui-p-slt
rv32ui-p-slti
rv32ui-p-sltiu
rv32ui-p-sltu
rv32ui-p-sra
rv32ui-p-srai
rv32ui-p-srl
rv32ui-p-srli
rv32ui-p-st_ld
rv32ui-p-sub
rv32ui-p-sw
rv32ui-p-xor
rv32ui-p-xori
rv32um-p-div
rv32um-p-divu
rv32um-p-mul
rv32um-p-mulh
rv32um-p-mulhsu
rv32um-p-mulhu
rv32um-p-rem
rv32um-p-remu
rv32mi-p-breakpoint
rv32mi-p-csr
#rv32mi-p-illegal
rv32mi-p-instret_overflow
rv32mi-p-lh-misaligned
rv32mi-p-lw-misaligned
rv32mi-p-ma_addr
rv32mi-p-ma_fetch
rv32mi-p-mcsr
rv32mi-p-pmpaddr
rv32mi-p-sbreak
rv32mi-p-scall
rv32mi-p-sh-misaligned
rv32mi-p-shamt
rv32mi-p-sw-misaligned
rv32mi-p-zicntr
rv32si-p-csr
rv32si-p-dirty
rv32si-p-ma_fetch
rv32si-p-sbreak
rv32si-p-scall
rv32si-p-wfi
rv32ua-p-amoadd_w
rv32ua-p-amoand_w
rv32ua-p-amomax_w
rv32ua-p-amomaxu_w
rv32ua-p-amomin_w
rv32ua-p-amominu_w
rv32ua-p-amoor_w
rv32ua-p-amoswap_w
rv32ua-p-amoxor_w
rv32ua-p-lrsc
'''.strip().splitlines()

failed = []
for test in rv32ui_p_tests:
    if test.startswith('#'):
        print(f'{test:<25} SKIP !!')
        failed.append(test)
        continue
    cmd = f'go run cmd/emulator/main.go -elf /Users/keisim/Projects/riscv-tests/isa/{test}'
    out = subprocess.run(shlex.split(cmd), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if out.returncode != 0:
        re_testno = re.search(r'INFO TESTS FAILED \(no\. (\d+)\)', out.stderr.decode())
        print(f'{test:<25} FAIL !! {"test no: " + re_testno.group(1) if re_testno else 'N/A'}')
        failed.append(test)
    else:
        print(f'{test:<25} PASS')

print(f"=== FAILED OR SKIPPED {len(failed)} (out of {len(rv32ui_p_tests)})")
print(f"=== Failed tests list: {', '.join(failed)}")