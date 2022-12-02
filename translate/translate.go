package translate

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("translate", translate)
}

func translate(ctx context.Context, evt event.Event) error {
	log.Println("translate")
	log.Printf("%v\n", evt)
	return nil
}
