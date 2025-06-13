ALTER TABLE orders
    ADD COLUMN payment_id INTEGER,
    ADD CONSTRAINT fk_orders_payment
    FOREIGN KEY (payment_id) REFERENCES payments(id); 