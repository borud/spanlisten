package main

import (
	"encoding/base64"
	"flag"
	"log"
	"time"

	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
	"google.golang.org/protobuf/proto"

	"github.com/borud/spanlisten/pkg/apipb"
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

	config := spanapi.NewConfiguration()
	config.Debug = true

	ctx, _ := apitools.ContextWithAuth(*token, 1*time.Hour)

	ds, err := apitools.NewCollectionDataStream(ctx, config, *collectionID)
	if err != nil {
		log.Fatalf("Unable to open CollectionDataStream: %v", err)
	}

	readDataStream(ds)
}

func readDataStream(ds apitools.DataStream) {
	defer ds.Close()

	log.Printf("Connected to Span")
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

		// decode bytePayload as protobuffer
		var pb apipb.CarrierModuleMeasurements
		err = proto.Unmarshal(bytePayload, &pb)
		if err != nil {
			log.Fatalf("Unable to unmarshal protobuffer: %v", err)
		}
		log.Printf("protobuffer %+v", &pb)
	}
}
