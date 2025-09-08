package services

import (
	"errors"
	"net/http"
	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/ports"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type cookieManager struct {
	store       sessions.Store
	cookieName  string
	defaultOpts sessions.Options
}

var _ ports.CookieManager = (*cookieManager)(nil)

// NewCookieManager creates and configures a new CookieManager instance for Gin.
// It initializes a cookie store based on the provided configuration.
func NewCookieManager(cfg config.CookieConfig) (*cookieManager, error) {
	if cfg.CookieAuthKey == "" || cfg.CookieEncKey == "" {
		return nil, errors.New("cookie authentication or encryption key cannot be empty")
	}

	// Create a new cookie-based session store.
	store := cookie.NewStore([]byte(cfg.CookieAuthKey), []byte(cfg.CookieEncKey))

	opt := sessions.Options{
		Path:     cfg.CookiePath,
		Domain:   cfg.CookieDomain,
		MaxAge:   cfg.CookieMaxAge,
		HttpOnly: cfg.CookieHTTPOnly,
		Secure:   cfg.CookieSecure,
		SameSite: parseSameSite(cfg.CookieSameSite),
	}
	// Configure the default options for the session cookie.
	store.Options(opt)

	cm := &cookieManager{
		store:       store,
		defaultOpts: opt,
		cookieName:  cfg.CookieName,
	}

	return cm, nil
}

// Middleware returns the gin.HandlerFunc necessary to enable session management.
// This middleware must be added to the Gin router for the session manager to work.
func (cm *cookieManager) Middleware() gin.HandlerFunc {
	return sessions.Sessions(cm.cookieName, cm.store)
}

// Set saves a key-value pair to the user's session.
// It uses the session stored in the gin.Context by the middleware.
func (cm *cookieManager) Set(c *gin.Context, key string, value interface{}) {
	session := sessions.Default(c)
	session.Set(key, value)
}

// Get retrieves a value from the session using its key.
func (cm *cookieManager) Get(c *gin.Context, key string) (interface{}, error) {
	session := sessions.Default(c)

	value := session.Get(key)
	if value == nil {
		return nil, errors.New("session value not found for key: " + key)
	}

	return value, nil
}

// Clear invalidates the session cookie, effectively logging the user out.
// It clears all data and sets the cookie's MaxAge to -1 to delete it.
func (cm *cookieManager) Clear(c *gin.Context) error {
	session := sessions.Default(c)

	// Clear all values from the session map.
	session.Clear()

	// Set the MaxAge to -1 to instruct the browser to delete the cookie.
	session.Options(sessions.Options{
		MaxAge: -1,
	})

	// Save the session to apply the clearing and expiration.
	return session.Save()
}

// UpdateOptions updates the Options an existing session cookie.
// This is useful for extending a session's lifetime (e.g., "keep me logged in").
func (cm *cookieManager) UpdateOptions(c *gin.Context) {
	session := sessions.Default(c)

	// Update default options
	session.Options(cm.defaultOpts)
}

// NewSession sets multiple key-value pairs and saves the session.
// This is useful for creating a new session with multiple values at once, for example, after a login.
func (cm *cookieManager) BulkSet(c *gin.Context, values map[string]interface{}) {
	session := sessions.Default(c)
	for key, value := range values {
		session.Set(key, value)
	}
}

// Save saves the current session to the response. This must be called after
// making any changes to the session (e.g., Set, Clear, UpdateOptions).
func (cm *cookieManager) Save(c *gin.Context) error {
	session := sessions.Default(c)
	return session.Save()
}

// parseSameSite is a helper function to convert the string representation
// of SameSite from the config to the corresponding http.SameSite type.
func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		// Default to Strict for better security if the config value is invalid.
		return http.SameSiteStrictMode
	}
}
