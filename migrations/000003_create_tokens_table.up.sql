CREATE TABLE IF NOT EXISTS tokens (
    hash    bytea        PRIMARY KEY,
    user_id bigint       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry  timestamptz  NOT NULL,
    scope   text         NOT NULL
);
