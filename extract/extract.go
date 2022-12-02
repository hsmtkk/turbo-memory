package extract

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("extract", extract)
}

func extract(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
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
