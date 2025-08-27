package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"
	"time"

	"github.com/gin-gonic/gin"
)


type UnauthAppointmentRepository interface {
	Save(ctx context.Context,
	serviceID, cityID, districtID, animalSizeID, specialistID int,
	amount float32,
	locationType, street, unit, apt, description, email string,
	date, startTime, endTime time.Time) (int64, error)
	IsTimeBooked(ctx context.Context, specialistID int, data, startTime, endTime time.Time) (bool, error)
}

type UnauthAppointmentHandler interface {
	Book(c *gin.Context)
}

type UnauthAppointmentService interface {
	BookUnauthAppointment(ctx context.Context, req domain.SaveUnauthAppointmentnRequest) (int64, error)
}