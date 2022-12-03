package translate

import (
	"context"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("translate", translate)
}

// https://cloud.google.com/functions/docs/tutorials/pubsub?hl=ja#preparing_the_application

type MessagePublishedData struct {
	Message PubSubMessage
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type message struct {
	Text     string `json:"text"`
	FileName string `json:"filename"`
	Lang     string `json:"lang"`
}

func translate(ctx context.Context, evt event.Event) error {
	log.Println("translate")
	log.Printf("%v\n", evt)

	var msg MessagePublishedData
	if err := evt.DataAs(&msg); err != nil {
		return fmt.Errorf("event.Event.DataAs failed; %w", err)
	}
	log.Println("decoded")
	log.Printf("%v\n", string(msg))

	return nil
}
