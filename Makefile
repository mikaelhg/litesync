GOCMD=go
GOTEST=$(GOCMD) test
BINARY_NAME=litesync

.PHONY: all test build

all: help

build:
	@mkdir -p out/bin
	$(GOCMD) build -o out/bin/$(BINARY_NAME) ./cmd/${BINARY_NAME}

clean:
	rm -fr ./out

test:
	$(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)

lint:
	golangci-lint run

pre-commit: test
	golangci-lint run -n
