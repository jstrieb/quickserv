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
#include <signal.h>

// Include guard so that this can only be included once
#ifndef HELPERS_H

#define HELPERS_H

// Type definitions
typedef void (*sighandler_t)(int);

// Helper functions
void fatal_error(const char *fmt, ...);
void print(const char* fmt, ...);
void dbg_print(const char* fmt, ...);

// Wrapper functions
void Signal(int signum, sighandler_t handler);

#endif /* HELPERS_H */
