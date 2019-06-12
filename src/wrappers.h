/*
 * wrappers.h
 *
 * Wrapper functions around system calls that automatically handle failures
 *
 * Created by Jacob Strieb
 */


// Used for handling signals
#include <signal.h>

// Used for opening a listening socket
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>


// Include guard so that this can only be included once
#ifndef _WRAPPERS_H

#define _WRAPPERS_H

// Macros
#define MAXLINE 1024

// Type definitions
typedef void (*sighandler_t)(int);

// Wrapper functions
void Signal(int signum, sighandler_t handler);
void Getaddrinfo(const char *node, const char *service, const struct addrinfo
    *hints, struct addrinfo **res);
void Setsockopt(int sockfd, int level, int optname, const void *optval,
    socklen_t optlen);
void Close(int fd);
void Listen(int fd, int backlog);
int Accept(int sockfd);

#endif /* _WRAPPERS_H */
