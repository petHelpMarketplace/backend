package ports

import "github.com/gin-gonic/gin"

// CookieManager provides a centralized way to manage session cookies. It abstracts
// the underlying session store and provides methods for setting, getting, and
// clearing session data.
type CookieManager interface {
	// Middleware returns a gin.HandlerFunc that must be registered with the Gin router.
	// This middleware is responsible for initializing the session on each request,
	// making it available via `sessions.Default(c)`.
	Middleware() gin.HandlerFunc

	// Set saves a key-value pair to the session and immediately commits the change
	// to the response cookie.
	Set(c *gin.Context, key string, value interface{})

	// Get retrieves a value from the session by its key.
	// It returns an error if the session is not found or the key does not exist.
	Get(c *gin.Context, key string) (interface{}, error)

	// Clear invalidates the session by clearing all its data and setting the
	// cookie's MaxAge to -1, which instructs the browser to delete it.
	// This change is saved immediately.
	Clear(c *gin.Context) error

	// BulkSet sets multiple key-value pairs in the session and saves it in a single
	// operation.
	BulkSet(c *gin.Context, values map[string]interface{})

	// UpdateOptions updates default options an existing session cookie.
	UpdateOptions(c *gin.Context)

	// Save saves the current session to the response.
	Save(c *gin.Context) error
}
