package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
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
	config.Debug = false

	ctx, _ := apitools.ContextWithAuth(*token, 100*time.Second)
	client := spanapi.NewAPIClient(config)

	messageChan := TraverseCollectionData(ctx, client, time.Now().Add(-1*time.Hour), time.Now())

	for m := range messageChan {
		log.Printf("ts='%s' deviceId='%s' msgId='%s' payload='%s'", *m.Received, *m.Device.DeviceId, *m.MessageId, *m.Payload)
	}

}

// TraverseCollectionData between two timestamps.  Returns a channel over which
// spanapi.OutputDataMessage instances will be streamed.
func TraverseCollectionData(ctx context.Context, client *spanapi.APIClient, start time.Time, end time.Time) <-chan spanapi.OutputDataMessage {
	// Create the output channel
	out := make(chan spanapi.OutputDataMessage)

	// Set targets
	startInt := start.UnixNano() / int64(time.Millisecond)
	startString := fmt.Sprintf("%d", startInt)
	endString := fmt.Sprintf("%d", end.UnixNano()/int64(time.Millisecond))

	// Fetch data
	go func() {
		for {
			result, _, err := client.CollectionsApi.ListCollectionData(ctx, *collectionID).
				Start(startString).
				End(endString).
				Execute()

			if err != nil {
				log.Printf("error listing collection: %v", err)
				close(out)
				return
			}

			data := result.GetData()

			// Check if result was empty
			if len(data) == 0 {
				close(out)
				return
			}

			// Loop over the data we got
			for _, msg := range data {
				msInt, err := strconv.ParseInt(*msg.Received, 10, 64)
				if err != nil {
					log.Printf("bogus timestamp in OutputDataMessage: %s", *msg.Received)
				}

				// Push message onto channel
				out <- msg

				// Terminate when we have hit the start time
				if msInt <= startInt {
					close(out)
					return
				}
			}

			// Update endpoint
			endString = *data[len(data)-1].Received
		}
	}()

	return out
}
