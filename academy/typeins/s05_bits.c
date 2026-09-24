/* BITS -- what numbers really are inside the machine. */
#include <stdio.h>

int main(void) {
    unsigned char small = 255;
    small = small + 1;                       /* 8 bits: wraps around */
    printf("255 + 1 in 8 bits = %d\n", small);

    signed char s = 127;
    s = (signed char)(s + 1);
    printf("127 + 1 in signed 8 bits = %d\n", s);

    double a = 0.1 + 0.2;
    printf("0.1 + 0.2 == 0.3 ? %s\n", a == 0.3 ? "yes" : "no");
    printf("0.1 + 0.2 = %.17f\n", a);

    printf("sizeof(int) = %zu bytes, sizeof(double) = %zu bytes\n",
           sizeof(int), sizeof(double));
    return 0;
}
