package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/entity/urls"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/repository/lib/dsn"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertAliasQuery   = `INSERT INTO urls_alias(url, alias) VALUES($1, $2) RETURNING ID`
	selectURLItemQuery = `SELECT ID, url, alias FROM urls_alias WHERE alias = $1`
	deleteURLQuery     = `DELETE FROM urls_alias WHERE alias = $1`
	updateURLQuery     = `UPDATE urls_alias SET url = $1 WHERE alias = $2`
)

const (
	createUserQuery       = `INSERT INTO users(uuid, username, password) VALUES (uuid_generate_v4(), $1, $2) RETURNING uuid`
	selectUserByUUIDQuery = `SELECT uuid, username, password FROM users WHERE uuid = $1`
	selectUserByUsername  = `SELECT uuid, username, password FROM users WHERE username = $1`
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

	var id int

	for rows.Next() {
		rows.Scan(&id)
	}
	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return -1, fmt.Errorf("%s: %w", op, storageErrors.ErrAliasAlreadyExists)
	}

	return id, nil
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

func (conn Postgres) CreateUser(ctx context.Context, username string, password string) (uuid string, err error) {
	const op = "internal/repo/postgres/CreateUser"

	tx, err := conn.pool.Begin(ctx)

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
			err = fmt.Errorf("transaction finis: %s: %w", op, err)
		}
	}()

	rows, err := tx.Query(ctx, createUserQuery, username, password)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var id string

	for rows.Next() {
		rows.Scan(&id)
	}
	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return "", fmt.Errorf("%s: %w", op, storageErrors.ErrUsernameAlreadyExists)
	}

	return id, nil
}

func (conn Postgres) GetUserByUUID(ctx context.Context, uuid string) (users.User, error) {
	const op = "internal/repo/postgres/CreateUser"

	tx, err := conn.pool.Begin(ctx)

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("transaction finis: %s: %w", op, err)
		}
	}()

	rows, err := tx.Query(ctx, selectUserByUUIDQuery, uuid)

	if err != nil {
		if err == pgx.ErrNoRows {
			return users.User{}, fmt.Errorf("%s: %w", op, storageErrors.ErrAliasNotFound)
		}

		return users.User{}, fmt.Errorf("%s:%v", op, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[users.User])

	if err != nil {
		if err.Error() == "no rows in result set" {
			return users.User{}, fmt.Errorf("%s: %w", op, generalerrors.ErrUserNotFound)
		}

		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (conn Postgres) GetUserByUsername(ctx context.Context, username string) (users.User, error) {
	const op = "internal/repository/postgres/GetUserByUsername"

	tx, err := conn.pool.Begin(ctx)

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		var e error

		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("transaction finis: %s: %w", op, err)
		}
	}()

	rows, err := tx.Query(ctx, selectUserByUsername, username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.User{}, fmt.Errorf("%s: %w", op, generalerrors.ErrUserNotFound)
		}

		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[users.User])

	if err != nil {
		if err.Error() == "no rows in result set" {
			return users.User{}, fmt.Errorf("%s: %w", op, generalerrors.ErrUserNotFound)
		}
	}

	return user, nil
}
