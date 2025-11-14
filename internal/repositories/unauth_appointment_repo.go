package repositories

import (
	"context"
	"fmt"
	"pethelp-backend/internal/core/domain"
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
	animalSizeTableName = "animal_sizes"
	unauthUserTableName = "unauthorized_users_emails"
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

	innerSel := sq.Select("1").From(appointmentsTableName).Where(sq.Eq{"specialist_id": specialistID}).Where(sq.Eq{"appointment_date": date}).Where(sq.Expr("status <> 'pending'")).
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
	locationType, street, unit, apt, description, email, status string,
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
	startTime = startTime.In(loc)
	endTime = endTime.In(loc)

	// Build appointment insert
	appInsert := sq.Insert(profileTable).
		Columns("appointment_date", "location_type", "description", "specialist_id", "amount", "status", "created_at", "updated_at", "start_time", "end_time").
		Values(date, locationType, description, specialistID, amount, status, now, now, startTime, endTime).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)


	sqlAppt, argsAppt, err := appInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building appointment insert: %w", operationUnauthAppointment, err)
	}

	// Address insert
	addrInsert := sq.Insert(addressTable).
		Columns("city_id", "area_id", "street", "unit", "apt").
		Values(cityID, districtID, street, unit, apt).
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

	// Animal size select
	var sizeName string
	var minWeight, maxWeight float32

	sizeSelect := sq.Select("name_eng", "min_weight", "max_weight").
    From(sizeTable).
    Where(sq.Eq{"id": animalSizeID}).
    PlaceholderFormat(sq.Dollar)

	sqlSize, argsSize, err := sizeSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building size select: %w", operationUnauthAppointment, err)
	}

	if err := ar.DBPool.Pool().QueryRow(ctx, sqlSize, argsSize...).Scan(&sizeName, &minWeight, &maxWeight); err != nil {
		return 0, fmt.Errorf("%s querying animal size: %w", operationUnauthAppointment, err)
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

	// Link appointment with service
	svcInsert := sq.Insert(serviceTable).
    Columns("appointment_id", "service_id").
    Values(appointmentID, serviceID).
    Suffix("RETURNING appointment_id"). 
    PlaceholderFormat(sq.Dollar)

	sqlSvc, argsSvc, err := svcInsert.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s building service insert: %w", operationUnauthAppointment, err)
	}

	var returnedID int64
	if err := tx.QueryRow(ctx, sqlSvc, argsSvc...).Scan(&returnedID); err != nil {
		return 0, fmt.Errorf("%s inserting appointment service: %w", operationUnauthAppointment, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s committing transaction: %w", operationUnauthAppointment, err)
	}

	return appointmentID, nil

}


func (ar *UnauthAppointmentRepositoryImpl) GetExpiredAndUnnotified(ctx context.Context) ([]domain.Appointment, error) {

	var appointments []domain.Appointment

	cutoffTime := time.Now().Add(-1 * time.Hour)

	q := sq.Select("id", "appointment_date", "user_id", "specialist_id", "status", "start_time", "end_time").
		From(appointmentsTableName).
		Where(sq.Lt{"end_time": cutoffTime}).
		Where(sq.Eq{"is_expiration_email_sent": false}).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s failed to build SQL query: %w", operationUnauthAppointment, err)
	}

	conn, err := ar.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s failed to take DB pool connection: %w", operationUnauthAppointment, err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%s query failed: %w", operationUnauthAppointment, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var appointmentDate time.Time
		var userID int64
		var specialistID int64
		var status string
		var startTime time.Time
		var endTime time.Time

		if err := rows.Scan(&id, &appointmentDate, &userID, &specialistID, &status, &startTime, &endTime); err != nil {
			return nil, fmt.Errorf("%s scanning row: %w", operationUnauthAppointment, err)
		}

		// append a zero-value appointment for each row (populate fields as needed)
		appointments = append(appointments, domain.Appointment{})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s rows iteration error: %w", operationUnauthAppointment, err)
	}

	if len(appointments) == 0 {
		return nil, domain.ErrNoAppointmentsToExpire
	}

	return appointments, nil
}
	




   