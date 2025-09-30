package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
)

type AuthMiddlewareParams struct {
	fx.In
	TokenService ports.TokenService
}

// NewAuthMiddleware - це функція-конструктор для Gin middleware.
func NewAuthMiddleware(p AuthMiddlewareParams) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "authorization header is required",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token format. Expected 'Bearer <token>'",
			})
			return
		}

		jti, userID, err := p.TokenService.ValidateToken(c.Request.Context(), tokenString, true)
		if err != nil {
			if errors.Is(err, domain.ErrTokenMalformed) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "token malformed",
				})
				return
			} else if errors.Is(err, domain.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "token signature invalid",
				})
				return
			} else if errors.Is(err, domain.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "token expired",
				})
				return
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				})
				return
			}
		}

		isBlacklisted, err := p.TokenService.IsAccessTokenBlacklisted(c.Request.Context(), jti)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				domain.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Failed to check token in blacklist",
				})
			return
		}
		if isBlacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				domain.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "Access token has been revoked (blacklisted)",
				})
			return
		}

		c.Set("userID", userID)
		c.Set("jti", jti)

		c.Next()
	}
}
