package dbo

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

//go:generate sqlc generate
//go:embed migrations
var migrations embed.FS

func Dial(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)

	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	if err := applySchema(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return pool, nil
}

func applySchema(ctx context.Context, conn *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(conn)
	defer db.Close()

	_, err := migrate.ExecContext(ctx, db, "postgres", &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}, migrate.Up)
	return err
}
