CREATE TABLE IF NOT EXISTS consumer_notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    offer_negotiation_id UUID,
    payment_id UUID,
    order_id UUID,
    type INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS service_provider_notifications (
    id UUID PRIMARY KEY,
    service_provider_id UUID NOT NULL,
    offer_id UUID,
    offer_negotiation_id UUID,
    order_id UUID,
    type INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES service_providers(id)
);

