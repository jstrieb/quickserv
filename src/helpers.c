/*
 * helpers.c
 *
 * Implementation of miscellaneous helper functions defined in helpers.h
 *
 * Created by Jacob Strieb
 */


#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>

// Used for variable arguments
#include <stdarg.h>

// Used for opening a listening socket
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>

#include "helpers.h"

// Use contracts if debugging is enabled
#include "contracts.h"


/******************************************************************************
 * Global Variables and Type Definitions
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


/*
 * Open a socket listening on the input port number. Returns the socket
 * descriptor of the listening socket.
 *
 * Based on a function of the same name lin "csapp.c" from Computer Systems, A
 * Programmer's Perspective. This book is used in 15-213 at Carnegie Mellon
 * University.
 */
int open_listenfd(const char *port) {
  REQUIRES(port != NULL);

  dbg_print("Trying to open a socket to listen on port %s...\n", port);

  int sockfd = -1;

  // Create an empty, zeroed address hints structure such that the returned
  // addresses are suitable for use by a listening server process
  struct addrinfo hints;
  memset(&hints, 0, sizeof(struct addrinfo));

  // Info about these options is in the man page for 'getaddrinfo'
  hints.ai_socktype = SOCK_STREAM;
  hints.ai_flags = AI_PASSIVE | AI_ADDRCONFIG | AI_NUMERICSERV;

  // Get a linked list of (local) server address structures on the input port
  // to which we will try to bind
  struct addrinfo *addrinfo_list;
  Getaddrinfo(NULL, port, &hints, &addrinfo_list);

  // Traverse the list until we have bound successfully or reached the end
  struct addrinfo *cur;
  for (cur = addrinfo_list; cur != NULL; cur = cur->ai_next) {
    // Try to create a socket descriptor
    sockfd = socket(cur->ai_family, cur->ai_socktype, cur->ai_protocol);
    if (sockfd < 0) continue;
    ASSERT(sockfd >= 0);
    dbg_print("Acquired socket descriptor %d\n", sockfd);

    // Allow a restarted server to accept connections immediately
    int optval = 1;
    Setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, (const void *) &optval,
        sizeof(int));

    // Try to bind the descriptor to the address; if it works, we're done
    if (bind(sockfd, cur->ai_addr, cur->ai_addrlen) == 0) {
      dbg_print("Bound sockfd %d to an address\n", sockfd);
      break;
    }
    dbg_print("Could not bind sockfd %d\n", sockfd);

    // If binding was unsuccessful, close the socket descriptor and loop again
    Close(sockfd);
  }

  // Clean up the addrinfo struct allocated by getaddrinfo
  freeaddrinfo(addrinfo_list);

  // If we reached the end of the list, we couldn't bind
  if (cur == NULL) fatal_error("Could not bind to port %s\n", port);

  ASSERT(cur != NULL);
  ASSERT(sockfd >= 0);

  // If we're here, all went well and the OS should make this bound descriptor
  // into a listening socket to await and accept incoming connections
  Listen(sockfd, BACKLOG);

  // If we've reached here without error, return the file descriptor
  dbg_print("Listening on port %s, bound to sockfd %d\n", port, sockfd);

  ENSURES(sockfd >= 0);
  return sockfd;
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


/*
 * Try to get an address info structure list, abort if it fails
 */
/* void Getaddrinfo(const char *node, const char *service, struct addrinfo
    *hints, struct addrinfo **res) {
    */
void Getaddrinfo(const char *node, const char *service, const struct addrinfo
    *hints, struct addrinfo **res) {
  REQUIRES(service != NULL && hints != NULL);

  int err;
  if ((err = getaddrinfo(node, service, hints, res)) != 0) {
    if (err == EAI_SYSTEM) perror("getaddrinfo");
    fatal_error("Error getting address info on port %s\n%s\n", service,
        gai_strerror(err));
  }

  ENSURES(res != NULL);
}


/*
 * Set socket options
 */
void Setsockopt(int sockfd, int level, int optname, const void *optval,
    socklen_t optlen) {
  if (setsockopt(sockfd, level, optname, optval, optlen) < 0) {
    perror("setsockopt");
    fatal_error("Failed to set socket options\n");
  }
}


/*
 * Close file descriptor
 */
void Close(int fd) {
  if (close(fd) < 0) {
    perror("close");
    fatal_error("Failed to close file descriptor %d\n", fd);
  }
}

/*
 * Listen on a file descriptor as a server
 */
void Listen(int fd, int backlog) {
  if (listen(fd, backlog) < 0) {
    Close(fd);
    perror("listen");
    fatal_error("Failed to listen on the file descriptor %d\n", fd);
  }
}
