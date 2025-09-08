package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

// CookieConfig holds the configuration for cookie management.
type CookieConfig struct {
	// CookieAuthKey is used to authenticate the cookie.
	CookieAuthKey string `env:"COOKIE_AUTH_KEY"`

	// CookieEncKey is used to encrypt the cookie.
	CookieEncKey string `env:"COOKIE_ENC_KEY"`

	// CookieName is the name of the cookie stored in the browser.
	CookieName string `yaml:"name" env:"COOKIE_NAME" default:"pethelp_cookie"`

	// CookieMaxAge is the lifetime of the cookie in seconds.
	CookieMaxAge int `yaml:"max_age" env:"COOKIE_MAX_AGE" default:"600"`

	// CookieDomain specifies the domain for which the cookie is valid.
	CookieDomain string `yaml:"domain" env:"COOKIE_DOMAIN" default:"localhost"`

	// CookiePath specifies the URL path for which the cookie is valid.
	CookiePath string `yaml:"path" env:"COOKIE_PATH" default:"/"`

	// CookieHTTPOnly prevents JavaScript from accessing the cookie.
	CookieHTTPOnly bool `yaml:"http_only" env:"COOKIE_HTTP_ONLY" default:"false"`

	// CookieSecure ensures the cookie is only sent over HTTPS.
	CookieSecure bool `yaml:"secure" env:"COOKIE_SECURE" default:"true"`

	// CookieSameSite controls cross-site request forgery protections.
	// Valid values: "strict", "lax", "none".
	CookieSameSite string `yaml:"same_site" env:"COOKIE_SAME_SITE" default:"strict"`
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
		logger.Error("failed to read cookie config file", zap.Error(err))
		return cfg.CookieConfig, err
	}
	return cfg.CookieConfig, nil
}
