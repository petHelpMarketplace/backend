package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/pkg/database/postgres"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	curentTableName     = "specialists"
	operationSpecialist = "specialist_repo: "
)

type SpecialistRepositoryImpl struct {
	DBPool *postgres.DB
}

var _ ports.SpecialistRepository = (*SpecialistRepositoryImpl)(nil)

func NewSpecialistRepository(pool *postgres.DB) *SpecialistRepositoryImpl {
	return &SpecialistRepositoryImpl{
		DBPool: pool,
	}
}

func (sr *SpecialistRepositoryImpl) Save(ctx context.Context, name, email, phone, passHash string) (int64, error) {

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("%s failed to time load location: %w", operationSpecialist, err)
		return 0, locErr
	}
	saveTime := time.Now().In(loc)

	query, args, err := sq.Insert(curentTableName).
		Columns(
			"name",
			"email",
			"phone",
			"password_hash",
			"created_at",
			"updated_at",
		).
		Values(
			name,
			email,
			phone,
			passHash,
			saveTime,
			saveTime,
		).
		Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s failed to make insert builder: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return 0, fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to insert data into DB: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return userID, nil
}

func (sr *SpecialistRepositoryImpl) GetByEmail(ctx context.Context, email string) (domain.Specialist, error) {

	var item domain.Specialist
	query, args, err := sq.Select(
		"id",
		"name",
		"email",
		"password_hash",
		"created_at",
	).
		From(curentTableName).
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return item, err
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return item, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return item, fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query, args...)
	err = row.Scan(&item.ID, &item.Name, &item.Email, &item.PasswordHash, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Specialist{}, sql.ErrNoRows
		}
		return item, fmt.Errorf("%s failed to scan data from query row: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return item, fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return item, nil
}

func (sr *SpecialistRepositoryImpl) GetByID(ctx context.Context, id int64) (domain.Specialist, error) {

	var item domain.Specialist
	query, args, err := sq.Select(
		"id",
		"name",
		"email",
		"password_hash",
		"created_at",
	).
		From(curentTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return item, err
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return item, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return item, fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query, args...)
	err = row.Scan(&item.ID, &item.Name, &item.Email, &item.PasswordHash, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Specialist{}, sql.ErrNoRows
		}
		return item, fmt.Errorf("%s failed to scan data from query row: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return item, fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return item, nil
}

func (sr *SpecialistRepositoryImpl) CheckFieldValueExists(ctx context.Context, fieldName string, fieldValue string) (bool, error) {
	if fieldName == "" || fieldValue == "" {
		return false, fmt.Errorf("%s field name or field value cannot be empty", operationSpecialist)
	}

	innerSQL, innerArgs, err := sq.Select("1").
		From(curentTableName).
		Where(sq.Eq{fieldName: fieldValue}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return false, fmt.Errorf("%s failed to build inner query: %w", operationSpecialist, err)
	}

	// Construct the final EXISTS query string
	finalSQL := fmt.Sprintf("SELECT EXISTS (%s)", innerSQL)
	finalArgs := innerArgs

	// conn, err := sr.database.Connection()
	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return false, fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = sr.DBPool.Pool().QueryRow(ctx, finalSQL, finalArgs...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s failed to query data from DB: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return exists, nil
}
