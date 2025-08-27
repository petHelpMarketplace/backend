package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//prefix log messages so it’s clear which handler produced the log
const (
	operationUnauthAppHandler = "unauth_appointment_handler: "
)

type UnauthAppintmentHandlerImpl struct {
	//checks request payload fields for correctness
	validator         ports.UnauthAppointmentValidator
	unauthAppointmentService ports.UnauthAppointmentService
	logger            *zap.Logger
}

//Compile-time check that this struct implements the SpecialistHandlers interface
var _ ports.UnauthAppointmentHandler = (*UnauthAppintmentHandlerImpl)(nil)

//Creates a new handler and injects dependencies.
func NewUnauthAppointmentHandler(unauthAppointmentSrv ports.UnauthAppointmentService, tokenSrv ports.TokenService, validator ports.UnauthAppointmentValidator, logger *zap.Logger) *UnauthAppintmentHandlerImpl {
	return &UnauthAppintmentHandlerImpl{
		validator:         validator,
		unauthAppointmentService: unauthAppointmentSrv,
		logger:            logger,
	}
}

type successSaveUnauthAppointment struct {
	ID      string `json:"id" example:"1"`
	Message string `json:"message" default:"An appointment booked successfully"`
}

// @Summary Book un appointment
// @Description New  appointment booking by unauth user
// @Tags SpecialistAppointment
// @Accept       json
// @Produce      json
// @Param request body domain.SaveUnauthAppointmentnRequest true "UnauthAppointment request body"
// @Success 201 {object} successSaveUnauthAppointment "Booking appointment succeeded"
// @Failure      400,409,500 {object} domain.ErrorResponse
// @Router /public-appointment-request [post]

//Handles HTTP POST requests to create unauthenticated appointments
func (ah *UnauthAppintmentHandlerImpl) Book (c *gin.Context) {

	//Bind JSON request
	req := domain.SaveUnauthAppointmentnRequest{}

	//parses request body into req struct
	if err := c.ShouldBindJSON(&req); err != nil {

		var fieldErrors []domain.FieldError
		message := "Invalid request payload"

		//UnmarshalTypeError → field type mismatch (e.g., number sent as string).
		var jsonErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError

		bindErr := fmt.Errorf("%s invalid bookin unauth appointment payload: %w", operationUnauthAppHandler, err)

		if errors.As(err, &jsonErr) {
			message = "The request contains invalid data types."
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   jsonErr.Field,
				Message: fmt.Sprintf("Expected type '%s' for field.", jsonErr.Type),
			})
		} else if errors.As(err, &syntaxErr) {
			message = "The request body is not valid JSON."
		//io.EOF --> empty body.
		} else if err == io.EOF {
			message = "Request body cannot be empty."
		}

		ah.logger.Error("bind JSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}

	//Validate request fields
	if validationErrors := ah.validator.ValidateUnauthAppointmentRequest(req); len(validationErrors) > 0 {
		ah.logger.Error("validation failed", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: validationErrors,
		})
		return
	}

	id, err :=ah.unauthAppointmentService.BookUnauthAppointment(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrTimeUnavailable) {
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Code:    http.StatusConflict,
				Message: fmt.Sprintf("time %s %s-%s already booked", req.Date, req.StartTime, req.EndTime),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	// — Success response
	c.JSON(http.StatusCreated, successSaveUnauthAppointment{
		ID:      strconv.FormatInt(id, 10),
		Message: "An appointment booked successfully",
	})

}





