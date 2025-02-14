package postgres

import (
	"context"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/domain/urls"
	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/repository/lib/dsn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertAliasQuery = `INSERT INTO urls_alias(url, alias) VALUES($1, $2) RETURNING id`

	selectIdQuery = `SELECT id FROM urls_alias WHERE alias = $1`

	selectURLItemQuery = `SELECT id, url, alias FROM urls_alias WHERE alias = $1`

	deleteURLQuery = `DELETE FROM urls_alias WHERE alias = $1`
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg config.Database) (Postgres, error) {
	const path = "repo/postgres/Connect"

	dsn := dsn.NewDsn("postgresPool", cfg)

	c, _ := pgxpool.ParseConfig(dsn)

	pool, err := pgxpool.NewWithConfig(ctx, c)

	if err != nil {
		return Postgres{}, fmt.Errorf("%s:%w", path, err)
	}

	err = pool.Ping(ctx)

	if err != nil {
		return Postgres{}, fmt.Errorf("%s:%w", path, err)
	}

	return Postgres{
		pool: pool,
	}, nil
}

func (conn Postgres) SaveAlias(ctx context.Context, url, alias string) (int, error) {
	const path = "repository/postgres/SaveAlias"

	rows, err := conn.pool.Query(ctx, insertAliasQuery, url, alias)

	if err != nil {
		return -1, fmt.Errorf("%s:%w", path, err)
	}
	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return -1, fmt.Errorf("%s:%w", path, storageErrors.ErrAliasAlreadyExists)
	}

	rows, err = conn.pool.Query(ctx, selectIdQuery, alias)

	if err != nil {
		return -1, fmt.Errorf("%s:%w", path, storageErrors.ErrAliasAlreadyExists)
	}

	var urlItem urls.URL

	for rows.Next() {
		rows.Scan(&urlItem.Id)
	}

	return urlItem.Id, nil
}

func (conn Postgres) GetURL(ctx context.Context, alias string) (string, error) {
	const path = "repo/postgresql/postgres.go.GetURL"

	rows, err := conn.pool.Query(ctx, selectURLItemQuery, alias)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("%s:%w", path, storageErrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s:%v", path, err)
	}
	defer rows.Close()

	urlItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[urls.URL])

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", fmt.Errorf("%s:%w", path, storageErrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s:%v", path, err)
	}

	return urlItem.URL, nil
}

func (conn Postgres) DeleteURL(ctx context.Context, alias string) error {
	const path = "repo/postgresql/postgres.go.DeleteURL"

	rows, err := conn.pool.Query(ctx, deleteURLQuery, alias)

	if err != nil {
		return fmt.Errorf("%s:%v", path, err)
	}
	rows.Close()

	return nil
}
