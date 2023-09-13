.PHONY: all install test lint 

export GO111MODULE=on

all: install test lint

build:
	go build ./...

install:
	go install -v ./cmd/...

test:
	go test -coverprofile=coverage.out -v ./...

lint: 
	golangci-lint run -v "./..."
