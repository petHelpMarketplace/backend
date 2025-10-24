// internal/handlers/helpers.go (New File)
package handlers

import (
	"fmt"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// getUserIDFromContext extracts, validates, and parses the userID from the Gin context.
// It handles errors and writes the appropriate JSON response if any step fails.
// It returns the userID and a boolean indicating success.
func getUserIDFromContext(c *gin.Context, logger *zap.Logger) (int64, bool) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		logger.Warn("userID not found in context, middleware might not have run or failed")
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized user access attempt",
		})
		return 0, false
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		logger.Error("userID in context is not a string", zap.Any("type", fmt.Sprintf("%T", userIDRaw)))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return 0, false
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID < 0 {
		logger.Error("failed to parse userID from context", zap.String("userID", userIDStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error"})
		return 0, false
	}

	return userID, true
}
