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

// Used for walking through the directory tree to show the user what will be
// available on their server
#include <ftw.h>

// Use contracts if debugging is enabled
#include "contracts.h"

// Forward declarations
#include "helpers.h"


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
 * Print a message informing the user of the current working directory so that
 * they know where their files are being served from.
 */
void print_wd(void) {
  char *wd;
  // See the note in the DESCRIPTION section of the getwd man page for when buf
  // is null and size is 0 in getcwd
  if ((wd = getcwd(NULL, 0)) == NULL) {
    perror("get_current_dir_name");
    fatal_error("Failed to get current directory name.\n");
  }
  ASSERT(wd != NULL);

  print("\nCurrently serving files from the following folder:\n%s/\n\n", wd);

  free(wd);
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
  print("Server running on port %s; it will run until this program exits\n"
    "To connect, type http://localhost:%s/ into your browser\n\n", port, port);

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
  REQUIRES(sockfd >= 0);

  if (setsockopt(sockfd, level, optname, optval, optlen) < 0) {
    perror("setsockopt");
    fatal_error("Failed to set socket options\n");
  }
}


/*
 * Close file descriptor
 */
void Close(int fd) {
  REQUIRES(fd >= 0);

  if (close(fd) < 0) {
    perror("close");
    fatal_error("Failed to close file descriptor %d\n", fd);
  }
}

/*
 * Listen on a file descriptor as a server
 */
void Listen(int fd, int backlog) {
  REQUIRES(fd >= 0);

  if (listen(fd, backlog) < 0) {
    Close(fd);
    perror("listen");
    fatal_error("Failed to listen on the file descriptor %d\n", fd);
  }
}


/*
 * Accept a client connection and return a file descriptor to read/write
 */
int Accept(int sockfd) {
  REQUIRES(sockfd >= 0);

  // Initialize data structure to hold address information
  socklen_t clientlen = sizeof(struct sockaddr_storage);
  struct sockaddr_storage client;

  // Accept a connection
  int result;
  if ((result = accept(sockfd, (struct sockaddr *) &client, &clientlen)) < 0) {
    perror("accept");
  }

  // Don't print info for invalid connections
  if (result < 0) return result;
  ASSERT(result >= 0);

  // Get and print information about the client connection
  char client_name[MAXLINE], client_port[MAXLINE];
  int err;
  if ((err = getnameinfo((struct sockaddr *) &client, clientlen, client_name,
          MAXLINE, client_port, MAXLINE, 0)) != 0) {
    print("%s\n", gai_strerror(err));
    return -1;
  }

  print("Client connected at %s:%s\n", client_name, client_port);

  ENSURES(result >= 0);
  return result;
}
