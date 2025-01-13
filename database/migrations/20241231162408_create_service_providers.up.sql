CREATE TABLE IF NOT EXISTS service_providers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    has_physical_office BOOLEAN NOT NULL,
    office_coordinates GEOGRAPHY(Point, 4326),
    address VARCHAR(255) NOT NULL,
    mobile_phone_number VARCHAR(13) NOT NULL,
    telephone VARCHAR(15),
    logo_image VARCHAR(255),
    credit NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    average_rating NUMERIC(3, 2) NOT NULL DEFAULT 0.00,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES users(id)
);