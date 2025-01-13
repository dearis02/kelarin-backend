CREATE TABLE IF NOT EXISTS service_provider_areas(
    id BIGSERIAL PRIMARY KEY,
    service_provider_id UUID NOT NULL,
    province_id BIGINT NOT NULL,
    city_id BIGINT NOT NULL,
    district_id BIGINT
);