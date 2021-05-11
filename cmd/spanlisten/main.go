package main

import (
	"container/ring"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/borud/spanlisten/pkg/spanlistener"
)

const (
	// length of ringbuffer
	bufferLen = 5
)

var (
	token        = flag.String("token", "", "API token for Span")
	collectionID = flag.String("collection-id", "", "Collection ID")

	// Ringbuffer
	ringBuffer = ring.New(bufferLen)
)

func ringBufferHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	fmt.Fprintf(w, "<h1>Last %d values</h1>", ringBuffer.Len())
	fmt.Fprint(w, "<ol>\n")

	// Iterate over the ringbuffer in the forward direction and call
	// callback for each entry
	ringBuffer.Do(func(p interface{}) {
		if p != nil {
			fmt.Fprintf(w, "<li>%s</li>", p)
		}
	})

	fmt.Fprint(w, "</ol>\n")
}

func main() {
	flag.Parse()

	if *token == "" || *collectionID == "" {
		log.Fatalf("Please provide me with both -token and -collection-id")
	}

	spanListener := spanlistener.New(*token, *collectionID)
	err := spanListener.Start()
	if err != nil {
		log.Fatalf("Unable to start SpanListener: %v", err)
	}

	// Loop over measurements channel.
	// spanlistener.Measurements() returns a channel of apipb.CarrierModuleMeasurements.
	// when the other end of the channel closes, the for-loop will terminate.  This is
	// a good pattern for how to use channels
	go func() {
		for m := range spanListener.Measurements() {
			log.Printf("measurement: %+v", m)

			// Assign the value and then set the ringBuffer pointer to
			// the previous entry in the ring – effectively stepping
			// backwards
			ringBuffer.Value = m
			ringBuffer = ringBuffer.Prev()
		}
		log.Printf("exits range loop over Measurements")
	}()

	// Set up webserver
	mux := http.NewServeMux()
	mux.HandleFunc("/", ringBufferHandler)

	server := http.Server{
		Addr:    ":9091",
		Handler: mux,
	}

	// Handle Ctrl-C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		server.Shutdown(context.Background())
		spanListener.Shutdown()
		os.Exit(0)
	}()

	// Start the webserver.  This call blocks until the webserver finishes.
	server.ListenAndServe()
}
