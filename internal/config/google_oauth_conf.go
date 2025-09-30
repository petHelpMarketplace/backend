package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type GoogleOAuthConfig struct {
	ClientID          string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret      string `envconfig:"CLIENT_SECRET" required:"true"`
	ClientCallbackURL string `envconfig:"CLIENT_CALLBACK_URL" required:"true"`
	SessionSecret     string `envconfig:"SESSION_SECRET" required:"true"`
}

func LoadOAuthConf() (GoogleOAuthConfig, error) {

	operationName := "load_google_oauth_config"
	cfg := GoogleOAuthConfig{}

	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("%s loading Oauth config: %w", operationName, err)
	}
	if l := len(cfg.SessionSecret); l < 32 {
		return cfg, fmt.Errorf("SESSION_SECRET too short (%d bytes, need ≥ 32)", l)
	}

	// Check if SESSION_SECRET is already set and not empty
	// val, ok := os.LookupEnv("SESSION_SECRET")
	// if !ok || val == "" {
	// 	err = os.Setenv("SESSION_SECRET", secret)
	// 	if err != nil {
	// 		return cfg, fmt.Errorf("%s: failed to set session secret env: %w", operationName, err)
	// 	}
	// }

	// cfg = GoogleOAuthConfig{
	// 	ClientID:          os.Getenv("CLIENT_ID"),
	// 	ClientSecret:      os.Getenv("CLIENT_SECRET"),
	// 	ClientCallbackURL: os.Getenv("CLIENT_CALLBACK_URL"),
	// 	SessionSecret:     os.Getenv("SESSION_SECRET"),
	// }

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.ClientCallbackURL == "" || cfg.SessionSecret == "" {
		return cfg, fmt.Errorf("%s: some env variables are missing", operationName)
	}

	return cfg, nil
}
