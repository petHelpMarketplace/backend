package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type SpecialistHandlers interface {
	Registration(c *gin.Context)
	Login(c *gin.Context)
	Me(c *gin.Context)
	ChangePassword(c *gin.Context)
	Logout(c *gin.Context)
	UpdateProfile(c *gin.Context)
	GetSpecialistsByAreaAnimalService(c *gin.Context) 
}

type SpecialistService interface {
	Registration(ctx context.Context, req domain.RegistrationRequest) (int64, error)
	Login(ctx context.Context, email string, password string) (domain.SpecialistProfDTO, error)
	ShowByID(ctx context.Context, id int64) (domain.SpecialistProfDTO, error)
	ShowByEmail(ctx context.Context, email string) (domain.SpecialistProfDTO, error)
	ChangePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateAvatar(ctx context.Context, specialistID int64, avatarURL string) error
	UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.SpecialistProfDTO, error)
	AddImages(ctx context.Context, specialistID int64, imageURLs []string) error
	DeleteImage(ctx context.Context, specialistID int64, imageURL string) error
	SearchSpecialistByServicePetArea(ctx context.Context, specialist domain.SearchSpecialistParams) ([]domain.SpecialistProfDTO, error) 
}

type SpecialistRepository interface {
	Save(ctx context.Context, name, email, phone, hash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (domain.Specialist, error)
	GetByID(ctx context.Context, id int64) (domain.Specialist, error)
	CheckFieldValueExists(ctx context.Context, fieldName string, fieldValue string) (bool, error)
	UpdatePasswordHash(ctx context.Context, id int64, newHash string) error
	UpdateAvatar(ctx context.Context, id int64, avatarURL string) error
	UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.Specialist, error)
}
