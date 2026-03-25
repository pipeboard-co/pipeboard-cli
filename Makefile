VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X github.com/pipeboard-co/pipeboard-cli/cmd.Version=$(VERSION) -X github.com/pipeboard-co/pipeboard-cli/cmd.Commit=$(COMMIT)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o bin/pipeboard .

install:
	go install $(LDFLAGS) .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
