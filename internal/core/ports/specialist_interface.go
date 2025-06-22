package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type SpecialistHandlers interface {
	Registration(c *gin.Context)
	Login(c *gin.Context)
}

type SpecialistService interface {
	Registration(ctx context.Context, req *domain.RegistrationRequest) (int64, error)
	Login(ctx context.Context, email string, password string) (domain.Specialist, error)
	ShowByID(ctx context.Context, id int64) (domain.Specialist, error)
	ShowByEmail(ctx context.Context, email string) (domain.Specialist, error)
	// List(filter dto.ListFilter) ([]domain.User, error)
	// Update(user domain.User) error
	// Delete(id uint) error
}

type SpecialistRepository interface {
	Save(ctx context.Context, name, family_name, email, phone, hash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (domain.Specialist, error)
	GetByID(ctx context.Context, id int64) (domain.Specialist, error)
	CheckFieldValueExists(ctx context.Context, fieldName string, fieldValue string) (bool, error)
	// Find(id uint) (domain.User, error)
	// List(filter dto.ListFilter) ([]domain.User, error)
	// Update(user domain.User) error
	// Delete(id uint) error
}
