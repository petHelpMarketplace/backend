ALTER TABLE fees
    ADD COLUMN payment_id INTEGER,
    ADD CONSTRAINT fk_fees_payment
    FOREIGN KEY (payment_id) REFERENCES payments(id); 