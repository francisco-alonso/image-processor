package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/francisco-alonso/image-processor/internal/config"
	image "github.com/francisco-alonso/image-processor/internal/services"
	"github.com/francisco-alonso/image-processor/internal/storage"

	"cloud.google.com/go/pubsub"
)

type GCSObject struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func StartListener(ctx context.Context, cfg config.Config) {
	log.Printf("Starting Pub/Sub listener... Project: %s, Subscription: %s", cfg.ProjectID, cfg.SubscriptionID)

	pubsubClient, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatal("Failed to create Pub/Sub client:", err)
	}
	defer pubsubClient.Close()

	sub := pubsubClient.Subscription(cfg.SubscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 1
	sub.ReceiveSettings.MaxExtension = 0

	for {
		err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			var obj GCSObject
			if err := json.Unmarshal(msg.Data, &obj); err != nil {
				log.Println("Failed to parse message:", err)
				msg.Nack()
				return
			}

			imgData, format, err := storage.DownloadImage(ctx, obj.Bucket, obj.Name)
			if err != nil {
				log.Println("Error downloading image:", err)
				msg.Nack()
				return
			}

			processedImg, err := image.ProcessImage(imgData, format)
			if err != nil {
				log.Println("Error processing image:", err)
				msg.Nack()
				return
			}

			err = storage.UploadImage(ctx, cfg.DestBucket, obj.Name, processedImg, format)
			if err != nil {
				log.Println("Error uploading image:", err)
				msg.Nack()
				return
			}

			msg.Ack()
		})

		if err != nil {
			log.Println("Error receiving messages:", err)
		}

		log.Println("No messages received, sleeping...")
		time.Sleep(5 * time.Second) // Reduce noise in logs
	}
}
