CREATE TABLE IF NOT EXISTS appointments (
    id                INTEGER PRIMARY KEY,
    appointment_date  TIMESTAMPTZ,
    user_id          INTEGER NOT NULL,
    location_type    location_type,  -- "at home" or "at specialist"
    description      TEXT,
    organisation_id  INTEGER, 
    specialist_id    INTEGER,              
    amount           DOUBLE PRECISION,
    status           appointment_status DEFAULT 'pending',
    address_id       INTEGER,  -- user, organisation or specialist address
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ,

    CONSTRAINT fk_appointments_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_appointments_address
        FOREIGN KEY (address_id) REFERENCES addresses(id),
    CONSTRAINT fk_appointments_organisations
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
    CONSTRAINT fk_appointments_specialists
        FOREIGN KEY (specialist_id) REFERENCES specialists(id)
);

CREATE TABLE IF NOT EXISTS appointment_services (
    appointment_id INTEGER,
    service_id    INTEGER,
    PRIMARY KEY (appointment_id, service_id),
    CONSTRAINT fk_appointment_services_appointment
        FOREIGN KEY (appointment_id) REFERENCES appointments(id),
    CONSTRAINT fk_appointment_services_service
        FOREIGN KEY (service_id) REFERENCES services(id)
); 