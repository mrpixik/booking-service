package room

import (
	"context"
	"errors"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) GetAll(ctx context.Context) ([]model.Room, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, name, description, capacity, created_at FROM rooms`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, repository.ErrInternalError
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, repository.ErrInternalError
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *RoomRepository) GetById(ctx context.Context, roomID string) (*model.Room, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, name, description, capacity, created_at
              FROM rooms WHERE id = $1`
	row := q.QueryRow(ctx, query, roomID)

	var u model.Room
	err := row.Scan(&u.ID, &u.Name, &u.Description, &u.Capacity, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrRoomNotFound
		}
		return nil, repository.ErrInternalError
	}
	return &u, nil
}

func (r *RoomRepository) Create(ctx context.Context, room *model.Room) (*model.Room, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `INSERT INTO rooms (name, description, capacity)
              VALUES ($1, $2, $3)
              RETURNING id, name, description, capacity, created_at`

	var created model.Room
	err := q.QueryRow(ctx, query, room.Name, room.Description, room.Capacity).
		Scan(&created.ID, &created.Name, &created.Description, &created.Capacity, &created.CreatedAt)
	if err != nil {
		return nil, repository.ErrInternalError
	}

	return &created, nil
}
