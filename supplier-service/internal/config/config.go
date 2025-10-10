package config

import (
	"os"
)

type Config struct {
	Port             string
	AWSRegion        string
	DynamoDBEndpoint string
	NATSURL          string
	Environment      string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		AWSRegion:        getEnv("AWS_REGION", "us-east-1"),
		DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", ""),
		NATSURL:          getEnv("NATS_URL", "nats://localhost:4222"),
		Environment:      getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
