package save

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("save", save)
}

func save(ctx context.Context, evt event.Event) error {
	log.Println("save")
	log.Printf("%v\n", evt)
	return nil
}
