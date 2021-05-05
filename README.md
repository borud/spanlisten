# Spanlisten

This project is a template project for a workshop on how to integrate with Span.  In this case for Autronica.

## Setup

Create directory and module

    go mod init github.com/borud/spanlisten
    mkdir -p cmd/spanlisten

## Makefile

```Makefile
all: test lint vet build

build: spanlisten

spanlisten:
	@go mod tidy
	@cd cmd/spanlisten && go build -o ../../bin/spanlisten

lint:
	@revive ./...

vet:
	@go vet ./...

test:
	@go test ./...

```

Add `README.md` and `.gitignore` and perform first commit.