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

## Push first commit

- Add `README.md` and `.gitignore` 
- run `git init`
- create git repository

# Add dependency to Span

Add dependency to Span library

    "github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"

## Set up listening

```go
config := spanapi.NewConfiguration()
config.Debug = true
ctx, _ := apitools.ContextWithAuth(*token, 1*time.Hour)
ds, err := apitools.NewCollectionDataStream(ctx, config, *collectionID)

if err != nil {
	log.Fatalf("Unable to open CollectionDataStream: %v", err)
}

readDataStream(ds)
```

## Iterate over the incoming stream

```go
for {
	msg, err := ds.Recv()
	if err != nil {
		log.Fatalf("Error reading message: %v", err)
	}
	// We only care about messages containing data
	if *msg.Type != "data" {
		continue
	}

	// base64 decode the payload to a string
	bytePayload, err := base64.StdEncoding.DecodeString(*msg.Payload)
	if err != nil {
		log.Fatalf("Unable to decode payload: %v", err)
	}

	log.Printf("%s %s", *msg.Device.DeviceId, *msg.Payload)
	log.Printf("hex %x", bytePayload)
}
```
# Protobuffer

Create `buf.yaml`

```yaml
version: v1beta1
build:
  roots:
    - proto
```

and `buf.gen.yaml`

```yaml
version: v1beta1
plugins:
  - name: go
    out: pkg/apipb
    opt: paths=source_relative
```

Then run `buf generate` and observe that `pkg/apipb` is created.

Add `gen` rule to `Makefile`

Remember to run `go mod tidy`.

```Makefile

gen:
	@buf generate
```
