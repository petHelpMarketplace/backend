package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

// CookieConfig holds the configuration for cookie management.
type CookieConfig struct {
	// CookieSecretKey is used to authenticate and encrypt the cookie.
	CookieSecretKey string `env:"COOKIE_SECRET_KEY"`

	// CookieName is the name of the cookie stored in the browser.
	CookieName string `yaml:"name" default:"pet_cookie"`

	// CookieMaxAge is the lifetime of the cookie in seconds.
	CookieMaxAge int `yaml:"max_age" default:"600"`

	// CookieDomain specifies the domain for which the cookie is valid.
	CookieDomain string `yaml:"domain" default:"localhost"`

	// CookiePath specifies the URL path for which the cookie is valid.
	CookiePath string `yaml:"path" default:"/"`

	// CookieHTTPOnly prevents JavaScript from accessing the cookie.
	CookieHTTPOnly bool `yaml:"http_only" default:"false"`

	// CookieSecure ensures the cookie is only sent over HTTPS.
	CookieSecure bool `yaml:"secure" default:"true"`

	// CookieSameSite controls cross-site request forgery protections.
	// Valid values: "strict", "lax", "none".
	CookieSameSite string `yaml:"same_site" default:"strict"`
}

// LoadCookieConfig reads configuration from a file (e.g., config.yml)
// and populates a Config struct. It also reads environment
// variables, which will override values from the config file.
func LoadCookieConfig(confPath string, logger *zap.Logger) (CookieConfig, error) {

	var cfg struct {
		CookieConfig `yaml:"cookie_config"`
	}

	// ReadConfig will read the file and then override with any existing
	// environment variables that match the `env` tags.
	err := cleanenv.ReadConfig(confPath, &cfg)
	if err != nil {
		logger.Fatal("failed to read cookie config file", zap.Error(err))
		return cfg.CookieConfig, err
	}
	return cfg.CookieConfig, nil
}
