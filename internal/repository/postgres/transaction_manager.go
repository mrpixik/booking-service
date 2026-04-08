package postgres

import (
	"context"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) Begin(parent context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(parent)
	if err != nil {
		return repository.ErrInternalError
	}

	ctx := context.WithValue(parent, txContextKey{}, tx)

	if err = fn(ctx); err != nil {
		_ = tx.Rollback(parent)
		return err
	}
	return tx.Commit(parent)
}
