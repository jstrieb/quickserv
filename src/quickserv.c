/*
 * quickserv.c
 *
 * Main file
 *
 * Created by Jacob Strieb
 */


#include <stdlib.h>
#include <stdio.h>

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
 * Main Function
 *****************************************************************************/

int main(int argc, char *argv[]) {
  REQUIRES(argc >= 1 && argv != NULL);

  return 0;
}
