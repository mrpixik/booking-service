package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MAX_RETRIES = 5
	BASE_DELAY  = 100 * time.Millisecond
)

// Решил вывести подключение к бд в отдельную функцию, чтобы в дальнейшем, при горизонтальном масштабировании было проще
// Теперь чтобы создавать репозитории, буду просто передавать туда *pgxpool.Pool, который сам будет управлять подключениями.
func ConnectPostgres(ctx context.Context, cfg config.PostgresStorage, env string) (*pgxpool.Pool, error) {
	const op = "repository.postgres.ConnectPostgres"

	pgxCfg, err := parseConfig(cfg, env)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for i := 0; i <= MAX_RETRIES; i++ {
		err := conn.Ping(ctx)

		if err != nil && i == MAX_RETRIES {
			return nil, fmt.Errorf("%s: %w", op, err)
		} else if err != nil {
			//exponential backoff
			backoff := time.Duration(1<<i) * BASE_DELAY
			time.Sleep(backoff)
		} else {
			break
		}
	}
	return conn, nil
}

func parseConfig(cfg config.PostgresStorage, env string) (*pgxpool.Config, error) {
	const op = "repository.postgres.parseConfig"

	if env == "local" {
		cfg.Host = "localhost"
	}

	storageDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)

	pgxCfg, err := pgxpool.ParseConfig(storageDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pgxCfg.MaxConns = cfg.MaxConns
	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifeTime

	return pgxCfg, nil
}
