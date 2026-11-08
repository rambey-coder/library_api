CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    first_name VARCHAR NOT NULL,
    last_name  VARCHAR NOT NULL,
    gender VARCHAR CHECK (gender IN ('male', 'female')),
    phone VARCHAR,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('admin', 'member', 'librarian')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

ALTER TABLE members ADD COLUMN user_id INT UNIQUE REFERENCES users(id);

CREATE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL;