/*
 * helpers.h
 *
 * Header file containing forward declarations of miscellaneous helper
 * functions. Additionally contains wrapper functions that handle errors
 * returned by system calls. These wrappers begin with a capital letter.
 *
 * Created by Jacob Strieb
 */

#include <stdio.h>


// Include guard so that this can only be included once
#ifndef _HELPERS_H
#define _HELPERS_H

// Macros
#define BACKLOG 128
#define MAX_FILE_NUM 20
#define MAXLINE 1024 // 1kb per line, max

// Type definitions and structs
typedef struct {
  char *line;
  char *method;
  char *target;
  char *version;
} requestline_t;

// Helper functions
void fatal_error(const char *fmt, ...);
void print(const char* fmt, ...);
void dbg_print(const char* fmt, ...);
void print_wd(void);
char *read_line(int fd);

void print_files(void);
int open_listenfd(const char *port);
requestline_t *parse_requestline(int connfd);

#endif /* _HELPERS_H */
