/* TINY CPU -- a type-in in the spirit of 1980s magazine listings.
 * A pretend computer with 32 memory cells and one register (acc).
 * Instructions are numbers: opcode * 100 + address.
 *   1xx LOAD  acc = mem[xx]     4xx STORE mem[xx] = acc
 *   2xx ADD   acc += mem[xx]    5xx JNZ   jump to xx if acc != 0
 *   3xx SUB   acc -= mem[xx]    600 PRINT acc      000 HALT
 * The program below adds 5 + 4 + 3 + 2 + 1. */
#include <stdio.h>

int main(void) {
    int mem[32] = {
        120, 221, 420,      /* 0: sum = sum + n            */
        121, 322, 421,      /* 3: n = n - 1                */
        500,                /* 6: if n != 0 go back to 0   */
        120, 600, 0,        /* 7: print sum, then halt     */
    };
    mem[20] = 0;            /* sum */
    mem[21] = 5;            /* n   */
    mem[22] = 1;            /* the constant 1 */

    int pc = 0, acc = 0, steps = 0;
    for (;;) {
        int instr = mem[pc];            /* FETCH  */
        int op = instr / 100;           /* DECODE */
        int addr = instr % 100;
        pc++;
        steps++;
        if (steps <= 7)
            printf("step %d: pc=%d instr=%03d acc=%d\n", steps, pc - 1, instr, acc);
        switch (op) {                   /* EXECUTE */
        case 1: acc = mem[addr]; break;
        case 2: acc += mem[addr]; break;
        case 3: acc -= mem[addr]; break;
        case 4: mem[addr] = acc; break;
        case 5: if (acc != 0) pc = addr; break;
        case 6: printf("OUTPUT: %d\n", acc); break;
        case 0: printf("halted after %d steps\n", steps); return 0;
        }
    }
}
