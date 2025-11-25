CREATE TABLE IF NOT EXISTS refresh_sessions (
  jti TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  user_agent TEXT,
  ip_address TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT fk_user
    FOREIGN KEY(user_id)
    REFERENCES users(id) ON DELETE CASCADE
);

