/*
 * helpers.h
 *
 * Header file containing forward declarations of miscellaneous helper
 * functions.
 *
 * Created by Jacob Strieb
 */

#include <stdio.h>

// Include guard so that this can only be included once
#ifndef HELPERS_H

#define HELPERS_H

void fatal_error(const char *fmt, ...);
void print(const char* fmt, ...);
void dbg_print(const char* fmt, ...);

#endif /* HELPERS_H */
