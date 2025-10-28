package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"

	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/pkg/utils"
)

type SpecialistServiceImpl struct {
	specialistRepo ports.SpecialistRepository
	logger         *zap.Logger
	defaultTimeout time.Duration
}

var _ ports.SpecialistService = (*SpecialistServiceImpl)(nil)

func NewSpecialistService(repo ports.SpecialistRepository, logger *zap.Logger, cfg config.AuthConfig) *SpecialistServiceImpl {
	return &SpecialistServiceImpl{
		specialistRepo: repo,
		logger:         logger,
		defaultTimeout: cfg.DefaultTimeout,
	}
}

func (ss *SpecialistServiceImpl) Registration(ctx context.Context, specialist domain.RegistrationRequest) (int64, error) {

	//hash password
	hashedPassword, err := utils.HashGen(specialist.Password)
	if err != nil {
		ss.logger.Error("failed to generate password hash during registration",
			zap.String("email", specialist.Email),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	exists, err := ss.specialistRepo.CheckFieldValueExists(timeoutCtx, "email", specialist.Email)
	if err != nil {
		ss.logger.Error("database check for existing email failed",
			zap.String("email", specialist.Email),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}
	if exists {
		ss.logger.Warn("registration attempt with existing email", zap.String("email", specialist.Email))
		return 0, domain.ErrAccountAlreadyExists
	}

	phone, err := utils.NormalizePhoneNumber(specialist.Phone)
	if err != nil {
		ss.logger.Error("failed to convert phone number before save",
			zap.String("phone", specialist.Phone),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}

	id, err := ss.specialistRepo.Save(timeoutCtx, specialist.Name, specialist.Email, phone, hashedPassword)
	if err != nil {
		ss.logger.Error("failed to save new specialist to database", zap.String("email", specialist.Email), zap.Error(err))
		return 0, domain.ErrInternalServer
	}

	return id, nil
}

func (ss *SpecialistServiceImpl) Login(ctx context.Context, email, password string) (domain.SpecialistProfDTO, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfDTO{}

	// Check email exists
	specialistModel, err := ss.specialistRepo.GetByEmail(timeoutCtx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("login attempt for non-existent user", zap.String("email", email))
			return specialistDTO, domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to get specialist by email during login",
			zap.String("email", email),
			zap.Error(err))
		return specialistDTO, domain.ErrInternalServer
	}

	//verify password
	if err := utils.HashCompare(specialistModel.PasswordHash, password); err != nil {
		ss.logger.Warn("login failed due to invalid password", zap.String("email", email))
		return specialistDTO, domain.ErrInvalidCredentials
	}

	specialistDTO.ID = specialistModel.ID
	specialistDTO.Name = specialistModel.Name
	specialistDTO.Email = specialistModel.Email
	specialistDTO.IsVerified = specialistModel.IsVerified

	return specialistDTO, nil
}

func (ss *SpecialistServiceImpl) ShowByID(ctx context.Context, id int64) (domain.SpecialistProfDTO, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfDTO{}

	specialistModel, err := ss.specialistRepo.GetByID(timeoutCtx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("specialist not found by ID",
				zap.Int64("id", id),
				zap.Error(err))
			return specialistDTO, domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to retrieve specialist by ID from database",
			zap.Int64("id", id),
			zap.Error(err))
		return specialistDTO, domain.ErrInternalServer
	}

	specialistDTO = utils.ToSpecialistProfileDTO(specialistModel)

	return specialistDTO, nil
}

func (ss *SpecialistServiceImpl) ShowByEmail(ctx context.Context, email string) (domain.SpecialistProfDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfDTO{}

	specialistModel, err := ss.specialistRepo.GetByEmail(timeoutCtx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("specialist not found by email",
				zap.String("email", email),
				zap.Error(err))
			return specialistDTO, domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to retrieve specialist by email from database",
			zap.String("email", email),
			zap.Error(err))
		return specialistDTO, domain.ErrInternalServer
	}

	specialistDTO = utils.ToSpecialistProfileDTO(specialistModel)

	return specialistDTO, nil
}

func (ss *SpecialistServiceImpl) ChangePassword(ctx context.Context, id int64, CurrentPass, newPass string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialist, err := ss.specialistRepo.GetByID(timeoutCtx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("attempt to update password for non-existent user", zap.Int64("id", id))
			return domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to get specialist by id during password update", zap.Int64("id", id), zap.Error(err))
		return domain.ErrInternalServer
	}

	if err := utils.HashCompare(specialist.PasswordHash, CurrentPass); err != nil {
		ss.logger.Warn("password update failed due to invalid old password", zap.Int64("id", id))
		return domain.ErrInvalidCredentials
	}

	hashedPassword, err := utils.HashGen(newPass)
	if err != nil {
		ss.logger.Error("failed to generate password hash during password update",
			zap.Int64("id", id),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	if err := ss.specialistRepo.UpdatePasswordHash(timeoutCtx, id, hashedPassword); err != nil {
		ss.logger.Error("failed to update password hash in database",
			zap.Int64("id", id),
			zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return domain.ErrInternalServer
	}

	return nil
}

func (ss *SpecialistServiceImpl) UpdateAvatar(ctx context.Context, specialistID int64, avatarURL string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	// First, check if the specialist exists.
	if _, err := ss.specialistRepo.GetByID(timeoutCtx, specialistID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("attempt to update avatar for non-existent specialist", zap.Int64("id", specialistID))
			return domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to get specialist by ID during avatar update", zap.Int64("id", specialistID), zap.Error(err))
		return domain.ErrInternalServer
	}

	if err := ss.specialistRepo.UpdateAvatar(timeoutCtx, specialistID, avatarURL); err != nil {
		ss.logger.Error("failed to update avatar in DB", zap.Int64("id", specialistID), zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return domain.ErrInternalServer
	}

	return nil
}

func (ss *SpecialistServiceImpl) UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.SpecialistProfDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	updatedSpec := domain.SpecialistProfDTO{}
	// First, check if the specialist exists.
	// This also helps to ensure we're not trying to update a non-existent record and provides a better error message.
	if _, err := ss.specialistRepo.GetByID(timeoutCtx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("attempt to update profile for non-existent specialist", zap.Int64("id", id))
			return updatedSpec, domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to get specialist by ID during profile update check",
			zap.Int64("id", id),
			zap.Error(err))
		return updatedSpec, domain.ErrInternalServer
	}

	// Normalize the phone number only if it's provided in the request.
	if req.Phone != nil {
		phone, err := utils.NormalizePhoneNumber(*req.Phone)
		if err != nil {
			ss.logger.Error("failed to convert phone number before save",
				zap.String("phone", *req.Phone),
				zap.Error(err))
			return updatedSpec, domain.ErrInternalServer
		}
		*req.Phone = phone
	}

	// Perform the update and get the updated model back
	updatedModel, err := ss.specialistRepo.UpdateProfile(timeoutCtx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This can happen in a race condition if the user is deleted between the check and the update.
			ss.logger.Warn("specialist disappeared during profile update", zap.Int64("id", id))
			return updatedSpec, domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to update specialist profile in database",
			zap.Int64("id", id),
			zap.Error(err))
		return updatedSpec, domain.ErrInternalServer
	}

	// Convert the updated model to a DTO and return it
	return utils.ToSpecialistProfileDTO(updatedModel), nil
}


func (ss *SpecialistServiceImpl) AddImages(ctx context.Context, specialistID int64, imageURLs []string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	if err := ss.specialistRepo.AddImages(timeoutCtx, specialistID, imageURLs); err != nil {
		ss.logger.Error("failed to add images to specialist profile in DB",
			zap.Int64("id", specialistID),
			zap.Strings("urls", imageURLs),
			zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return domain.ErrInternalServer
	}

	return nil
}

func (ss *SpecialistServiceImpl) DeleteImage(ctx context.Context, specialistID int64, imageURL string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	if err := ss.specialistRepo.DeleteImage(timeoutCtx, specialistID, imageURL); err != nil {
		ss.logger.Error("failed to delete images from specialist profile in DB",
			zap.Int64("id", specialistID),
			zap.String("url", imageURL),
			zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return domain.ErrInternalServer
	}

	return nil
}

func (ss *SpecialistServiceImpl) SearchSpecialistByServicePetArea(ctx context.Context, specialist domain.SearchSpecialistParams) ([]domain.SpecialistProfDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := []domain.SpecialistProfDTO{}

	limit := 0
	offset := 0

	specialistModels, err := ss.specialistRepo.SearchSpecialistByServicePetArea(timeoutCtx, specialist, limit, offset) 
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
	}
	
	if err != nil { 
		if errors.Is(err, sql.ErrNoRows)  || errors.Is(err, domain.ErrNotFound){
			ss.logger.Warn("specialist not found by search params",
				zap.Int64("animal", specialist.Animal),
				zap.Int64("animal_size", specialist.AnimalSize),
				zap.Int64("service", specialist.Service),
				zap.Int64("area", specialist.Area),
				zap.Int("limit", limit),
				zap.Int("offset", offset),
				zap.Error(err))
			return nil, domain.ErrSpecislistsNotFound 
		}
		ss.logger.Error("failed to retrieve specialist by service/pet/area from database",
		    zap.Int64("animal", specialist.Animal),
			zap.Int64("animal_size", specialist.AnimalSize),
			zap.Int64("service", specialist.Service),
			zap.Int64("area", specialist.Area),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	if len(specialistModels) == 0 {
		return []domain.SpecialistProfDTO{}, nil
	}

	profiles := make([]domain.SpecialistProfDTO, 0, len(specialistModels))

	for _, specialistModel := range specialistModels {
		profiles = append(specialistDTO, utils.ToSpecialistProfileDTO(specialistModel))
	}

	return profiles, nil
}

func (ss *SpecialistServiceImpl) GetSpecialistDetailsById(ctx context.Context, specialistId int64) (domain.SpecialistDetailsDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistDetailsDTO{}

	specialistDetails, err := ss.specialistRepo.GetSpecialistDetailsById(timeoutCtx, specialistId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("specialist not found by ID",
				zap.Int64("id", specialistId),
				zap.Error(err))
			return specialistDTO, domain.ErrSpecislistsNotFound
		}
		ss.logger.Error("failed to retrieve specialist by ID",
			zap.Int64("id", specialistId),
			zap.Error(err))
		return specialistDTO, domain.ErrInternalServer
	}

	specialistDTO = utils.ToSpecialistsDetailsDTO(specialistDetails)

	return specialistDTO, nil
}

