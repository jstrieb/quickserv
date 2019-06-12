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

// Helper functions
void fatal_error(const char *fmt, ...);
void print(const char* fmt, ...);
void dbg_print(const char* fmt, ...);
void print_wd(void);

void print_files(void);
int open_listenfd(const char *port);

#endif /* _HELPERS_H */
