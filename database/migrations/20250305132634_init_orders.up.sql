DO $$
BEGIN
    CREATE TYPE order_status AS ENUM (
        'pending',
        'ongoing',
        'finished'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'order_status type already exists';
END $$;

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    service_provider_id UUID NOT NULL,
    offer_id UUID NOT NULL,
    payment_id UUID,
    payment_fulfilled BOOLEAN DEFAULT FALSE,
    service_fee DECIMAL(15, 2) NOT NULL,
    service_date DATE NOT NULL,
    service_time TIMETZ NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (service_provider_id) REFERENCES service_providers(id),
    FOREIGN KEY (offer_id) REFERENCES offers(id),
    FOREIGN KEY (payment_id) REFERENCES payments(id)
);