package spanlistener

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/borud/spanlisten/pkg/apipb"
	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
	"google.golang.org/protobuf/proto"
)

// SpanListener listens to a given collection on Span
type SpanListener struct {
	Token            string
	CollectionID     string
	measurementCh    chan *apipb.CarrierModuleMeasurements
	ctx              context.Context
	done             context.CancelFunc
	shutdownComplete sync.WaitGroup
}

// New creates a new SpanListener instance
func New(token string, collectionID string) *SpanListener {
	return &SpanListener{
		Token:         token,
		CollectionID:  collectionID,
		measurementCh: make(chan *apipb.CarrierModuleMeasurements),
	}
}

// Start fires up the Spanlistener
func (s *SpanListener) Start() error {
	config := spanapi.NewConfiguration()
	config.Debug = true

	s.ctx, s.done = apitools.ContextWithAuth(s.Token, 1*time.Hour)
	ds, err := apitools.NewCollectionDataStream(s.ctx, config, s.CollectionID)
	if err != nil {
		return fmt.Errorf("unable to open CollectionDataStream: %v", err)
	}

	// Start goroutine running readDataStream() function
	go s.readDataStream(ds)

	return nil
}

// Shutdown the listener
func (s *SpanListener) Shutdown() {
	if s.done != nil {
		s.done()
		s.shutdownComplete.Wait()
	}
}

// Measurements returns a chan apipb.CarrierModuleMeasurements
func (s *SpanListener) Measurements() <-chan *apipb.CarrierModuleMeasurements {
	return s.measurementCh
}

func (s *SpanListener) readDataStream(ds apitools.DataStream) {
	defer ds.Close()

	// Signal that we have started
	s.shutdownComplete.Add(1)

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

		s.measurementCh <- &pb

		if s.ctx.Err() == context.Canceled {
			log.Printf("shutting down spanlistener")
			close(s.measurementCh)
			s.shutdownComplete.Done()
			return
		}
	}
}
