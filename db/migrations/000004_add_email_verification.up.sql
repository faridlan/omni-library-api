ALTER TABLE users
ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN verification_token VARCHAR(255),
ADD COLUMN verification_expires_at TIMESTAMPTZ;