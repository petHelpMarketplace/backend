DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role') THEN
        CREATE TYPE role AS ENUM ('user', 'moderator', 'admin');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id               INTEGER PRIMARY KEY,
    name             VARCHAR(100),
    family_name      VARCHAR(100),
    email            VARCHAR(200) UNIQUE,
    password_hash    VARCHAR(200),
    role             role NOT NULL DEFAULT 'user',
    is_banned        BOOLEAN,
    is_deleted       BOOLEAN,
    image_id         TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ,
    shipping_address_id INTEGER,
    billing_address_id  INTEGER,
    phone            VARCHAR(50),

    CONSTRAINT fk_users_shipping_address
        FOREIGN KEY (shipping_address_id) REFERENCES addresses(id),
    CONSTRAINT fk_users_billing_address
        FOREIGN KEY (billing_address_id) REFERENCES addresses(id)
);