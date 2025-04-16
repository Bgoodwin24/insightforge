-- +goose Up
CREATE TABLE IF NOT EXISTS pending_users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    expires_at TIMESTAMP NOT NULL DEFAULT now() + INTERVAL '1 day'
);

-- +goose Down
DROP TABLE IF EXISTS pending_users;
