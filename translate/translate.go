package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"golang.org/x/text/language"

	"cloud.google.com/go/pubsub"
	googletranslate "cloud.google.com/go/translate"
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

type transMessage struct {
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
	log.Println("decoded event")
	log.Printf("%v\n", string(msg.Message.Data))

	var extractMsg transMessage
	if err := json.Unmarshal(msg.Message.Data, &extractMsg); err != nil {
		return fmt.Errorf("json.Unmarshal failed; %w", err)
	}
	log.Println("decoded message")
	log.Printf("%v\n", extractMsg)

	translated, err := callTranslate(ctx, extractMsg.Text, extractMsg.Lang)
	if err != nil {
		return err
	}
	log.Println("translated")
	log.Printf("%v\n", translated)

	saveMsg := transMessage{
		Text:     translated,
		FileName: extractMsg.FileName,
		Lang:     extractMsg.Lang,
	}
	msgBytes, err := json.Marshal(saveMsg)
	if err != nil {
		return fmt.Errorf("json.Marshal failed; %w", err)
	}

	projectID := os.Getenv("GCP_PROJECT")
	topicName := os.Getenv("RESULT_TOPIC")

	if err := publishData(ctx, projectID, topicName, msgBytes); err != nil {
		return err
	}
	log.Println("published")
	log.Printf("%v\n", string(msgBytes))

	return nil
}

func callTranslate(ctx context.Context, text, lang string) (string, error) {
	langTag, err := languageTag(lang)
	if err != nil {
		return "", err
	}
	transClient, err := googletranslate.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("translate.NewClient failed; %w", err)
	}
	translated, err := transClient.Translate(ctx, []string{text}, langTag, &googletranslate.Options{Format: googletranslate.Text})
	if err != nil {
		return "", fmt.Errorf("translate.Client.DetectLanguage failed; %w", err)
	}
	return translated[0].Text, nil
}

func languageTag(lang string) (language.Tag, error) {
	switch lang {
	case "es":
		return language.EuropeanSpanish, nil
	case "en":
		return language.English, nil
	case "fr":
		return language.French, nil
	case "ja":
		return language.Japanese, nil
	}
	return language.English, fmt.Errorf("non supported language %s", lang)
}

func publishData(ctx context.Context, projectID, topicName string, msgBytes []byte) error {
	clt, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient failed; %w", err)
	}
	defer clt.Close()
	topic := clt.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msgBytes,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub.PublishResult.Get failed; %w", err)
	}
	log.Printf("published: %s\n", id)
	return nil
}

/*
exports.translateText = async event => {
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

  console.log(`Translating text into ${lang}`);
  const [translation] = await translate.translate(text, lang);

  console.log('Translated text:', translation);

  const messageData = {
    text: translation,
    filename: filename,
    lang: lang,
  };

  await publishResult(process.env.RESULT_TOPIC, messageData);
  console.log(`Text translated to ${lang}`);
};
*/
