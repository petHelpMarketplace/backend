package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/pkg/database/postgres"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	operationEmailSender = "email_repo: "
)

type EmailRepositoryImpl struct {
	DBPool *postgres.DB
}

var _ ports.EmailRepository = (*EmailRepositoryImpl)(nil)

func NewEmailRepository(pool *postgres.DB) *EmailRepositoryImpl {
	return &EmailRepositoryImpl{
		DBPool: pool,
	}
}

func(e *EmailRepositoryImpl) MarkEmailSent(ctx context.Context, ID int64) (domain.Appointment, error) {

	var updatedAppointment domain.Appointment
	
	q := sq.Update(appointmentsTableName).Set("is_expiration_email_sent", true).Where(sq.Eq{"id": ID}).Suffix("RETURNING *").PlaceholderFormat(sq.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return updatedAppointment, fmt.Errorf("%s building expiration email update: %w", operationEmailSender, err)
	 }

	conn, err := e.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return updatedAppointment, fmt.Errorf("%s failed to take DB pool connection: %w", operationEmailSender, err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return updatedAppointment, fmt.Errorf("%s failed to begin transaction: %w", operationEmailSender, err)
	}
	defer tx.Rollback(ctx)


	res, err := tx.Query(ctx, query, args...)

	if err != nil {
		return updatedAppointment, fmt.Errorf("%s failed to execute update profile query: %w", operationEmailSender, err)
	}

	updatedAppointment, err = pgx.CollectOneRow(res, pgx.RowToStructByName[domain.Appointment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updatedAppointment, sql.ErrNoRows
		}
		return updatedAppointment, fmt.Errorf("%s failed to scan returned data from update: %w", operationEmailSender, err)
	}

	return updatedAppointment, tx.Commit(ctx)
}