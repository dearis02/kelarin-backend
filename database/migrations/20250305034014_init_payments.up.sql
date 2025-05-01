DO $$
BEGIN
    CREATE TYPE payment_method_type AS ENUM (
        'va',
        'qr'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'payment_method_type type already exists';
END $$;

DO $$
BEGIN
    CREATE TYPE payment_method_admin_fee_unit AS ENUM (
        'fixed',
        'percent'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'payment_method_admin_fee_unit type already exists';
END $$;

DO $$
BEGIN
    CREATE TYPE payment_status AS ENUM (
        'pending',
        'paid',
        'canceled',
        'expired',
        'failed'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'payment_status type already exists';
END $$;

CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type payment_method_type NOT NULL,
    code VARCHAR(20) NOT NULL,
    admin_fee INT NOT NULL,
    admin_fee_unit payment_method_admin_fee_unit NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    logo TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    reference VARCHAR(50) NOT NULL UNIQUE,
    payment_method_id UUID NOT NULL,
    user_id UUID NOT NULL,
    amount DECIMAL (15,2) NOT NULL,
    admin_fee INT NOT NULL,
    platform_fee INT NOT NULL,
    payment_link TEXT NOT NULL,
    status payment_status NOT NULL DEFAULT 'pending',
    expired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);