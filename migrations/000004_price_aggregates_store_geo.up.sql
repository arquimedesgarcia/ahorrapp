-- 005-price-ranking-engine: price_aggregates table, store geolocation, unaccent

-- 1. Add geolocation to stores
ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS long DOUBLE PRECISION;

-- 2. PostGIS geography column + GiST index for proximity queries
ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS geo geography(Point, 4326)
    GENERATED ALWAYS AS (
        CASE WHEN lat IS NOT NULL AND long IS NOT NULL
             THEN ST_MakePoint(long, lat)::geography
        END
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_stores_geo ON stores USING GIST (geo);

-- 3. Price aggregates cache table
CREATE TABLE IF NOT EXISTS price_aggregates (
    product_id       UUID        NOT NULL REFERENCES products(id),
    store_id         UUID        NOT NULL REFERENCES stores(id),
    currency         TEXT        NOT NULL CHECK (currency IN ('USD', 'Bs.')),
    average_price    NUMERIC(12,2) NOT NULL,
    min_price        NUMERIC(12,2) NOT NULL,
    sample_count     INT         NOT NULL CHECK (sample_count >= 0),
    last_observed_at TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, store_id, currency)
);

CREATE INDEX IF NOT EXISTS idx_price_aggregates_product_currency
    ON price_aggregates(product_id, currency);

-- 4. Accent-insensitive search support
CREATE EXTENSION IF NOT EXISTS unaccent;