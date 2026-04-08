-- +goose Up
CREATE UNIQUE INDEX unique_active_booking_on_slot
    ON bookings (slot_id)
    WHERE status = 'active';

-- +goose Down
DROP INDEX unique_active_booking_on_slot;