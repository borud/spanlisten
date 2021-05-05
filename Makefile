all: gen test lint vet build

build: spanlisten

spanlisten:
	@cd cmd/spanlisten && go build -o ../../bin/spanlisten

lint:
	@revive ./...

vet:
	@go vet ./...

test:
	@go test ./...

gen:
	@buf generate