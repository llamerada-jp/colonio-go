
# commands
SUDO ?= sudo

.PHONY: all
all: test

.PHONY: setup
setup:
	if [ $(shell uname -s) = 'Linux' ]; then \
		export DEBIAN_FRONTEND=noninteractive; \
		$(SUDO) apt update; \
		$(SUDO) apt -y install --no-install-recommends golang; \
	elif [ $(shell uname -s) = 'Darwin' ]; then \
		$(MAKE) setup-macos; \
	else \
		brew update; \
		brew install golang; \
	fi

.PHONY: test
test: bridge.a
	export COLONIO_SEED_BIN_PATH=$(PWD)/../colonio-seed/seed; \
	go test test/*.go

_obj/_cgo_export.h: colonio.go
	go tool cgo colonio.go

bridge.o: bridge.c _obj/_cgo_export.h
	gcc -c $^ -I. -I_obj

bridge.a: bridge.o
	ar cr $@ $^

.PHONY: clean
clean:
	$(RM) bridge.o bridge.a
	$(RM) -r _obj
