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

func (ss *SpecialistServiceImpl) Registration(ctx context.Context, specialist *domain.RegistrationRequest) (int64, error) {

	//hash password
	hashedPassword, err := utils.HashGen(specialist.Password)
	if err != nil {
		ss.logger.Error("Failed to generate password hash during registration",
			zap.String("email", specialist.Email),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	exists, err := ss.specialistRepo.CheckFieldValueExists(timeoutCtx, "email", specialist.Email)
	if err != nil {
		ss.logger.Error("Database check for existing email failed",
			zap.String("email", specialist.Email),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}
	if exists {
		ss.logger.Warn("Registration attempt with existing email", zap.String("email", specialist.Email))
		return 0, domain.ErrAccountAlreadyExists
	}

	id, err := ss.specialistRepo.Save(timeoutCtx, specialist.Name, specialist.Email, specialist.Phone, hashedPassword)
	if err != nil {
		ss.logger.Error("Failed to save new specialist to database", zap.String("email", specialist.Email), zap.Error(err))
		return 0, domain.ErrInternalServer
	}

	return id, nil
}

func (ss *SpecialistServiceImpl) Login(ctx context.Context, email, password string) (domain.Specialist, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	// Check email exists
	specialistData, err := ss.specialistRepo.GetByEmail(timeoutCtx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("Login attempt for non-existent user", zap.String("email", email))
			return domain.Specialist{}, domain.ErrAccountNotFound
		}
		ss.logger.Error("Failed to get specialist by email during login",
			zap.String("email", email),
			zap.Error(err))
		return domain.Specialist{}, domain.ErrInternalServer
	}

	//verify password
	if err := utils.HashCompare(specialistData.PasswordHash, password); err != nil {
		ss.logger.Warn("Login failed due to invalid password", zap.String("email", email))
		return domain.Specialist{}, domain.ErrInvalidCredentials
	}

	return specialistData, nil
}

func (ss *SpecialistServiceImpl) ShowByID(ctx context.Context, id int64) (domain.Specialist, error) {
	//Retrieve stored hash and user ID
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	spec, err := ss.specialistRepo.GetByID(timeoutCtx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("Specialist not found by ID",
				zap.Int64("id", id),
				zap.Error(err))
			return domain.Specialist{}, domain.ErrAccountNotFound
		}
		ss.logger.Error("Failed to retrieve specialist by ID from database",
			zap.Int64("id", id),
			zap.Error(err))
		return domain.Specialist{}, domain.ErrInternalServer
	}
	return spec, nil
}

func (ss *SpecialistServiceImpl) ShowByEmail(ctx context.Context, email string) (domain.Specialist, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	spec, err := ss.specialistRepo.GetByEmail(timeoutCtx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ss.logger.Warn("Specialist not found by email",
				zap.String("email", email),
				zap.Error(err))
			return domain.Specialist{}, domain.ErrAccountNotFound
		}
		ss.logger.Error("Failed to retrieve specialist by email from database",
			zap.String("email", email),
			zap.Error(err))
		return domain.Specialist{}, domain.ErrInternalServer
	}
	return spec, nil
}
