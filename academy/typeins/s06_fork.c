/* FORK -- one program becomes many processes. */
#include <stdio.h>
#include <sys/wait.h>
#include <unistd.h>

int main(void) {
    fork();                  /* 1 process becomes 2 */
    fork();                  /* each of those becomes 2 */
    printf("hello from process %d\n", getpid());
    fflush(stdout);
    while (wait(NULL) > 0)   /* parents wait for their children */
        ;
    return 0;
}
