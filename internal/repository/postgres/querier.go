package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type txContextKey struct{}

func GetQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return pool
}
