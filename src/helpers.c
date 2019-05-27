/*
 * helpers.c
 *
 * Implementation of miscellaneous helper functions defined in helpers.h
 *
 * Created by Jacob Strieb
 */


#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>

#include "helpers.h"


/******************************************************************************
 * Global Variables
 *****************************************************************************/

extern int quiet;
extern int verbose;



/******************************************************************************
 * Helper Functions
 *****************************************************************************/

/*
 * Print an error message to stderr and exit with an unsuccessful status
 */
void fatal_error(const char *fmt, ...) {
  va_list args;
  va_start(args, fmt);

  vfprintf(stderr, fmt, args);

  va_end(args);

  exit(EXIT_FAILURE);
}


/*
 * Behave exactly like printf, but only print if quiet is set to false
 */
void print(const char* fmt, ...) {
  if (quiet) return;

  va_list args;
  va_start(args, fmt);

  vprintf(fmt, args);

  va_end(args);
}


/*
 * Behave exactly like printf, but only print if verbose is set to true
 */
void dbg_print(const char* fmt, ...) {
  if (!verbose) return;

  va_list args;
  va_start(args, fmt);

  vprintf(fmt, args);

  va_end(args);
}



/******************************************************************************
 * Wrapper Functions
 *****************************************************************************/

/*
 * Try to set a signal handler, and exit if it fails
 */
void Signal(int signum, sighandler_t handler) {
  if (signal(signum, handler) == SIG_ERR) {
    perror("signal");
    fatal_error("Failed to set SIGINT handler\n");
  }
}
