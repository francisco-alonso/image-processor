package main

import (
	"context"
	"log"
	"net/http"

	"github.com/francisco-alonso/image-processor/internal/config"
	"github.com/francisco-alonso/image-processor/pkg/pubsub"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadEnv()

	if cfg.ProjectID == "" || cfg.SubscriptionID == "" {
		log.Fatal("PROJECT_ID or SUBSCRIPTION_NAME is not set!")
	}

	// Start HTTP health check server
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Service is running!"))
		})
		log.Println("Starting HTTP server on port 8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Start Pub/Sub listener
	pubsub.StartListener(ctx, cfg)
}
