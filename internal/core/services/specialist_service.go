package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/pkg/utils"
)

const (
	operationSpecServ = "specialist_service: "
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
		hashErr := fmt.Errorf("%s failed to generate password hash: %w", operationSpecServ, err)
		ss.logger.Error(domain.ErrFailedToHashPassword.Error(), zap.Error(hashErr))
		return 0, domain.ErrInternalServer
	}

	//run
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	exists, err := ss.specialistRepo.CheckFieldValueExists(timeoutCtx, "email", specialist.Email)
	if err != nil {
		checkErr := fmt.Errorf("%s failed to checking exists email: %w", operationSpecServ, err)
		ss.logger.Error("email exists check failed", zap.String("email", specialist.Email), zap.Error(checkErr))
		return 0, checkErr
	} else if exists {
		existErr := fmt.Errorf("%s email already registered", operationSpecServ)
		ss.logger.Error(domain.ErrEmailAlreadyInUse.Error(), zap.Error(existErr))
		return 0, domain.ErrAccountAlreadyExists
	}

	id, err := ss.specialistRepo.Save(timeoutCtx, specialist.Name, specialist.FamilyName, specialist.Email, specialist.Phone, hashedPassword)
	if err != nil {
		saveErr := fmt.Errorf("%s failed to save specialist: %w", operationName, err)
		ss.logger.Error("failed insert to database", zap.Error(saveErr))
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
			getEmailErr := fmt.Errorf("%s failed to get specialist by email - %s: %w", operationSpecServ, email, err)
			ss.logger.Error("specialist not found", zap.Error(getEmailErr))
			return specialistData, domain.ErrAccountNotFound
		}
		getEmailErr := fmt.Errorf("%s failed to get specialist by email - %s: %w", operationSpecServ, email, err)
		ss.logger.Error("specialist found error", zap.Error(getEmailErr))
		return specialistData, domain.ErrInternalServer

	}

	//verify password
	if err := utils.HashCompare(specialistData.PasswordHash, password); err != nil {
		ss.logger.Warn("login failed: bad credentials", zap.String("email", email), zap.Error(err))
		return specialistData, domain.ErrInvalidCredentials
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
			getEmailErr := fmt.Errorf("%s failed to get specialist by id - %d: %w", operationSpecServ, id, err)
			ss.logger.Error("specialist not found", zap.Error(getEmailErr))
			return spec, domain.ErrAccountNotFound
		}
		getEmailErr := fmt.Errorf("%s failed to get specialist by id - %d: %w", operationSpecServ, id, err)
		ss.logger.Error("specialist found error", zap.Error(getEmailErr))
		return spec, domain.ErrInternalServer

	}

	return spec, nil
}

func (ss *SpecialistServiceImpl) ShowByEmail(ctx context.Context, email string) (domain.Specialist, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ss.defaultTimeout)
	defer cancel()

	spec, err := ss.specialistRepo.GetByEmail(timeoutCtx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			getEmailErr := fmt.Errorf("%s failed to get specialist by email - %s: %w", operationSpecServ, email, err)
			ss.logger.Error("specialist not found", zap.Error(getEmailErr))
			return spec, domain.ErrAccountNotFound
		}
		getEmailErr := fmt.Errorf("%s failed to get specialist by email - %s: %w", operationSpecServ, email, err)
		ss.logger.Error("specialist found error", zap.Error(getEmailErr))
		return spec, domain.ErrInternalServer

	}

	return spec, nil
}
