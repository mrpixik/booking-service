-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
                       id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       email         VARCHAR(255) NOT NULL UNIQUE,
                       role          VARCHAR(10)  NOT NULL CHECK (role IN ('admin', 'user')),
                       password_hash TEXT,
                       created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE rooms (
                       id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       name        VARCHAR(255) NOT NULL,
                       description TEXT,
                       capacity    INT,
                       created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE schedules (
                           id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           room_id      UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
                           days_of_week INT[] NOT NULL,
                           start_time   TIME NOT NULL,
                           end_time     TIME NOT NULL,
                           CHECK (start_time < end_time)
);

CREATE TABLE slots (
                       id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       room_id  UUID        NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
                       start_at TIMESTAMPTZ NOT NULL,
                       end_at   TIMESTAMPTZ NOT NULL,
                       UNIQUE (room_id, start_at)
);

CREATE INDEX idx_slots_room_date ON slots (room_id, start_at);

CREATE TABLE bookings (
                          id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          slot_id         UUID        NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
                          user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
                          conference_link TEXT,
                          created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_bookings_active_slot ON bookings (slot_id) WHERE status = 'active';
CREATE INDEX idx_bookings_user_id ON bookings (user_id);

-- +goose Down
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";