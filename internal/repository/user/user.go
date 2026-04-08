package user

import (
	"context"
	"errors"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `INSERT INTO users (email, role, password_hash)
              VALUES ($1, $2, $3)
              RETURNING id, email, role, password_hash, created_at`

	var created model.User
	err := q.QueryRow(ctx, query, user.Email, user.Role, user.PasswordHash).
		Scan(&created.ID, &created.Email, &created.Role, &created.PasswordHash, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":

				return nil, repository.ErrEmailExists
			}
			return nil, repository.ErrInternalError
		}
	}

	return &created, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, email, role, password_hash, created_at
              FROM users WHERE email = $1`
	row := q.QueryRow(ctx, query, email)

	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, repository.ErrInternalError
	}
	return &u, nil
}
