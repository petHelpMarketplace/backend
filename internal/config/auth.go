package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type AuthConfig struct {
	JWTSecret      []byte        `envconfig:"JWT_SECRET"          required:"true"`
	DefaultTimeout time.Duration `envconfig:"AUTH_TIMEOUT"        default:"5s"`
	AccessTTL      time.Duration `envconfig:"ACCESS_TOKEN_TTL"    default:"15m"`
	RefreshTTL     time.Duration `envconfig:"REFRESH_TOKEN_TTL"   default:"168h"`
}

func LoadAuthConfig() (AuthConfig, error) {
	var cfg AuthConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("loading auth config: %w", err)
	}
	if l := len(cfg.JWTSecret); l < 32 {
		return cfg, fmt.Errorf("JWT_SECRET too short (%d bytes, need ≥ 32)", l)
	}
	return cfg, nil
}
