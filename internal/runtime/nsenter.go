package runtime

/*
#cgo CFLAGS: -Wall

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <sched.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/wait.h>

__attribute__((constructor)) static void nsenter(void) {
    const char *exec_flag = getenv("__FORKER_EXEC__");
    const char *id        = getenv("__FORKER_ID__");

    if (!exec_flag || strcmp(exec_flag, "1") != 0) return;
    if (!id) { fprintf(stderr, "[nsenter] __FORKER_ID__ not set\n"); exit(1); }

    char pid_path[256];
    snprintf(pid_path, sizeof(pid_path), "/var/run/forker/%s/pid", id);

    FILE *f = fopen(pid_path, "r");
    if (!f) { fprintf(stderr, "[nsenter] open pid file: %s\n", strerror(errno)); exit(1); }

    int pid;
    if (fscanf(f, "%d", &pid) != 1) {
        fclose(f);
        fprintf(stderr, "[nsenter] bad pid file\n");
        exit(1);
    }
    fclose(f);

    const char *namespaces[] = { "uts", "ipc", "net", "pid", "mnt", NULL };

    for (int i = 0; namespaces[i]; i++) {
        char ns_path[256];
        snprintf(ns_path, sizeof(ns_path), "/proc/%d/ns/%s", pid, namespaces[i]);

        int fd = open(ns_path, O_RDONLY);
        if (fd < 0) { fprintf(stderr, "[nsenter] open %s: %s\n", ns_path, strerror(errno)); exit(1); }

        if (setns(fd, 0) != 0) {
            fprintf(stderr, "[nsenter] setns %s: %s\n", namespaces[i], strerror(errno));
            close(fd);
            exit(1);
        }
        close(fd);
    }

    pid_t child = fork();
    if (child < 0) { fprintf(stderr, "[nsenter] fork failed: %s\n", strerror(errno)); exit(1); }

    if (child > 0) {
        int status;
        waitpid(child, &status, 0);
        if (WIFEXITED(status)) exit(WEXITSTATUS(status));
        if (WIFSIGNALED(status)) exit(128 + WTERMSIG(status));
        exit(1);
    }
}
*/
import "C"
