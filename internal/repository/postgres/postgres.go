package postgres

import (
	"context"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/repository/lib/dsn"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ReadOnlyAccessMode  = "read only"
	ReadWriteAccessMode = "read write"
)

const (
	NoRowsInCollectedSet = "no rows in result set"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg config.Database) (Postgres, error) {
	const op = "repo/postgres/Connect"

	dsn, err := dsn.NewDSN("postgresPool", cfg)

	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	c, _ := pgxpool.ParseConfig(dsn)

	pool, err := pgxpool.NewWithConfig(ctx, c)

	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(ctx)

	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	return Postgres{
		pool: pool,
	}, nil
}

func (conn Postgres) Close() chan error {
	conn.pool.Close()
	ch := make(chan error, 1)
	ch <- nil
	return ch
}

func (conn Postgres) ContextInfo() string {
	return "postgres"
}
