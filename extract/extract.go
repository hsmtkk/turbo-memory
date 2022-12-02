package extract

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("extract", extract)
}

// https://cloud.google.com/functions/docs/tutorials/storage?hl=ja#object_finalize_deploying_the_function

type StorageObjectData struct {
	Bucket         string    `json:"bucket,omitempty"`
	Name           string    `json:"name,omitempty"`
	Metageneration int64     `json:"metageneration,string,omitempty"`
	TimeCreated    time.Time `json:"timeCreated,omitempty"`
	Updated        time.Time `json:"updated,omitempty"`
}

func extract(ctx context.Context, evt event.Event) error {
	log.Println("extract")
	log.Printf("%v\n", evt)

	var data StorageObjectData
	if err := evt.DataAs(&data); err != nil {
		return fmt.Errorf("event.Event.DataAs failed; %w", err)
	}
	log.Printf("%v\n", data)

	return nil
}

/*
exports.processImage = async event => {
  const {bucket, name} = event;

  if (!bucket) {
    throw new Error(
      'Bucket not provided. Make sure you have a "bucket" property in your request'
    );
  }
  if (!name) {
    throw new Error(
      'Filename not provided. Make sure you have a "name" property in your request'
    );
  }

  await detectText(bucket, name);
  console.log(`File ${name} processed.`);
};

const detectText = async (bucketName, filename) => {
  console.log(`Looking for text in image ${filename}`);
  const [textDetections] = await vision.textDetection(
    `gs://${bucketName}/${filename}`
  );
  const [annotation] = textDetections.textAnnotations;
  const text = annotation ? annotation.description.trim() : '';
  console.log('Extracted text from image:', text);

  let [translateDetection] = await translate.detect(text);
  if (Array.isArray(translateDetection)) {
    [translateDetection] = translateDetection;
  }
  console.log(
    `Detected language "${translateDetection.language}" for ${filename}`
  );

  // Submit a message to the bus for each language we're going to translate to
  const TO_LANGS = process.env.TO_LANG.split(',');
  const topicName = process.env.TRANSLATE_TOPIC;

  const tasks = TO_LANGS.map(lang => {
    const messageData = {
      text: text,
      filename: filename,
      lang: lang,
    };

    // Helper function that publishes translation result to a Pub/Sub topic
    // For more information on publishing Pub/Sub messages, see this page:
    //   https://cloud.google.com/pubsub/docs/publisher
    return publishResult(topicName, messageData);
  });

  return Promise.all(tasks);
};
*/
