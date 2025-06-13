ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS fk_orders_payment,
    DROP COLUMN IF EXISTS payment_id; 