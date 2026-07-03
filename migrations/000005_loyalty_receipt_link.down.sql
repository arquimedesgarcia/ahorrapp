-- 006-loyalty-points: reverse the receipt_id link migration.

DROP INDEX IF EXISTS idx_loyalty_tx_user_day;
DROP INDEX IF EXISTS uniq_loyalty_tx_receipt;

ALTER TABLE loyalty_transactions
    DROP COLUMN IF EXISTS receipt_id;