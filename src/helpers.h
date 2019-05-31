/*
 * helpers.h
 *
 * Header file containing forward declarations of miscellaneous helper
 * functions. Additionally contains wrapper functions that handle errors
 * returned by system calls. These wrappers begin with a capital letter.
 *
 * Created by Jacob Strieb
 */

#include <stdio.h>
#include <signal.h>

// Used for opening a listening socket
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>

// Include guard so that this can only be included once
#ifndef HELPERS_H

#define HELPERS_H

#define BACKLOG 128
#define MAXLINE 1024

// Type definitions
typedef void (*sighandler_t)(int);

// Helper functions
void fatal_error(const char *fmt, ...);
void print(const char* fmt, ...);
void dbg_print(const char* fmt, ...);

int open_listenfd(const char *port);

// Wrapper functions
void Signal(int signum, sighandler_t handler);
void Getaddrinfo(const char *node, const char *service, const struct addrinfo
    *hints, struct addrinfo **res);
void Setsockopt(int sockfd, int level, int optname, const void *optval,
    socklen_t optlen);
void Close(int fd);
void Listen(int fd, int backlog);
int Accept(int sockfd);

#endif /* HELPERS_H */
