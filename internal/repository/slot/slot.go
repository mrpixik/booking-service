package slot

import (
	"context"
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct {
	pool *pgxpool.Pool
}

func NewSlotRepository(pool *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{pool: pool}
}

// BulkCreate вставляет слоты по списку. Операция идемпотентная,
// это позволяет избежать гонки, если два пользователя захотят одновременно создать слоты одну и ту же дату
func (r *SlotRepository) BulkCreate(ctx context.Context, slots []model.Slot) error {
	q := postgres.GetQuerier(ctx, r.pool)

	for _, s := range slots {
		query := `INSERT INTO slots (room_id, start_at, end_at)
                  VALUES ($1, $2, $3)
                  ON CONFLICT (room_id, start_at) DO NOTHING`
		_, err := q.Exec(ctx, query, s.RoomID, s.Start, s.End)
		if err != nil {
			return repository.ErrInternalError
		}
	}

	return nil
}

// GetAvailableByRoomAndDate возвращает незабронированные слоты
func (r *SlotRepository) GetAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]model.Slot, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT s.id, s.room_id, s.start_at, s.end_at
              FROM slots s
              LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
              WHERE s.room_id = $1
                AND s.start_at >= $2
                AND s.start_at < $3
                AND b.id IS NULL
              ORDER BY s.start_at`

	dayStart := date
	dayEnd := date.AddDate(0, 0, 1)

	rows, err := q.Query(ctx, query, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, repository.ErrInternalError
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var slot model.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, repository.ErrInternalError
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

// GetByID ищет слот по айди
func (r *SlotRepository) GetByID(ctx context.Context, slotID string) (*model.Slot, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, room_id, start_at, end_at
              FROM slots
              WHERE id = $1`

	var slot model.Slot
	err := q.QueryRow(ctx, query, slotID).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrSlotNotFound
		}
		return nil, repository.ErrInternalError
	}
	return &slot, nil
}

func (r *SlotRepository) CountByRoomAndDate(ctx context.Context, roomID string, date time.Time) (int, error) {
	q := postgres.GetQuerier(ctx, r.pool)
	var count int
	err := q.QueryRow(ctx,
		`SELECT COUNT(*) FROM slots WHERE room_id = $1 AND start_at >= $2 AND start_at < $3`,
		roomID, date, date.AddDate(0, 0, 1),
	).Scan(&count)
	if err != nil {
		return 0, repository.ErrInternalError
	}
	return count, nil
}
