BEGIN;

CREATE TYPE post_status AS ENUM ('created', 'deleted', 'updated', 'removed');

CREATE TABLE IF NOT EXISTS posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
       user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    headline   TEXT        NOT NULL CHECK (char_length(headline) BETWEEN 1 AND 200),
    body       TEXT        NOT NULL CHECK (char_length(body) BETWEEN 1 AND 10000),
    status post_status NOT NULL DEFAULT "created",
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id    ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);

COMMIT;