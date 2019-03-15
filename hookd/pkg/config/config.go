package config

import (
	"github.com/navikt/deployment/common/pkg/kafka"
	"os"
)

type S3 struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	BucketName     string
	BucketLocation string
	UseTLS         bool
}

type Config struct {
	ListenAddress string
	LogFormat     string
	LogLevel      string
	WebhookURL    string
	WebhookSecret string
	ApplicationID int
	InstallID     int
	KeyFile       string
	Kafka         kafka.Config
	S3            S3
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddress: ":8080",
		LogFormat:     "text",
		LogLevel:      "debug",
		WebhookURL:    "https://hookd/events",
		WebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		ApplicationID: 0,
		InstallID:     0,
		KeyFile:       "private-key.pem",
		Kafka:         kafka.DefaultConfig(),
		S3: S3{
			Endpoint:       "localhost:9000",
			AccessKey:      os.Getenv("S3_ACCESS_KEY"),
			SecretKey:      os.Getenv("S3_SECRET_KEY"),
			BucketName:     "deployments.nais.io",
			BucketLocation: "",
			UseTLS:         true,
		},
	}
}