# Created by Jacob Strieb


# Variables
CC = gcc
CFLAGS = -std=gnu99 -pedantic -Wall -Werror -O2
VPATH = src


# Main target
quickserv: helpers.c quickserv.c


# Compile debug target with DEBUG flag enabled
.PHONY: debug
debug: CFLAGS += -DDEBUG -g
debug: clean quickserv


# Clean up generated binary
.PHONY: clean
clean:
	rm -f quickserv


# Install quickserv
.PHONY: install
install: quickserv
	cp quickserv /usr/local/bin/quickserv


# Uninstall (remove) quickserv
.PHONY: uninstall
uninstall:
	rm -f /usr/local/bin/quickserv

.PHONY: remove
remove: uninstall
