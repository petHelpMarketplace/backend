package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	// "pethelp-backend/pkg/email"
)

type UnauthAppointmentServiceImpl struct {
	emailSender ports.EmailSender
	//interface to interact with appointment storage 
	unauthAppointmentRepo ports.UnauthAppointmentRepository
	logger         *zap.Logger
	//how long each database call or operation should wait before timing out
	defaultTimeout time.Duration
}

//Interface Check
//it ensures UnauthAppointmentServiceImpl implements the UnauthAppointmentService interface
var _ ports.UnauthAppointmentService = (*UnauthAppointmentServiceImpl)(nil)

//Constructor Function
func NewUnauthAppointmentService(repo ports.UnauthAppointmentRepository, logger *zap.Logger, cfg config.AuthConfig, emailSender ports.EmailSender) *UnauthAppointmentServiceImpl {
	return &UnauthAppointmentServiceImpl{
		unauthAppointmentRepo: repo,
		logger:         logger,
		defaultTimeout: cfg.DefaultTimeout,
		emailSender: emailSender,
	}
}

//saves an unauthenticated appointment, returns appointment ID or error
func (aa *UnauthAppointmentServiceImpl) BookUnauthAppointment(ctx context.Context, unauthAppointment domain.SaveUnauthAppointmentRequest) (int64, error) {

	// Basic validation, ensures start time is before end time
	if !unauthAppointment.StartTime.Before(unauthAppointment.EndTime) {
		aa.logger.Warn("invalid time window",
			zap.Time("start_time", unauthAppointment.StartTime),
			zap.Time("end_time", unauthAppointment.EndTime))
		return 0, domain.ErrInvalidTimeWindow
	}

    //Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, aa.defaultTimeout)
	defer cancel()

	// Check time availability
	exists, err := aa.unauthAppointmentRepo.IsTimeBooked(timeoutCtx, unauthAppointment.SpecialistId, unauthAppointment.Date, unauthAppointment.StartTime, unauthAppointment.EndTime)
	if err != nil {
		aa.logger.Error("database check for existing appointment time failed",
			zap.Int("specialist_id", unauthAppointment.SpecialistId),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}
	if exists {
		aa.logger.Warn("booking attempt of unavailable time",
			zap.Int("specialist_id", unauthAppointment.SpecialistId),
			zap.String("date", unauthAppointment.Date.Format("2006-01-02")),
			zap.Time("start_time", unauthAppointment.StartTime),
			zap.Time("end_time", unauthAppointment.EndTime))
		return 0, domain.ErrTimeUnavailable
	}

	create := domain.SaveUnauthAppointmentRequest{
		ServiceId:    unauthAppointment.ServiceId,
	    CityId:       unauthAppointment.CityId,
		DistrictId:   unauthAppointment.DistrictId,
		Street:  	  unauthAppointment.Street,
		LocationType: unauthAppointment.LocationType,
		Unit:         unauthAppointment.Unit,
		Apt:          unauthAppointment.Apt,
		AnimalSizeId: unauthAppointment.AnimalSizeId,
		Description:  unauthAppointment.Description,
		Date: 		  unauthAppointment.Date,
		StartTime:    unauthAppointment.StartTime,
		EndTime:      unauthAppointment.EndTime,
		Amount:       unauthAppointment.Amount,
		Email:        unauthAppointment.Email,
		SpecialistId: unauthAppointment.SpecialistId,
		Status:       unauthAppointment.Status,
	}
	
	id, err := aa.unauthAppointmentRepo.Save(
		timeoutCtx,
		create.ServiceId,
		create.CityId,
		create.DistrictId,
		create.AnimalSizeId,
		create.SpecialistId,
		create.Amount,
		create.LocationType, 
		create.Street,       
		create.Unit,
		create.Apt,
		create.Description,
		create.Email,
		"pending",       
		create.Date,        
		create.StartTime,
		create.EndTime,
	)
	if err != nil {
		aa.logger.Error("failed to save new appointment to database",
			zap.String("email", unauthAppointment.Email),
			zap.Error(err))
		return 0, domain.ErrInternalServer
	}


	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := aa.emailSender.SendAppointmentConfirmationEmail(bgCtx, unauthAppointment.Email, unauthAppointment.Date, unauthAppointment.StartTime, unauthAppointment.EndTime); err != nil {
			aa.logger.Error("failed to send email", zap.Error(err))
		}
	}()


	return id, nil
}

