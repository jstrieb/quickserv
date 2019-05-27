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

/*
 * Developer's note:
 *
 * For now, this server only runs on port 42069 to discourage production use.
 * Since I am a highly amateur software developer and this was not written with
 * security, performance, or robustness in-mind, I don't want it to be run on
 * other ports unless the user really knows what they're doing.
 *
 * This deliberate design choice also serves to discourage users from trying to
 * run QuickServ on privileged ports as root. Trying to use privileged
 * ports (i.e., < 1024) would result in 'permission denied' errors and, given
 * that the target audience is inexperienced developers, a natural reaction
 * might be to just run it as root. Doing this would make the errors go away,
 * but would make any users extremely vulnerable to attack. I see no need to
 * exacerbate the risk of running what is already (probably) a vulnerable
 * program.
 */
char *port = "42069";



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

  // Open a socket listening on port 42069
  int listenfd = open_listenfd(port);
  (void)listenfd;

  return 0;
}