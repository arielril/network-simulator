SOURCEDIR=.

GOCMD=go
GOBUILD=$(GOCMD) build
BINARY=simulador
BINARY_PATH=$(SOURCEDIR)/cmd/simulador/cli.go

.DEFAULT_GOAL: $(BINARY)

all: clean build

build: 
	$(GOBUILD) -o ${BINARY} $(BINARY_PATH)

clean: 
	rm -f $(BINARY)
