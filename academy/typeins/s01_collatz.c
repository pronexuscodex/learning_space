/* COLLATZ -- a type-in in the spirit of 1980s magazine listings.
 * Start with n. If n is even, halve it; if odd, make it 3n + 1.
 * Count the steps until n reaches 1. */
#include <stdio.h>

int main(void) {
    long n = 27;
    int steps = 0;
    long peak = n;

    while (n != 1) {
        if (n % 2 == 0) {
            n = n / 2;
        } else {
            n = 3 * n + 1;
        }
        steps++;
        if (n > peak) {
            peak = n;
        }
    }
    printf("steps: %d\n", steps);
    printf("highest value: %ld\n", peak);
    printf("sizeof(long) on this machine: %zu bytes\n", sizeof n);
    return 0;
}
