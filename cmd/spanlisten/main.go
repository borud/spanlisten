package main

import (
	"flag"
	"log"

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
}
