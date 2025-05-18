package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.uber.org/zap"

	"pethelp-backend/internal/domain/models"
	"pethelp-backend/internal/domain/service"
)

const maxBodySize = 1 << 20 // 1 MiB

type RegistrationRequest struct {
	Name                 string `json:"name" binding:"required,min=2" example:"John"`
	FamilyName           string `json:"family_name" binding:"required,min=2" example:"Doe"`
	Phone                string `json:"phone" binding:"required,regexp=^\\+[0-9]{1,3}[0-9\\- ()]{7,}$" example:"+12345678901"`
	Email                string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password             string `json:"password" binding:"required,min=12" example:"SuperSecret123"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password" example:"SuperSecret123"`
}

func isValidPassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (r *RegistrationRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name", "FamilyName":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least 2 characters", err.Field()))
			case "Phone":
				errorMessages = append(errorMessages, "Phone must be in E.123 format (e.g., +38 (XXX) XXX-XX-XX)")
			case "Email":
				errorMessages = append(errorMessages, "Invalid email format")
			case "Password":
				errorMessages = append(errorMessages, "Password must be at least 12 characters")
			case "PasswordConfirmation":
				errorMessages = append(errorMessages, "Passwords do not match")
			}
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}
	return isValidPassword(r.Password)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
	Token   string `json:"token"`
}

// RegisterSpecialist godoc
// @Summary Register a new specialist
// @Description Register a new specialist account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegistrationRequest true "Specialist registration payload"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/specialists/register [post]
func RegisterSpecialistHandler(authService *service.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		lr := &io.LimitedReader{R: c.Request.Body, N: maxBodySize}
		bodyBytes, err := io.ReadAll(lr)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			return
		}
		if lr.N <= 0 {
			logger.Warn("Request body too large")
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var req RegistrationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		if err := req.Validate(); err != nil {
			logger.Error("Validation failed", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exists, err := authService.CheckEmailExists(req.Email)
		if err != nil {
			logger.Error("Failed to check email existence", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		exists, err = authService.CheckPhoneExists(req.Phone)
		if err != nil {
			logger.Error("Failed to check phone existence", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
			return
		}

		newSpecialist := &models.Specialist{
			Name:       req.Name,
			FamilyName: req.FamilyName,
			Phone:      req.Phone,
			Email:      req.Email,
			Password:   req.Password,
			IsBanned:   false,
			IsDeleted:  false,
			IsActive:   true,
			IsVerified: false,
		}

		err = authService.RegisterSpecialist(newSpecialist)
		if err != nil {
			logger.Error("Registration failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register specialist"})
			return
		}

		token, err := authService.GenerateToken(newSpecialist)
		if err != nil {
			logger.Error("Failed to generate token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Specialist registered successfully",
			"id":      newSpecialist.ID,
			"token":   token,
		})
	}
}
