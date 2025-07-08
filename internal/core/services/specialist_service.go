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

func (ss *SpecialistServiceImpl) Login(ctx context.Context, email, password string) (domain.SpecialistProfileDTO, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfileDTO{}

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

func (ss *SpecialistServiceImpl) ShowByID(ctx context.Context, id int64) (domain.SpecialistProfileDTO, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfileDTO{}

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

func (ss *SpecialistServiceImpl) ShowByEmail(ctx context.Context, email string) (domain.SpecialistProfileDTO, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	specialistDTO := domain.SpecialistProfileDTO{}

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
