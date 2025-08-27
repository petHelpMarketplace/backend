package repositories

import (
	"context"
	"fmt"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/pkg/database/postgres"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	unauthAppointmentTableName    = "appointment_services"
	operationUnauthAppointment = "unauth_appointment_repo: "
	appointmentsTableName = "appointments"
	addressTableName = "addresses"
	animalSizeTableName = "animal_size"
	unauthUserTableName = "unauthorized_user_email"
)

type UnauthAppointmentRepositoryImpl struct {
	DBPool *postgres.DB
}

var _ ports.UnauthAppointmentRepository = (*UnauthAppointmentRepositoryImpl)(nil)

func NewUnauthAppointmentRepository(pool *postgres.DB) *UnauthAppointmentRepositoryImpl {
	return &UnauthAppointmentRepositoryImpl{
		DBPool: pool,
	}
}

// IsTimeBooked checks if the specialist has an appointment overlapping with the given time window.
func (ar *UnauthAppointmentRepositoryImpl) IsTimeBooked(ctx context.Context, specialistID int,date, startTime, endTime time.Time) (bool, error) {
	if !startTime.Before(endTime) {
		return false, fmt.Errorf("invalid time window: start must be before end")
	}

	innerSel := sq.Select("1").From(appointmentsTableName).Where(sq.Eq{"specialist_id": specialistID}).Where(sq.Expr("status <> 'canceled'")).
		Where(sq.Expr("(start_time, end_time) OVERLAPS (?, ?)", startTime, endTime)).PlaceholderFormat(sq.Dollar)

	innerSQL, innerArgs, err := innerSel.ToSql()
	if err != nil {
		return false, fmt.Errorf("build inner exists query: %w", err)
	}

	finalSQL := "SELECT EXISTS (" + innerSQL + ")"

	var exists bool
	if err := ar.DBPool.Pool().QueryRow(ctx, finalSQL, innerArgs...).Scan(&exists); err != nil {
		return false, fmt.Errorf("query exists: %w", err)
	}

	return exists, nil
}

func (ar *UnauthAppointmentRepositoryImpl) Save(ctx context.Context,
	serviceID, cityID, districtID, animalSizeID, specialistID int,
	amount float32,
	locationType, street, unit, apt, description, email string,
	date, startTime, endTime time.Time) (int64, error) {

	const (
		profileTable  = appointmentsTableName
		addressTable  = addressTableName
		emailTable    = unauthUserTableName
		sizeTable     = animalSizeTableName
		serviceTable  = unauthAppointmentTableName
	)

	// Load timezone
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return 0, fmt.Errorf("%s failed to load time location: %w", operationUnauthAppointment, err)
	}
	now := time.Now().In(loc)

	// Build appointment insert
	appInsert := sq.Insert(profileTable).
		Columns("appointment_date", "location_type", "description", "specialist_id", "amount", "status", "created_at", "updated_at", "start_time", "end_time").
		Values(date, locationType, description, specialistID, amount, "pending", now, now, startTime, endTime).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlAppt, argsAppt, err := appInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building appointment insert: %w", operationUnauthAppointment, err)
	}

	// Address insert
	addrInsert := sq.Insert(addressTable).
		Columns("city_id", "area_id", "street", "unit", "apt", "created_at").
		Values(cityID, districtID, street, unit, apt, now).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlAddr, argsAddr, err := addrInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building address insert: %w", operationUnauthAppointment, err)
	}

	// Email insert
	emailInsert := sq.Insert(emailTable).
		Columns("email", "created_at").
		Values(email, now).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlEmail, argsEmail, err := emailInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building email insert: %w", operationUnauthAppointment, err)
	}

	// Animal size insert
	sizeInsert := sq.Insert(sizeTable).
		Columns("size_id", "created_at"). // Assuming size_id is correct
		Values(animalSizeID, now).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlSize, argsSize, err := sizeInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building size insert: %w", operationUnauthAppointment, err)
	}

	// Acquire DB connection
	conn, err := ar.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s failed to acquire DB connection: %w", operationUnauthAppointment, err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return 0, fmt.Errorf("%s failed to begin transaction: %w", operationUnauthAppointment, err)
	}
	defer tx.Rollback(ctx)

	// Execute inserts
	var appointmentID int64
	if err := tx.QueryRow(ctx, sqlAppt, argsAppt...).Scan(&appointmentID); err != nil {
		return 0, fmt.Errorf("%s inserting appointment: %w", operationUnauthAppointment, err)
	}

	var addressID int64
	if err := tx.QueryRow(ctx, sqlAddr, argsAddr...).Scan(&addressID); err != nil {
		return 0, fmt.Errorf("%s inserting address: %w", operationUnauthAppointment, err)
	}

	var emailID int64
	if err := tx.QueryRow(ctx, sqlEmail, argsEmail...).Scan(&emailID); err != nil {
		return 0, fmt.Errorf("%s inserting email: %w", operationUnauthAppointment, err)
	}

	var sizeID int64
	if err := tx.QueryRow(ctx, sqlSize, argsSize...).Scan(&sizeID); err != nil {
		return 0, fmt.Errorf("%s inserting animal size: %w", operationUnauthAppointment, err)
	}

	// Link appointment with service
	svcInsert := sq.Insert(serviceTable).
		Columns("appointment_id", "service_id").
		Values(appointmentID, serviceID).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlSvc, argsSvc, err := svcInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building service insert: %w", operationUnauthAppointment, err)
	}

	var svcID int64
	if err := tx.QueryRow(ctx, sqlSvc, argsSvc...).Scan(&svcID); err != nil {
		return 0, fmt.Errorf("%s inserting appointment service: %w", operationUnauthAppointment, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s committing transaction: %w", operationUnauthAppointment, err)
	}

	return appointmentID, nil
}




   