GO := go
GO_PATH := $(shell $(GO) env GOPATH)
GO_BIN := $(GO_PATH)/bin

GO_MOD := $(GO) mod
GO_GET := $(GO) get -u -v
GO_FMT := $(GO) fmt
GO_RUN := $(GO) run
GO_TEST:= $(GO) test -p 1 -v -failfast
GO_LINT := golangci-lint run
# BUG: go vet: structtag field repeats json warning with valid override #40102
# https://github.com/golang/go/issues/40102
GO_VET:= $(GO) vet -v -structtag=false

#$(GO_LINT):
#	$(GO_GET) golang.org/x/lint/golint
# brew install golangci-lint

deps:
	$(GO_MOD) tidy
	$(GO_MOD) download

fmt:
	$(GO_FMT) ./...

lint:
	$(GO_LINT) ./...

vet:
	$(GO_VET) ./...

# disable lint for now
pr: vet

test:
	$(GO_TEST) $(TEST_FLAGS) ./...

# build any example with make <name>
EXAMPLE_NAME := $(word 1, $(MAKECMDGOALS))
$(EXAMPLE_NAME):
ifneq ($(filter examples/$(EXAMPLE_NAME),$(wildcard examples/*)),)
	$(GO_RUN) examples/$(EXAMPLE_NAME)/main.go
endif

all: pr test

.DEFAULT_GOAL := all

.PHONY: deps fmt lint vet keystore
