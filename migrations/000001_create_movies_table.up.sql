CREATE TABLE IF NOT EXISTS movies (
    id         bigserial    PRIMARY KEY,
    title      text         NOT NULL,
    year       integer      NOT NULL,
    runtime    integer      NOT NULL,
    genres     text[]       NOT NULL,
    created_at timestamptz  NOT NULL DEFAULT now(),
    version    integer      NOT NULL DEFAULT 1
);
