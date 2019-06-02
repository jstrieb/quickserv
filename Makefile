# Created by Jacob Strieb


# Variables
CC = gcc
CFLAGS = -std=gnu99 -pedantic -Wall -Wextra -Werror -O2
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


# XXX: Undocumented command to recompile with a different port
# 		 Usage: make port -PORT=8080
.PHONY: port
port: CFLAGS += -DPORT=\"$(PORT)\"
port: clean quickserv


# Install quickserv
.PHONY: install
install: quickserv
	cp quickserv /usr/local/bin/quickserv
	@echo QuickServ successfully installed. Run using the \"quickserv\" command.


# Uninstall (remove) quickserv
.PHONY: uninstall
uninstall:
	rm -f /usr/local/bin/quickserv
	@echo QuickServ successfully uninstalled.

.PHONY: remove
remove: uninstall
