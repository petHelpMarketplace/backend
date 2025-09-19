package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type S3Config struct {
	AccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	Region          string `envconfig:"AWS_REGION" required:"true"`
	BucketName      string `envconfig:"AWS_S3_BUCKET" required:"true"`
	EndpointURL     string `envconfig:"AWS_S3_ENDPOINT_URL" required:"true"`
}

func LoadS3Config() (S3Config, error) {
	var cfg S3Config
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load S3 config: %w", err)
	}

	return cfg, nil
}
