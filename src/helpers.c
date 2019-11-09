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

// Used for blocking signals before reading to prevent interruptions
#include <signal.h>

// Use contracts if debugging is enabled
#include "contracts.h"

// Wrappers around syscalls to handle failures automatically
#include "wrappers.h"

// Forward declarations
#include "helpers.h"


/******************************************************************************
 * Global Variables and Type Definitions
 *****************************************************************************/

extern int quiet;
extern int verbose;



/******************************************************************************
 * Internal Helper Functions
 *****************************************************************************/

/*
 * Callback used by ftw to print statically-served files
 */
int _static_callback(const char *fpath, const struct stat *sb, int typeflag) {
  REQUIRES(fpath != NULL && sb != NULL);

  // See "man 7 inode" to read about permissions macros
  if (typeflag == FTW_F && !(sb->st_mode & S_IXUSR)) {
    print("%s\n", fpath);
  }

  return 0;
}


/*
 * Callback used by ftw to print dynamically-served files that will be executed
 * when accessed
 */
int _dynamic_callback(const char *fpath, const struct stat *sb, int typeflag) {
  REQUIRES(fpath != NULL && sb != NULL);

  // See "man 7 inode" to read about permissions macros
  if (typeflag == FTW_F && sb->st_mode & S_IXUSR) {
    print("%s\n", fpath);
  }

  return 0;
}



/******************************************************************************
 * Public Helper Functions
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
 * Read bytes into a buffer of size MAXLINE until it reaches capacity, or
 * until it finds a newline character. Allocate a char* of the correct size,
 * copy the buffer into it, and return it to the user.
 *
 * The user is responsible for freeing the response.
 *
 * TODO: Check that I did the signal blocking and unblocking properly
 */
char *read_line(int fd) {
  REQUIRES(fd > 0);

  // Block all signals so the read can't be interrupted
  sigset_t oldset, newset;
  Sigfillset(&newset);
  Sigprocmask(SIG_SETMASK, &newset, &oldset);

  char buf[MAXLINE];
  char *next_char = (char *) &buf;
  size_t size = 0;
  char *result = NULL;

  // Read one character at a time until we reach '\n' or the buffer is full
  while (size < MAXLINE) {
    ssize_t bytes_read = read(fd, next_char, 1);
    ASSERT(bytes_read <= 1);

    // Successfully read a character
    if (bytes_read == 1) {
      size++;
      if (*(next_char++) == '\n') {
        break;
      }
    }

    // EOF
    else if (bytes_read == 0) {
      size = 1;
      break;
    }

    // Error reading a character
    // TODO: Check if should actually kill the program for error reading
    else if (bytes_read < 0) {
      fatal_error("Failed to read from file descriptor %d\n", fd);
    }
  }

  // Allocate a return value and copy the buffer into it
  result = (char *) memcpy(Calloc(size, sizeof(char)), &buf, size - 1);

  // Unblock signals, restore the sigset from the beginning
  Sigprocmask(SIG_SETMASK, &oldset, NULL);

  ENSURES(result != NULL);
  return result;
}



/******************************************************************************
 * Main Procedure Helper Functions
 *****************************************************************************/

/*
 * List files being served and note whether they will be served statically or
 * executed dynamically.
 */
void print_files(void) {
  print("Files that will be accessible while the server is running:\n");
  if (ftw(".", _static_callback, MAX_FILE_NUM) < 0) {
    perror("ftw");
    fatal_error("Failed to list files that will be statically served.\n");
  }
  print("\n");

  print("(Note: if you expect to see files below that aren't there, use the\n"
      "following to make the file executable: 'chmod +x filename')\n");
  print("Files that will be run when accessed from the server:\n");
  if (ftw(".", _dynamic_callback, MAX_FILE_NUM) < 0) {
    perror("ftw");
    fatal_error("Failed to list files that will be dynamically served.\n");
  }
  print("\n");
}


/*
 * Open a socket listening on the input port number. Returns the socket
 * descriptor of the listening socket.
 *
 * This code is aggressively commented for my own benefit. I know that when I
 * look back at it after any more than a month, I'll have forgotten the details
 * of opening listening sockets. If you are me from the future reading this,
 * you're welcome.
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


/*
 * Parse the request line into a requestline_t struct.
 *
 * Returned struct will need to be freed by the user.
 */
requestline_t *parse_requestline(int connfd) {
  REQUIRES(connfd > 0);

  // Allocate a request line struct
  requestline_t *line = (requestline_t *) Calloc(1, sizeof(requestline_t));
  ASSERT(line != NULL);

  // Read the request line from the connection
  char *raw_line = read_line(connfd);

  // Parse the request line on the stack
  char method[MAXLINE], target[MAXLINE], version[MAXLINE];
  if (sscanf(raw_line, "%s %s %s", (char *) &method, (char *) &target,
        (char *) &version) < 3) {
    print("Parsing request line failed\n");
    return NULL;
  }

  // If successfully parsed, store the parts on the heap in the struct
  line->method = Strndup(method, MAXLINE);
  line->target = Strndup(target, MAXLINE);
  line->version = Strndup(version, MAXLINE);
  dbg_print("Received %s request for %s with HTTP version %s\n", line->method,
      line->target, line->version);

  // Free the now-parsed raw request line text
  free(raw_line);

  return line;
}


/*
 * Allow the client of the helper function to free the request line properly.
 */
void free_requestline(requestline_t *line) {
  if (line == NULL) return;

  // Free struct fields
  free(line->method);
  free(line->target);
  free(line->version);

  // Free struct
  free(line);
}
