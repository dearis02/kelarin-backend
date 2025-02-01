DO $$
BEGIN
    CREATE TYPE service_delivery_method AS ENUM (
        'onsite', 
        'online'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'service_delivery_method type already exists';
END $$;

CREATE TABLE IF NOT EXISTS services(
    id UUID PRIMARY KEY,
    service_provider_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    delivery_methods service_delivery_method[] NOT NULL,
    fee_start_at DECIMAL(15, 2) NOT NULL,
    fee_end_at DECIMAL(15, 2) NOT NULL,
    images VARCHAR(255)[] NOT NULL,
    rules JSONB NOT NULL DEFAULT '{}',
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (service_provider_id) REFERENCES service_providers(id)
);

CREATE TABLE IF NOT EXISTS service_categories(
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service_service_categories(
    service_id UUID NOT NULL,
    service_category_id UUID NOT NULL,
    PRIMARY KEY (service_id, service_category_id),
    FOREIGN KEY (service_id) REFERENCES services(id),
    FOREIGN KEY (service_category_id) REFERENCES service_categories(id)
);