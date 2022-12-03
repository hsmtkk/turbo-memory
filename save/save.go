package save

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("save", save)
}

// https://cloud.google.com/functions/docs/tutorials/pubsub?hl=ja#preparing_the_application

type MessagePublishedData struct {
	Message PubSubMessage
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type transMessage struct {
	Text     string `json:"text"`
	FileName string `json:"filename"`
	Lang     string `json:"lang"`
}

func save(ctx context.Context, evt event.Event) error {
	log.Println("save")
	log.Printf("%v\n", evt)

	var msg MessagePublishedData
	if err := evt.DataAs(&msg); err != nil {
		return fmt.Errorf("event.Event.DataAs failed; %w", err)
	}
	log.Println("decoded")
	log.Printf("%v\n", string(msg.Message.Data))

	var transMsg transMessage
	if err := json.Unmarshal(msg.Message.Data, &transMsg); err != nil {
		return fmt.Errorf("json.Unmarshal failed; %w", err)
	}

	bucket := os.Getenv("RESULT_BUCKET")

	newFileName := renameImageForSave(transMsg.FileName, transMsg.Lang)

	clt, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient failed; %w", err)
	}
	defer clt.Close()

	obj := clt.Bucket(bucket).Object(newFileName)
	writer := obj.NewWriter(ctx)
	if _, err := writer.Write([]byte(transMsg.Text)); err != nil {
		return fmt.Errorf("storage.Writer.Write failed; %w", err)
	}
	log.Println("saved")

	return nil
}

func renameImageForSave(name, lang string) string {
	return fmt.Sprintf("%s_to_%s.txt", name, lang)
}

/*
exports.saveResult = async event => {
  const pubsubData = event.data;
  const jsonStr = Buffer.from(pubsubData, 'base64').toString();
  const {text, filename, lang} = JSON.parse(jsonStr);

  if (!text) {
    throw new Error(
      'Text not provided. Make sure you have a "text" property in your request'
    );
  }
  if (!filename) {
    throw new Error(
      'Filename not provided. Make sure you have a "filename" property in your request'
    );
  }
  if (!lang) {
    throw new Error(
      'Language not provided. Make sure you have a "lang" property in your request'
    );
  }

  console.log(`Received request to save file ${filename}`);

  const bucketName = process.env.RESULT_BUCKET;
  const newFilename = renameImageForSave(filename, lang);
  const file = storage.bucket(bucketName).file(newFilename);

  console.log(`Saving result to ${newFilename} in bucket ${bucketName}`);

  await file.save(text);
  console.log('File saved.');
};

const renameImageForSave = (filename, lang) => {
  return `${filename}_to_${lang}.txt`;
};

*/
