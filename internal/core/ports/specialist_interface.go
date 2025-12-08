package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"
	"time"

	"github.com/gin-gonic/gin"
)

type SpecialistHandlers interface {
	Registration(c *gin.Context)
	Login(c *gin.Context)
	Me(c *gin.Context)
	ChangePassword(c *gin.Context)
	Logout(c *gin.Context)
	UpdateProfile(c *gin.Context)
	DeactivateProfile(c *gin.Context)
	DeleteAccount(c *gin.Context)
}

type SpecialistService interface {
	Registration(ctx context.Context, req domain.RegistrationRequest) (int64, error)
	Login(ctx context.Context, email string, password string) (domain.SpecialistProfDTO, error)
	ShowByID(ctx context.Context, id int64) (domain.SpecialistProfDTO, error)
	ShowByEmail(ctx context.Context, email string) (domain.SpecialistProfDTO, error)
	ChangePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateAvatar(ctx context.Context, specialistID int64, avatarURL string) error
	UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.SpecialistProfDTO, error)
	// DeactivateProfile handles the business logic for changing a profile's active status.
	DeactivateProfile(ctx context.Context, specialistID int64, isActive bool) error
	AddImages(ctx context.Context, specialistID int64, imageURLs []string) error
	DeleteImage(ctx context.Context, specialistID int64, imageURL string) error
	InitiateSoftDelete(ctx context.Context, id int64) error
	DeleteExpiredAccounts(ctx context.Context) error
}

type SpecialistRepository interface {
	Save(ctx context.Context, name, email, phone, hash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (domain.Specialist, error)
	GetByID(ctx context.Context, id int64) (domain.Specialist, error)
	CheckFieldValueExists(ctx context.Context, fieldName string, fieldValue string) (bool, error)
	UpdatePasswordHash(ctx context.Context, id int64, newHash string) error
	UpdateAvatar(ctx context.Context, id int64, avatarURL string) error
	UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.Specialist, error)
	// UpdateIsActive changes the active status of a specialist.
	UpdateIsActive(ctx context.Context, id int64, isActive bool) error
	AddImages(ctx context.Context, specialistID int64, imageURLs []string) error
	DeleteImage(ctx context.Context, specialistID int64, imageURL string) error
	MarkAsDeleted(ctx context.Context, id int64) error
	// GetExpiredAccounts returns a list of accounts whose deletion grace period has expired.
	GetExpiredAccounts(ctx context.Context, thresholdTime time.Time) ([]domain.Specialist, error)
	// DeleteAllServices removes all service records for a specific specialist.
	DeleteAllServices(ctx context.Context, specialistID int64) error
	// HardDelete permanently deletes a specialist record by ID.
	HardDelete(ctx context.Context, id int64) error
}
