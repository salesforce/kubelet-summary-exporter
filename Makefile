.PHONY: all install test lint 

export GO111MODULE=on

all: install test lint

build:
	go build ./...

install:
	go install -v ./cmd/...

test:
	go test -coverprofile=coverage.out -v ./...

lint: bin/golangci-lint-1.47.2
	./bin/golangci-lint-1.47.2 run ./...

bin/golangci-lint-1.47.2:
	./hack/fetch-golangci-lint.sh
