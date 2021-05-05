package spanlistener

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/borud/spanlisten/pkg/apipb"
	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
	"google.golang.org/protobuf/proto"
)

// SpanListener listens to a given collection on Span
type SpanListener struct {
	Token        string
	CollectionID string
}

// New creates a new SpanListener instance
func New(token string, collectionID string) *SpanListener {
	return &SpanListener{
		Token:        token,
		CollectionID: collectionID,
	}
}

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
