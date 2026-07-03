DROP INDEX IF EXISTS idx_price_aggregates_product_currency;
DROP TABLE IF EXISTS price_aggregates;
DROP INDEX IF EXISTS idx_stores_geo;
ALTER TABLE stores DROP COLUMN IF EXISTS geo;
ALTER TABLE stores DROP COLUMN IF EXISTS long;
ALTER TABLE stores DROP COLUMN IF EXISTS lat;