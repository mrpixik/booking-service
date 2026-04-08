package schedule

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

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error) {

	q := postgres.GetQuerier(ctx, r.pool)

	query := `INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
              VALUES ($1, $2, $3, $4)
              RETURNING id, room_id, days_of_week, start_time::text, end_time::text`

	var created model.Schedule
	err := q.QueryRow(ctx, query,
		schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		Scan(&created.ID, &created.RoomID, &created.DaysOfWeek, &created.StartTime, &created.EndTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrScheduleExists
		}
		return nil, repository.ErrInternalError
	}

	return &created, nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID string) (*model.Schedule, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, room_id, days_of_week, start_time::text, end_time::text
              FROM schedules WHERE room_id = $1`

	var s model.Schedule
	err := q.QueryRow(ctx, query, roomID).
		Scan(&s.ID, &s.RoomID, &s.DaysOfWeek, &s.StartTime, &s.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, repository.ErrInternalError
	}

	return &s, nil
}
