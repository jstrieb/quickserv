# Created by Jacob Strieb


# Variables
CC = gcc
CFLAGS = -std=gnu99 -Wall -Werror -O2
VPATH = src


# Main target
quickserv: quickserv.c


# Compile debug target with DEBUG flag enabled
.PHONY: debug
debug: CFLAGS += -DDEBUG -g
debug: clean quickserv


# Clean up generated binary
.PHONY: clean
clean:
	rm -rf quickserv
