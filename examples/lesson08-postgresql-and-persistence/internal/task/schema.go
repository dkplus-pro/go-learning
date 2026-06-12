package task

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string

// Migrate 执行本 demo 的最小 schema。
// 真实项目通常会使用专门的迁移工具，本课先保持可读和可运行。
func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schemaSQL)
	return err
}
