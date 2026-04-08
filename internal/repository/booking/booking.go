package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `INSERT INTO bookings (slot_id, user_id, status, conference_link)
              VALUES ($1, $2, $3, $4)
              RETURNING id, slot_id, user_id, status, conference_link, created_at`

	var b model.Booking
	err := q.QueryRow(ctx, query,
		booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)

	if err != nil {
		fmt.Println(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrBookingExists
		}
		return nil, repository.ErrInternalError
	}

	return &b, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID string) (*model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, slot_id, user_id, status, conference_link, created_at
              FROM bookings WHERE id = $1`

	var b model.Booking
	err := q.QueryRow(ctx, query, bookingID).Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, repository.ErrInternalError
	}

	return &b, nil
}

func (r *BookingRepository) GetAll(ctx context.Context, limit, offset int) ([]model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, slot_id, user_id, status, conference_link, created_at
              FROM bookings
              ORDER BY created_at DESC
              LIMIT $1 OFFSET $2`

	rows, err := q.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, repository.ErrInternalError
	}
	defer rows.Close()

	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, repository.ErrInternalError
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

func (r *BookingRepository) GetBySlotID(ctx context.Context, slotID string) (*model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT id, slot_id, user_id, status, conference_link, created_at
              FROM bookings WHERE slot_id = $1`

	var b model.Booking
	err := q.QueryRow(ctx, query, slotID).Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, repository.ErrInternalError
	}

	return &b, nil
}

func (r *BookingRepository) GetAllByUserID(ctx context.Context, userID string) ([]model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
              FROM bookings b
              JOIN slots s ON s.id = b.slot_id
              WHERE b.user_id = $1 AND s.start_at >= now()
              ORDER BY s.start_at`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, repository.ErrInternalError
	}
	defer rows.Close()

	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, repository.ErrInternalError
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

func (r *BookingRepository) Count(ctx context.Context) (int, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	var count int
	err := q.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&count)
	if err != nil {
		return 0, repository.ErrInternalError
	}

	return count, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, bookingID string) (*model.Booking, error) {
	q := postgres.GetQuerier(ctx, r.pool)

	query := `UPDATE bookings SET status = 'cancelled'
              WHERE id = $1
              RETURNING id, slot_id, user_id, status, conference_link, created_at`

	var b model.Booking
	err := q.QueryRow(ctx, query, bookingID).Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, repository.ErrInternalError
	}

	return &b, nil
}
