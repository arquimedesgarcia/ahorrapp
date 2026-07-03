-- 006-loyalty-points: link award rows to receipts, ensure award-once,
-- and optimize the daily-cap count.

ALTER TABLE loyalty_transactions
    ADD COLUMN IF NOT EXISTS receipt_id UUID REFERENCES receipts(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_loyalty_tx_receipt
    ON loyalty_transactions(receipt_id)
    WHERE receipt_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_loyalty_tx_user_day
    ON loyalty_transactions(user_id, created_at);