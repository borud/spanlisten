all: gen test lint vet build

build: spanlisten spanfetch

spanlisten:
	@cd cmd/spanlisten && go build -o ../../bin/spanlisten

spanfetch:
	@cd cmd/spanfetch && go build -o ../../bin/spanfetch

lint:
	@revive ./...

vet:
	@go vet ./...

test:
	@go test ./...

gen:
	@buf generate