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
	fileService    ports.FileUploadService
	logger         *zap.Logger
	defaultTimeout time.Duration
}

var _ ports.SpecialistService = (*SpecialistServiceImpl)(nil)

func NewSpecialistService(repo ports.SpecialistRepository, fileSrv ports.FileUploadService, logger *zap.Logger, cfg config.AuthConfig) *SpecialistServiceImpl {
	return &SpecialistServiceImpl{
		specialistRepo: repo,
		fileService:    fileSrv,
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
	if specialistModel.Name.Valid {
		specialistDTO.Name = &specialistModel.Name.String
	}
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

	// check district exist if it update
	district := req.District
	if district != nil {
		var exists bool
		var err error
		if exists, err = ss.specialistRepo.CheckDistrict(timeoutCtx, *district); err != nil {
			ss.logger.Error("failed to update profile with district",
				zap.String("district_name", *district),
				zap.Error(err))
			return updatedSpec, domain.ErrInternalServer
		}

		if !exists {
			ss.logger.Error("attempt to update profile with non-existent district", zap.String("district_name", *district))
			return updatedSpec, domain.ErrDistrictNotFound
		}
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

func (ss *SpecialistServiceImpl) DeactivateProfile(ctx context.Context, id int64, isActive bool) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	err := ss.specialistRepo.UpdateIsActive(timeoutCtx, id, isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("specialist disappeared during profile active status update", zap.Int64("id", id))
			return domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to update specialist profile active status in database",
			zap.Int64("id", id),
			zap.Bool("isActive", isActive),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil

}

// InitiateSoftDelete marks the specialist's account for deletion.
// It first checks if the user exists, then sets the 'is_deleted' flag and schedules the deletion.
func (ss *SpecialistServiceImpl) InitiateSoftDelete(ctx context.Context, id int64) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	// Check if the user exists before attempting to delete
	if _, err := ss.specialistRepo.GetByID(timeoutCtx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, domain.ErrAccountNotFound) {
			ss.logger.Warn("attempt to soft delete non-existent user", zap.Int64("id", id))
			return domain.ErrAccountNotFound
		}
		ss.logger.Error("failed to check specialist existence during soft delete",
			zap.Int64("id", id),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	// Proceed with marking the account as deleted
	err := ss.specialistRepo.MarkAsDeleted(timeoutCtx, id)
	if err != nil {
		ss.logger.Error("failed to mark specialist as deleted in database",
			zap.Int64("id", id),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}

// DeleteExpiredAccounts handles the comprehensive cleanup of expired profiles.
func (ss *SpecialistServiceImpl) DeleteExpiredAccounts(ctx context.Context) error {
	// Define timeout and threshold
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute) // Increased timeout for S3 operations
	defer cancel()

	retentionPeriod := 7 * 24 * time.Hour

	thresholdTime := time.Now().Add(-retentionPeriod)

	// Fetch accounts to be deleted
	expiredAccounts, err := ss.specialistRepo.GetExpiredAccounts(timeoutCtx, thresholdTime)
	if err != nil {
		ss.logger.Error("failed to fetch expired accounts", zap.Error(err))
		return domain.ErrInternalServer
	}

	if len(expiredAccounts) == 0 {
		ss.logger.Info("no expired accounts found for deletion")
		return nil
	}

	ss.logger.Info("starting cleanup for expired accounts", zap.Int("count", len(expiredAccounts)))

	// Iterate and clean up each account
	for _, specialist := range expiredAccounts {
		logger := ss.logger.With(zap.Int64("specialistID", specialist.ID))

		// Delete Associated Services from DB
		if err := ss.specialistRepo.DeleteAllServices(timeoutCtx, specialist.ID); err != nil {
			logger.Error("failed to delete associated services", zap.Error(err))
		}

		// Hard Delete the Specialist Record
		if err := ss.specialistRepo.HardDelete(timeoutCtx, specialist.ID); err != nil {
			logger.Error("failed to hard delete specialist record", zap.Error(err))
			continue
		}

		// Delete Avatar from S3
		if specialist.Avatar.Valid && specialist.Avatar.String != "" {
			if err := ss.fileService.DeleteAvatar(timeoutCtx, specialist.Avatar.String); err != nil {
				// Log error but continue deletion process
				logger.Error("failed to delete avatar from S3",
					zap.String("url", specialist.Avatar.String),
					zap.Error(err))
			}
		}

		// Delete Portfolio Images from S3
		for _, img := range specialist.ImageID {
			if img.Valid && img.String != "" {
				if err := ss.fileService.DeletePortfolioImage(timeoutCtx, img.String); err != nil {
					logger.Error("failed to delete portfolio image from S3",
						zap.String("url", img.String),
						zap.Error(err))
				}
			}
		}

		logger.Info("successfully deleted expired specialist account and resources")
	}

	return nil
}
func (ss *SpecialistServiceImpl) SearchSpecialistByServicePetArea(ctx context.Context, specialist domain.SearchSpecialistParams) ([]domain.SpecialistProfileSearchResponseDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

    
	const defaultLimit = 20
	limit := defaultLimit
	offset := 0

	specialistModels, err := ss.specialistRepo.SearchSpecialistByServicePetArea(timeoutCtx, specialist, limit, offset) 
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return nil, err
	}

	if errors.Is(err, domain.ErrInvalidParameter) {
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
			return nil, domain.ErrSpecialistsNotFound 
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
		return []domain.SpecialistProfileSearchResponseDTO{}, nil
	}

	return specialistModels, nil
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
			return specialistDTO, domain.ErrSpecialistsNotFound
		}
		ss.logger.Error("failed to retrieve specialist by ID",
			zap.Int64("id", specialistId),
			zap.Error(err))
		return specialistDTO, domain.ErrInternalServer
	}

	specialistDTO = utils.ToSpecialistsDetailsDTO(specialistDetails)

	return specialistDTO, nil
}