ALTER TABLE users ADD COLUMN verification_token VARCHAR(64);

CREATE INDEX idx_users_verification_token ON users(verification_token);