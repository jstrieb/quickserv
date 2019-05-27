/*
 * quickserv.c
 *
 * Main file
 *
 * Created by Jacob Strieb
 */


#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

// Include miscellaneous helper functions
#include "helpers.h"

// Include contracts for debugging if the DEBUG flag is used when compiling
#include "contracts.h"


/******************************************************************************
 * Global Variables
 *****************************************************************************/

// Both variables are set to maximum verbosity until command-line options to
// control them are added
int quiet = 0;
int verbose = 1;



/******************************************************************************
 * Signal Handlers
 *****************************************************************************/

/*
 * Handle Ctrl-c from the user cleanly
 */
void sigint_handler(int sig) {
  // Print a carraige return to overwrite the ugly-looking "^C" from the user
  if (write(STDOUT_FILENO, "\r", 1) != 1) {
    // No need to save and restore errno since the program is exiting
    perror("write");
  }
  exit(EXIT_SUCCESS);
}



/******************************************************************************
 * Main Function
 *****************************************************************************/

int main(int argc, char *argv[]) {
  REQUIRES(argc >= 1 && argv != NULL);

  // Register signal handlers
  Signal(SIGINT, sigint_handler);
  Signal(SIGPIPE, SIG_IGN);

  return 0;
}
