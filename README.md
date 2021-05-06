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

## Unmarshal the protobuffer

Add import `"google.golang.org/protobuf/proto"`

Then decode the protobuf

```go
// decode bytePayload as protobuffer
var pb apipb.CarrierModuleMeasurements
err = proto.Unmarshal(bytePayload, &pb)
if err != nil {
	log.Fatalf("Unable to unmarshal protobuffer: %v", err)
}
log.Printf("protobuffer %+v", &pb)
```

## Move to its own package

Create `pkg/spanlistener` and make a SpanListener type


```go
type SpanListener struct {
	Token        string
	CollectionID string
}
```

Add a new function with `*SpanListener` receiver

```go
// New creates a new SpanListener instance
func New(token string, collectionID string) *SpanListener {
	return &SpanListener{
		Token:        token,
		CollectionID: collectionID,
	}
}
```

Make a `Start()` function

```go
// Start fires up the Spanlistener
func (s *SpanListener) Start() error {
	config := spanapi.NewConfiguration()
	config.Debug = true

	ctx, _ := apitools.ContextWithAuth(s.Token, 1*time.Hour)
	ds, err := apitools.NewCollectionDataStream(ctx, config, s.CollectionID)
	if err != nil {
		return fmt.Errorf("unable to open CollectionDataStream: %v", err)
	}

	// Start goroutine running readDataStream() function
	go s.readDataStream(ds)

	return nil
}
````

And rewrite the readDataStream:

```go
func (s *SpanListener) readDataStream(ds apitools.DataStream) {
	defer ds.Close()

	log.Printf("connected to Span")
	for {
		msg, err := ds.Recv()
		if err != nil {
			log.Fatalf("error reading message: %v", err)
		}

		// We only care about messages containing data
		if *msg.Type != "data" {
			continue
		}

		// base64 decode the payload to a string
		bytePayload, err := base64.StdEncoding.DecodeString(*msg.Payload)
		if err != nil {
			log.Fatalf("unable to decode payload: %v", err)
		}

		// decode bytePayload as protobuffer
		var pb apipb.CarrierModuleMeasurements
		err = proto.Unmarshal(bytePayload, &pb)
		if err != nil {
			log.Fatalf("unable to unmarshal protobuffer: %v", err)
		}
		log.Printf("protobuffer %+v", &pb)
	}
}
```

## Make use of channels

Add a channel to the `SpanListener` type

```go
type SpanListener struct {
	Token         string
	CollectionID  string
	measurementCh chan *apipb.CarrierModuleMeasurements
}
```

and make sure we create a channel in `New()`:

```go
measurementCh: make(chan *apipb.CarrierModuleMeasurements),
```

Then output the `pb` to that channel, noting why we have to use a pointer

```go
s.measurementCh <- &pb
````

Then we make a function that returns a reference to the channel

```go
// Measurements returns a chan apipb.CarrierModuleMeasurements
func (s *SpanListener) Measurements() <-chan *apipb.CarrierModuleMeasurements {
	return s.measurementCh
}
```

Talk a bit about channel length and about sizing channels.

## Graceful shutdown

Introducing the context object.  First we add it to the `SpanListener` struct.

(Explain that we get to the sync.WaitGroup later.  Mention that WaitGroup are like CountdownLatch in Java)

```go
type SpanListener struct {
	Token            string
	CollectionID     string
	measurementCh    chan *apipb.CarrierModuleMeasurements
	ctx              context.Context
	done             context.CancelFunc
	shutdownComplete sync.WaitGroup
```

Then we capture it when making the context:

```go
s.ctx, s.done = apitools.ContextWithAuth(s.Token, 1*time.Hour)
ds, err := apitools.NewCollectionDataStream(s.ctx, config, s.CollectionID)
```

make sure the `shutdownComplete` is set up

```go
s.shutdownComplete.Add(1)
```

Check for context cancellation

```go
s.measurementCh <- &pb
if s.ctx.Err() == context.Canceled {
	log.Printf("shutting down spanlistener")
	close(s.measurementCh)
	s.shutdownComplete.Done()
	return
}
```

Then add a shutdown

```go
// Shutdown the listener
func (s *SpanListener) Shutdown() {
	if s.done != nil {
		s.done()
		s.shutdownComplete.Wait()
	}
}
```

In `main.go` we need some way of triggering this so we hook into the Ctrl-C handling.

```go
// Handle Ctrl-C
c := make(chan os.Signal)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
	<-c
	fmt.Println("\r- Ctrl+C pressed in Terminal")
	spanListener.Shutdown()
	os.Exit(0)
}()
```
