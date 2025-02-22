package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/nfnt/resize"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type GCSObject struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func processImage(ctx context.Context, bucketName, fileName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	srcBucket := client.Bucket(bucketName)
	srcObj := srcBucket.Object(fileName)
	r, err := srcObj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	defer r.Close()

	img, format, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resizedImg := resize.Resize(300, 0, img, resize.Lanczos3)

	destinationBucket := client.Bucket(os.Getenv("DESTINATION_BUCKET"))
	processedObj := destinationBucket.Object(fileName)
	w := processedObj.NewWriter(ctx)
	switch format {
	case "jpeg":
		jpeg.Encode(w, resizedImg, nil)
	case "png":
		png.Encode(w, resizedImg)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to upload processed image: %w", err)
	}

	log.Printf("Processed image uploaded: %s", fileName)
	return nil
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Service is running!")
}

func startHTTPServer() {
	http.HandleFunc("/", messageHandler)
	log.Println("Starting HTTP server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startPubSubListener(ctx context.Context, projectID, subscriptionID string) {
	log.Printf("Starting Pub/Sub listener... Project: %s, Subscription: %s", projectID, subscriptionID)

	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubsubClient.Close()

	sub := pubsubClient.Subscription(subscriptionID)
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var obj GCSObject
		if err := json.Unmarshal(msg.Data, &obj); err != nil {
			log.Printf("Failed to parse message: %v", err)
			msg.Nack()
			return
		}

		if err := processImage(ctx, obj.Bucket, obj.Name); err != nil {
			log.Printf("Error processing image: %v", err)
			msg.Nack()
			return
		}
		
		msg.Ack()
	})

	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}
}

func main() {
	ctx := context.Background()

	projectID := os.Getenv("PROJECT_ID")
	subscriptionID := os.Getenv("SUBSCRIPTION_NAME")

	if projectID == "" || subscriptionID == "" {
		log.Fatalf("PROJECT_ID or SUBSCRIPTION_NAME is not set!")
	}

	// Start HTTP Server in a Goroutine
	go startHTTPServer()

	// Start Pub/Sub Listener (blocks forever)
	startPubSubListener(ctx, projectID, subscriptionID)
}
