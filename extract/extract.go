package extract

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/translate"
	vision "cloud.google.com/go/vision/apiv1"
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

	if err := detectText(ctx, data.Bucket, data.Name); err != nil {
		return err
	}

	return nil
}

func detectText(ctx context.Context, bucket, name string) error {
	texts, err := callVision(ctx, bucket, name)
	if err != nil {
		return err
	}
	if err := callTranslate(ctx, texts); err != nil {
		return err
	}

	projectID := os.Getenv("GCP_PROJECT")
	toLang := os.Getenv("TO_LANG")
	topic := os.Getenv("TRANSLATE_TOPIC")
	langs := strings.Split(toLang, ",")
	text := strings.Join(texts, " ")

	for _, lang := range langs {
		attrs := map[string]string{
			"text":     text,
			"filename": name,
			"lang":     lang,
		}
		if err := publishData(ctx, projectID, topic, attrs); err != nil {
			log.Printf("publishData failed; %v", err.Error())
		}
		log.Printf("published: %v\n", attrs)
	}

	return nil
}

func callVision(ctx context.Context, bucket, name string) ([]string, error) {
	visionClt, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("vision.NewImageAnnotatorClient failed; %w", err)
	}
	url := fmt.Sprintf("gs://%s/%s", bucket, name)
	img := vision.NewImageFromURI(url)
	annotations, err := visionClt.DetectTexts(ctx, img, nil, 10)
	if err != nil {
		return nil, fmt.Errorf("vision.ImageAnnotatorClient.DetectTexts failed; %w", err)
	}
	texts := []string{}
	for _, annotation := range annotations {
		texts = append(texts, annotation.Description)
	}
	log.Printf("detected texts: %v\n", texts)
	return texts, nil
}

func callTranslate(ctx context.Context, texts []string) error {
	transClient, err := translate.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("translate.NewClient failed; %w", err)
	}
	detections, err := transClient.DetectLanguage(ctx, texts)
	if err != nil {
		return fmt.Errorf("translate.Client.DetectLanguage failed; %w", err)
	}
	langSet := map[string]bool{}
	for _, detects := range detections {
		for _, detect := range detects {
			langSet[detect.Language.String()] = true
		}
	}
	langs := []string{}
	for lang := range langSet {
		langs = append(langs, lang)
	}
	log.Printf("detected languages: %v\n", langs)
	return nil
}

func publishData(ctx context.Context, projectID, topicName string, attrs map[string]string) error {
	clt, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient failed; %w", err)
	}
	defer clt.Close()
	topic := clt.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		Data:       []byte("extract"),
		Attributes: attrs,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub.PublishResult.Get failed; %w", err)
	}
	log.Printf("published: %s\n", id)
	return nil
}

/*
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
