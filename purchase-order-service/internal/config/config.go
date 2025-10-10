package config

import (
	"os"
)

type Config struct {
	Port               string
	AWSRegion          string
	DynamoDBEndpoint   string
	NATSURL            string
	Environment        string
	SupplierServiceURL string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8081"),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		DynamoDBEndpoint:   getEnv("DYNAMODB_ENDPOINT", ""),
		NATSURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		SupplierServiceURL: getEnv("SUPPLIER_SERVICE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
