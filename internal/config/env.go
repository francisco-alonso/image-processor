package config

import "os"

type Config struct {
	ProjectID      string
	SubscriptionID string
	DestBucket     string
}

func LoadEnv() Config {
	return Config{
		ProjectID:      os.Getenv("PROJECT_ID"),
		SubscriptionID: os.Getenv("SUBSCRIPTION_NAME"),
		DestBucket:     os.Getenv("DESTINATION_BUCKET"),
	}
}