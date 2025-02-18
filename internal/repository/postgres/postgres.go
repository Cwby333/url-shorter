package postgres

import (
	"context"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/entity/urls"
	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/repository/lib/dsn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertAliasQuery = `INSERT INTO urls_alias(url, alias) VALUES($1, $2) RETURNING ID`

	selectIDQuery = `SELECT ID FROM urls_alias WHERE alias = $1`

	selectURLItemQuery = `SELECT ID, url, alias FROM urls_alias WHERE alias = $1`

	deleteURLQuery = `DELETE FROM urls_alias WHERE alias = $1`

	updateURLQuery = `UPDATE urls_alias SET url = $1 WHERE alias = $2`
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg config.Database) (Postgres, error) {
	const op = "repo/postgres/Connect"

	dsn := dsn.NewDsn("postgresPool", cfg)

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

func (conn Postgres) Close() {
	conn.pool.Close()
}

func (conn Postgres) SaveAlias(ctx context.Context, url, alias string) (int, error) {
	const op = "repository/postgres/SaveAlias"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})

	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("%s:finishing transaction:  %w", op, e)
		}
	}()

	rows, err := tx.Query(ctx, insertAliasQuery, url, alias)

	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return -1, fmt.Errorf("%s: %w", op, storageErrors.ErrAliasAlreadyExists)
	}

	rows, err = tx.Query(ctx, selectIDQuery, alias)

	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, storageErrors.ErrAliasAlreadyExists)
	}

	var urlItem urls.URL

	for rows.Next() {
		rows.Scan(&urlItem.ID)
	}

	return urlItem.ID, nil
}

func (conn Postgres) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "repo/postgresql/postgres.go.GetURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("%s:finishing transaction:  %w", op, e)
		}
	}()

	rows, err := tx.Query(ctx, selectURLItemQuery, alias)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storageErrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s:%v", op, err)
	}
	defer rows.Close()

	urlItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[urls.URL])

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", fmt.Errorf("%s: %w", op, storageErrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s:%v", op, err)
	}

	return urlItem.URL, nil
}

func (conn Postgres) DeleteURL(ctx context.Context, alias string) error {
	const op = "repo/postgresql/postgres.go.DeleteURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})

	if err != nil {
		return fmt.Errorf("%s:%v", op, err)
	}

	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("%s:finishing transaction: %w", op, e)
		}
	}()

	rows, err := tx.Query(ctx, deleteURLQuery, alias)

	if err != nil {
		return fmt.Errorf("%s:%v", op, err)
	}
	rows.Close()

	return nil
}

func (conn Postgres) UpdateURL(ctx context.Context, newURL, alias string) error {
	const op = "repo/postgresql/postgres.go.DeleteURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("%s:finishing transaction: %w", op, e)
		}
	}()

	rows, err := tx.Query(ctx, updateURLQuery, newURL, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return fmt.Errorf("%s: %w", op, storageErrors.ErrAliasNotFound)
	}

	return nil
}
