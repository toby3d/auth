#!/usr/bin/make -f
SHELL = /bin/sh

#### Start of system configuration section. ####

srcdir = .

GO ?= go
GOFLAGS ?= -buildvcs=true
EXECUTABLE ?= indieauth

#### End of system configuration section. ####

.PHONY: all clean check help

all: main.go
	$(GO) build -v $(GOFLAGS) -o $(EXECUTABLE)

clean: ## Delete all files in the current directory that are normally created by building the program
	-rm $(srcdir)/internal/testing/httptest/{cert,key}.pem
	$(GO) clean

check: ## Perform self-tests
	$(GO) generate $(srcdir)/internal/testing/httptest/...
	$(GO) test -v -cover -failfast -short -shuffle=on $(GOFLAGS) $(srcdir)/...

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
