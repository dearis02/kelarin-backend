CREATE TABLE IF NOT EXISTS service_feedbacks(
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    service_id UUID NOT NULL,
    rating SMALLINT NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (service_id) REFERENCES services(id)
);