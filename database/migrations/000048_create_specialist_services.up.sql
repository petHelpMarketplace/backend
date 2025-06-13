CREATE TABLE IF NOT EXISTS specialist_services (
    specialist_id INTEGER,
    service_id      INTEGER,

    PRIMARY KEY (specialist_id, service_id),
    CONSTRAINT fk_services_spec
        FOREIGN KEY (specialist_id) REFERENCES specialists(id),
    CONSTRAINT fk_services_service
        FOREIGN KEY (service_id) REFERENCES services(id)
); 