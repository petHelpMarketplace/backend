package config

import (
	"fmt"
	"os"

	"pethelp-backend/pkg/utils"
)

type GoogleOAuthConfig struct {
	ClientID          string
	ClientSecret      string
	ClientCallbackURL string
	SessionSecret     string
}

func LoadOAuthConf() (*GoogleOAuthConfig, error) {

	operationName := "load_google_oauth_config"
	cfg := &GoogleOAuthConfig{}

	// generate random secret key for gorilla package session init
	secret, err := utils.GenerateRandomKey(32)
	if err != nil {
		return cfg, fmt.Errorf("%s: failed to generate session secret: %w", operationName, err)
	}

	// Check if SESSION_SECRET is already set and not empty
	val, ok := os.LookupEnv("SESSION_SECRET")
	if !ok || val == "" {
		err = os.Setenv("SESSION_SECRET", secret)
		if err != nil {
			return cfg, fmt.Errorf("%s: failed to set session secret env: %w", operationName, err)
		}
	}

	cfg = &GoogleOAuthConfig{
		ClientID:          os.Getenv("CLIENT_ID"),
		ClientSecret:      os.Getenv("CLIENT_SECRET"),
		ClientCallbackURL: os.Getenv("CLIENT_CALLBACK_URL"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.ClientCallbackURL == "" || cfg.SessionSecret == "" {
		return cfg, fmt.Errorf("%s: some env variables are missing", operationName)
	}

	return cfg, nil
}
