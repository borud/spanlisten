package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/borud/spanlisten/pkg/spanlistener"
)

var (
	token        = flag.String("token", "", "API token for Span")
	collectionID = flag.String("collection-id", "", "Collection ID")
)

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

	// Handle Ctrl-C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		spanListener.Shutdown()
		os.Exit(0)
	}()

	// Loop over measurements channel
	for m := range spanListener.Measurements() {
		log.Printf("measurement: %+v", m)
	}
}
