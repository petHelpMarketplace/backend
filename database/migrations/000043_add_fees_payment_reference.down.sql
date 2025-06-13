ALTER TABLE fees
    DROP CONSTRAINT IF EXISTS fk_fees_payment,
    DROP COLUMN IF EXISTS payment_id; 