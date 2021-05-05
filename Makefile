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
