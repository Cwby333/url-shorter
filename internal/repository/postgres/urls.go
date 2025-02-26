package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/entity/urls"
	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/jackc/pgx/v5"
)

const (
	insertAliasQuery   = `INSERT INTO urls_alias(url, alias) VALUES($1, $2) RETURNING ID`
	selectURLItemQuery = `SELECT ID, url, alias FROM urls_alias WHERE alias = $1`
	deleteURLQuery     = `DELETE FROM urls_alias WHERE alias = $1`
	updateURLQuery     = `UPDATE urls_alias SET url = $1 WHERE alias = $2`
)

func (conn Postgres) SaveAlias(ctx context.Context, url, alias string) (id int, err error) {
	const op = "repository/postgres/SaveAlias"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: ReadWriteAccessMode})

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

	for rows.Next() {
		err = rows.Scan(&id)

		if err != nil {
			return -1, fmt.Errorf("%s: %w", op, err)
		}
	}

	rows.Close()

	if rows.CommandTag().RowsAffected() < 1 {
		return -1, fmt.Errorf("%s: %w", op, generalerrors.ErrAliasAlreadyExists)
	}

	return id, nil
}

func (conn Postgres) GetURL(ctx context.Context, alias string) (url string, err error) {
	const op = "repo/postgresql/postgres.go.GetURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: ReadOnlyAccessMode})

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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, generalerrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	urlItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[urls.URL])

	if err != nil {
		if err.Error() == NoRowsInCollectedSet {
			return "", fmt.Errorf("%s: %w", op, generalerrors.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return urlItem.URL, nil
}

func (conn Postgres) DeleteURL(ctx context.Context, alias string) (err error) {
	const op = "repo/postgresql/postgres.go.DeleteURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: ReadWriteAccessMode})

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

	rows, err := tx.Query(ctx, deleteURLQuery, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows.Close()

	return nil
}

func (conn Postgres) UpdateURL(ctx context.Context, newURL, alias string) (err error) {
	const op = "repo/postgresql/postgres.go.DeleteURL"
	tx, err := conn.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: ReadWriteAccessMode})

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
		return fmt.Errorf("%s: %w", op, generalerrors.ErrAliasNotFound)
	}

	return nil
}
