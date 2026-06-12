package app

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string

func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schemaSQL)
	return err
}
