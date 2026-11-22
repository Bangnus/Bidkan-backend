-- 21_email_verification.sql
-- Track email verification state for email/password rider accounts

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users
SET email_verified = TRUE
WHERE email IS NOT NULL AND status = 'active';
