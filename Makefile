SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

GOCMD=go
GOBUILD=$(GOCMD) build
BINARY=simulador
BINARY_PATH=./cmd/simulador/cli.go

.DEFAULT_GOAL: $(BINARY)

all: clean build

build: 
	$(GOBUILD) -o ${BINARY} $(BINARY_PATH)

.PHONY: clean
clean: 
	rm -f $(BINARY)
