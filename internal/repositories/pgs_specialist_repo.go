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
	"github.com/lib/pq"
)

const (
	currentTableName           = "specialists"
	serviceSpecialistTableName = "specialist_services"
	operationSpecialist        = "specialist_repo: "
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

	query, args, err := sq.Insert(currentTableName).
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

	query, args, err := sq.Select("s.*", "ca.area_name").
		From("specialists s").
		LeftJoin("city_areas ca ON s.city_area_id = ca.id").
		Where(sq.Eq{"s.email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return item, fmt.Errorf("%s failed to build query: %w", operationSpecialist, err)
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

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return item, fmt.Errorf("%s failed to query data from DB: %w", operationSpecialist, err)
	}
	defer rows.Close()

	item, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Specialist])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	query, args, err := sq.Select("s.*", "ca.area_name").
		From("specialists s").
		LeftJoin("city_areas ca ON s.city_area_id = ca.id").
		Where(sq.Eq{"s.id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return item, fmt.Errorf("%s failed to create new select builder: %w", operationSpecialist, err)
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

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return item, fmt.Errorf("%s failed to query data from DB: %w", operationSpecialist, err)
	}
	defer rows.Close()

	item, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Specialist])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
		From(currentTableName).
		Where(sq.Eq{fieldName: fieldValue}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return false, fmt.Errorf("%s failed to build inner query: %w", operationSpecialist, err)
	}

	// Construct the final EXISTS query string
	finalSQL := fmt.Sprintf("SELECT EXISTS (%s)", innerSQL)
	finalArgs := innerArgs

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
	err = tx.QueryRow(ctx, finalSQL, finalArgs...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s failed to query data from DB: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return exists, nil
}

func (sr *SpecialistRepositoryImpl) UpdatePasswordHash(ctx context.Context, id int64, newHash string) error {

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("%s failed to time load location: %w", operationSpecialist, err)
		return locErr
	}
	updateTime := time.Now().In(loc)

	query, args, err := sq.Update(currentTableName).
		Set("password_hash", newHash).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s failed to build update query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute update query: %w", operationSpecialist, err)
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}
	return nil
}

func (sr *SpecialistRepositoryImpl) UpdateAvatar(ctx context.Context, id int64, avatarURL string) error {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("%s failed to time load location: %w", operationSpecialist, err)
		return locErr
	}
	updateTime := time.Now().In(loc)

	query, args, err := sq.Update(currentTableName).
		Set("avatar", avatarURL).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s failed to build update avatar query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute update avatar query: %w", operationSpecialist, err)
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit(ctx)
}

func (sr *SpecialistRepositoryImpl) UpdateProfile(ctx context.Context, id int64, req domain.SpecialistProfUpdateReq) (domain.Specialist, error) {

	var updatedSpecialist domain.Specialist

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("%s failed to time load location: %w", operationSpecialist, err)
		return updatedSpecialist, locErr
	}
	updateTime := time.Now().In(loc)

	builder := sq.Update(currentTableName).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": id})

	if req.Name != nil {
		builder = builder.Set("name", *req.Name)
	}
	if req.FamilyName != nil {
		builder = builder.Set("family_name", *req.FamilyName)
	}
	if req.Phone != nil {
		builder = builder.Set("phone", *req.Phone)
	}
	if req.District != nil {
		districtSubQuery := sq.Select("id").
			From("city_areas").
			Where(sq.Eq{"area_name": *req.District})

		builder = builder.Set("city_area_id", districtSubQuery)
	}
	if req.Experience != nil {
		builder = builder.Set("experience", *req.Experience)
	}
	if req.Bio != nil {
		builder = builder.Set("bio", *req.Bio)
	}

	query, args, err := builder.
		Suffix("RETURNING *, (SELECT area_name FROM city_areas WHERE id = specialists.city_area_id) AS area_name").
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return updatedSpecialist, fmt.Errorf("%s failed to build update profile query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return updatedSpecialist, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return updatedSpecialist, fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return updatedSpecialist, fmt.Errorf("%s failed to execute update profile query: %w", operationSpecialist, err)
	}

	updatedSpecialist, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Specialist])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updatedSpecialist, sql.ErrNoRows
		}
		return updatedSpecialist, fmt.Errorf("%s failed to scan returned data from update: %w", operationSpecialist, err)
	}

	return updatedSpecialist, tx.Commit(ctx)
}

func (sr *SpecialistRepositoryImpl) SearchSpecialistByServicePetArea(ctx context.Context, specialist domain.SearchSpecialistParams, limit, offset int) ([]domain.Specialist, error) {

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	areaId := specialist.Area
	serviceId := specialist.Service

	var items []domain.Specialist
	builder := sq.Select("sp.*").
		From(currentTableName + " AS sp").
		Join(serviceSpecialistTableName + " AS s ON s.specialist_id = sp.id").
		Join(addressTableName + " AS adr ON adr.id = sp.addresses_id").
		PlaceholderFormat(sq.Dollar).Limit(uint64(limit)).Offset(uint64(offset))

	conds := make([]sq.Sqlizer, 0)
	if serviceId != 0 {
		conds = append(conds, sq.Eq{"s.service_id": serviceId})
	}
	if areaId != 0 {
		conds = append(conds, sq.Eq{"adr.area_id": areaId})
	}
	if len(conds) > 0 {
		return nil, fmt.Errorf("%s: no filters provided", operationSpecialist)
	}

	builder = builder.Where(sq.And(conds))

	query, args, err := builder.ToSql()

	if err != nil {
		return items, fmt.Errorf("%s failed to create new select builder: %w", operationSpecialist, err)
	}
	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return items, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return items, fmt.Errorf("%s failed to query data from DB: %w", operationSpecialist, err)
	}
	defer rows.Close()

	items, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.Specialist])
	if err != nil {
		return nil, fmt.Errorf("%s: scan: %w", operationSpecialist, err)
	}

	return items, nil
}

func (sr *SpecialistRepositoryImpl) AddImages(ctx context.Context, specialistID int64, imageURLs []string) error {
	if len(imageURLs) == 0 {
		return nil // Nothing to add
	}

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("%s failed to load time location: %w", operationSpecialist, err)
	}
	updateTime := time.Now().In(loc)

	// Use array_cat to append the new URLs to the existing array.
	// We use squirrel's Expr for custom SQL functions.
	query, args, err := sq.Update(currentTableName).
		Set("image_id", sq.Expr("array_cat(COALESCE(image_id, '{}'), ?)", pq.Array(imageURLs))).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": specialistID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s failed to build add images query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute add images query: %w", operationSpecialist, err)
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows // No specialist found with that ID
	}

	return nil
}

func (sr *SpecialistRepositoryImpl) DeleteImage(ctx context.Context, specialistID int64, imageURL string) error {
	// Return early if there's nothing to delete.
	if imageURL == "" {
		return nil
	}

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("%s failed to load time location: %w", operationSpecialist, err)
	}
	updateTime := time.Now().In(loc)

	// Use the simple and efficient ARRAY_REMOVE function for a single element.
	query, args, err := sq.Update(currentTableName).
		Set("image_id", sq.Expr("ARRAY_REMOVE(image_id, ?)", imageURL)).
		Set("updated_at", updateTime).
		Where(sq.And{
			sq.Eq{"id": specialistID},
			sq.Expr("? = ANY(image_id)", imageURL),
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s failed to build delete image query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute delete image query: %w", operationSpecialist, err)
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows //No specialist or image not present
	}

	return nil
}

func (sr *SpecialistRepositoryImpl) UpdateIsActive(ctx context.Context, id int64, isActive bool) error {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("%s failed to time load location: %w", operationSpecialist, err)
		return locErr
	}
	updateTime := time.Now().In(loc)

	query, args, err := sq.Update(currentTableName).
		Set("is_active", isActive).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s failed to build update active state query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute update active state query: %w", operationSpecialist, err)
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit(ctx)
}

// MarkAsDeleted marks the specialist profile as deleted (Soft Delete) within a transaction.
func (sr *SpecialistRepositoryImpl) MarkAsDeleted(ctx context.Context, id int64) error {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("%s failed to load time location: %w", operationSpecialist, err)
	}
	updateTime := time.Now().In(loc)

	// Build the update query
	query, args, err := sq.Update(currentTableName).
		Set("is_deleted", true).
		Set("is_active", false). // Deactivate profile immediately
		Set("delete_initiated_at", updateTime).
		Set("updated_at", updateTime).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s failed to build mark as deleted query: %w", operationSpecialist, err)
	}

	// Acquire connection from the pool
	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin sql transaction: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	// Execute the query within the transaction
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute mark as deleted query: %w", operationSpecialist, err)
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit sql transaction: %w", operationSpecialist, err)
	}

	return nil
}

// GetExpiredAccounts retrieves specialists marked for deletion before the threshold time.
func (sr *SpecialistRepositoryImpl) GetExpiredAccounts(ctx context.Context, thresholdTime time.Time) ([]domain.Specialist, error) {
	var items []domain.Specialist

	query, args, err := sq.Select("*").
		From(currentTableName).
		Where(sq.And{
			sq.Eq{"is_deleted": true},
			sq.Lt{"delete_initiated_at": thresholdTime},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s failed to build select expired query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("%s failed to begin tx: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s failed to query expired accounts: %w", operationSpecialist, err)
	}
	defer rows.Close()

	items, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.Specialist])
	if err != nil {
		return nil, fmt.Errorf("%s failed to scan expired accounts: %w", operationSpecialist, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s failed to commit tx: %w", operationSpecialist, err)
	}

	return items, nil
}

// DeleteAllServices removes all service associations for a specialist.
func (sr *SpecialistRepositoryImpl) DeleteAllServices(ctx context.Context, specialistID int64) error {
	query, args, err := sq.Delete(serviceSpecialistTableName).
		Where(sq.Eq{"specialist_id": specialistID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s failed to build delete services query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin tx: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute delete services: %w", operationSpecialist, err)
	}

	return tx.Commit(ctx)
}

// HardDelete permanently removes a specialist record by ID.
func (sr *SpecialistRepositoryImpl) HardDelete(ctx context.Context, id int64) error {
	query, args, err := sq.Delete(currentTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s failed to build hard delete query: %w", operationSpecialist, err)
	}

	conn, err := sr.DBPool.Pool().Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to take DB pool connection: %w", operationSpecialist, err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("%s failed to begin tx: %w", operationSpecialist, err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s failed to execute hard delete: %w", operationSpecialist, err)
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit(ctx)
}
