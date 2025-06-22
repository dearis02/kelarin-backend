CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	auth_provider SMALLINT NOT NULL,
	role SMALLINT NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	name VARCHAR(255) NOT NULL,
	password VARCHAR(100),
	is_suspended BOOLEAN DEFAULT FALSE,
	suspended_count INT DEFAULT 0,
	suspended_from TIMESTAMPTZ,
	suspended_to TIMESTAMPTZ,
	is_banned BOOLEAN DEFAULT FALSE,
	banned_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION increment_suspended_count() RETURNS TRIGGER AS $$
BEGIN
	UPDATE users SET suspended_count = suspended_count + 1 WHERE id = NEW.id;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER increment_suspended_count_trigger
AFTER UPDATE OF is_suspended ON users
FOR EACH ROW
WHEN (NEW.is_suspended = TRUE)
EXECUTE FUNCTION increment_suspended_count();