/*
 * wrappers.c
 *
 * Implementation of error-handling wrapper functions
 *
 * Created by Jacob Strieb
 */


#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

// Used for string duplication functions
#include <string.h>

// Used for opening a listening socket
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>

// Used for sigprocmask
#include <signal.h>

// Use contracts if debugging is enabled
#include "contracts.h"

// Forward declarations
#include "wrappers.h"

// Helper functions
#include "helpers.h"


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


/*
 * Set the current process signal block mask
 */
void Sigprocmask(int how, const sigset_t *set, sigset_t *oldset) {
  if (sigprocmask(how, set, oldset) < 0) {
    perror("sigprocmask");
    fatal_error("Failed to set signal mask\n");
  }
}


/*
 * Fill a sigset_t entirely
 */
void Sigfillset(sigset_t *set) {
  REQUIRES(set != NULL);

  if (sigfillset(set) < 0) {
    perror("sigfillset");
    fatal_error("Failed to create a filled signal set\n");
  }

  ENSURES(set != NULL);
}


/*
 * Allocate a zeroed block of memory
 */
void *Calloc(size_t n, size_t size) {
  void *result;

  if ((result = calloc(n, size)) == NULL) {
    fatal_error("Failed to allocate memory (calloc)\n");
  }

  ENSURES(result != NULL);
  return result;
}

/*
 * Duplicate a string on the heap
 */
char *Strndup(const char *s, size_t n) {
  char *result;

  if ((result = strndup(s, n)) == NULL) {
    fatal_error("Failed to allocate memory (strndup)\n");
  }

  ENSURES(result != NULL);
  return result;
}
