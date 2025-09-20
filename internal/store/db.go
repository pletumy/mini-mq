package store

import "database/sql"

func NewDB(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}
