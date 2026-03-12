CREATE TABLE IF NOT EXISTS users (
    id            bigserial    PRIMARY KEY,
    name          text         NOT NULL,
    email         text UNIQUE  NOT NULL,
    password_hash bytea        NOT NULL,
    activated     boolean      NOT NULL DEFAULT false,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    version       integer      NOT NULL DEFAULT 1
);
