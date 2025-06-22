package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	operationSpHandler = "specialist_handler: "
)

type SpecialistHandlerImpl struct {
	validator         ports.SpecialistValidator
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	logger            *zap.Logger
}

var _ ports.SpecialistHandlers = (*SpecialistHandlerImpl)(nil)

func NewSpecialistHandler(specialistSrv ports.SpecialistService, tokenSrv ports.TokenService, validator ports.SpecialistValidator, logger *zap.Logger) *SpecialistHandlerImpl {
	return &SpecialistHandlerImpl{
		validator:         validator,
		specialistService: specialistSrv,
		tokenService:      tokenSrv,
		logger:            logger,
	}
}

type successRegistration struct {
	ID      string `json:"id"`
	Message string `json:"message" default:"Registration successful"`
}

// @Summary Registration
// @Description New specialist registration
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body domain.RegistrationRequest true "Registration request body"
// @Success 201 {object} successRegistration "Sign in succeeded"
// @Failure      400,409,500 {object} domain.RequestResponse
// @Router /specialist/register [post]
func (sh *SpecialistHandlerImpl) Registration(c *gin.Context) {

	req := &domain.RegistrationRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		bindErr := fmt.Errorf("%s invalid registration payload %w", operationSpHandler, err)
		sh.logger.Error("bindJSON failed", zap.Error(bindErr))
		c.JSON(http.StatusBadRequest, domain.RequestResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid registration payload",
		})
		return
	}
	// — Run field‐validator on it
	if err := sh.validator.Validate(req); err != nil {
		validateErr := fmt.Errorf("%s invalid registration payload %w", operationSpHandler, err)
		sh.logger.Error("validate failed", zap.Error(validateErr))
		c.JSON(http.StatusBadRequest, domain.RequestResponse{
			Code:    http.StatusBadRequest,
			Message: "validation error, registration data failed",
		})
		return
	}

	//email uniqueness
	id, err := sh.specialistService.Registration(c.Request.Context(), req)
	if err != nil {
		if err == domain.ErrEmailAlreadyInUse {
			c.JSON(http.StatusConflict, domain.RequestResponse{
				Code:    http.StatusConflict,
				Message: "specialist with this email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	// — Success response
	c.JSON(http.StatusCreated, successRegistration{
		ID:      strconv.FormatInt(id, 10),
		Message: "Registration successful",
	})

}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Login
// @Description Login specialist
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body LoginDTO true "Login request body"
// @Success 200 {object} domain.TokensPair "Login succeeded"
// @Failure      400,401,500 {object} domain.RequestResponse
// @Router /specialist/login [post]
func (sh *SpecialistHandlerImpl) Login(c *gin.Context) {
	var dto LoginDTO

	//bind and validate JSON payload
	if err := c.ShouldBindJSON(&dto); err != nil {
		bindErr := fmt.Errorf("%s invalid registration payload %w", operationSpHandler, err)
		sh.logger.Error("bindJSON failed", zap.Error(bindErr))
		c.JSON(http.StatusBadRequest, domain.RequestResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid registration payload",
		})
		return
	}

	//perform authentication and token issuance
	spec, err := sh.specialistService.Login(c.Request.Context(), dto.Email, dto.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			loginErr := fmt.Errorf("%s invalid password %w", operationSpHandler, err)
			sh.logger.Error("login failed", zap.Error(loginErr))
			c.JSON(http.StatusUnauthorized, domain.RequestResponse{
				Code:    http.StatusUnauthorized,
				Message: "wrong password",
			})
		} else if errors.Is(err, domain.ErrAccountNotFound) {
			loginErr := fmt.Errorf("%s email not found %w", operationSpHandler, err)
			sh.logger.Error("login failed", zap.Error(loginErr))
			c.JSON(http.StatusNotFound, domain.RequestResponse{
				Code:    http.StatusNotFound,
				Message: "user with this email not found",
			})

		} else {
			loginErr := fmt.Errorf("%s failed to glogin specialist %w", operationSpHandler, err)
			sh.logger.Error("login failed", zap.Error(loginErr))
			c.JSON(http.StatusInternalServerError, domain.RequestResponse{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			})
		}
		return
	}

	tokens, err := sh.tokenService.GenerateTokenPair(c.Request.Context(), &spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	//return both tokens in the response body
	c.JSON(http.StatusOK, tokens)
}
