package db

import "database/sql"

type SQLDB interface {
	DB() *sql.DB
}

type Config struct {
	DSN 		           string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
}