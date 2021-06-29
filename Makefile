
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
test:
	export COLONIO_SEED_BIN_PATH=$(PWD)/../colonio-seed/seed; \
	go test -v test/*.go

.PHONY: clean
clean:
	$(RM) -r _obj
