DO $$
BEGIN
    CREATE TYPE offer_status AS ENUM (
        'pending', 
        'accepted',
        'rejected',
        'canceled'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'offer_status type already exists';
END $$;

DO $$
BEGIN
    CREATE TYPE offer_negotiation_status AS ENUM (
        'pending', 
        'accepted',
        'rejected',
        'canceled'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'offer_negotiation_status type already exists';
END $$;

CREATE TABLE IF NOT EXISTS offers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_address_id UUID NOT NULL,
    service_id UUID NOT NULL,
    detail TEXT NOT NULL,
    service_cost DECIMAL(15, 2) NOT NULL,
    service_start_date DATE NOT NULL,
    service_end_date DATE NOT NULL,
    service_start_time TIMETZ NOT NULL,
    service_end_time TIMETZ NOT NULL,
    status offer_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (user_address_id) REFERENCES user_addresses(id),
    FOREIGN KEY (service_id) REFERENCES services(id)
);

CREATE TABLE offer_negotiations (
    id UUID PRIMARY KEY,
    offer_id UUID NOT NULL,
    message TEXT NOT NULL,
    requested_service_cost DECIMAL(15,2),
    status offer_negotiation_status NOT NULL,
    created_at timestamp NOT NULL,
    FOREIGN KEY (offer_id) REFERENCES offers(id)
);