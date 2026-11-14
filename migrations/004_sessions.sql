CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) 
        ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT token, user_id, expires_at FROM sessions ORDER BY created_at DESC LIMIT 1;