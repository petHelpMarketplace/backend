CREATE TABLE IF NOT EXISTS orders (
    id              INTEGER PRIMARY KEY,
    order_date      TIMESTAMPTZ,
    user_id         INTEGER NOT NULL,
    description     TEXT,
    organisation_id INTEGER, 
    specialist_id   INTEGER, 
    amount          DOUBLE PRECISION,
    status          order_status DEFAULT 'pending',
    updated_at      TIMESTAMPTZ,

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_orders_organisations
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
    CONSTRAINT fk_orders_specialists
        FOREIGN KEY (specialist_id) REFERENCES specialists(id)
);

CREATE TABLE IF NOT EXISTS order_services (
    order_id   INTEGER,
    service_id INTEGER,
    PRIMARY KEY (order_id, service_id),
    CONSTRAINT fk_order_services_order
        FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_order_services_service
        FOREIGN KEY (service_id) REFERENCES services(id)
);