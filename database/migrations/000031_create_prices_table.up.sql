
CREATE TABLE IF NOT EXISTS prices (
    id              INTEGER PRIMARY KEY,
    service_id      INTEGER NOT NULL,
    amount          DOUBLE PRECISION,
    animal_id       INTEGER,     -- if referencing a single Animal
    organisation_id INTEGER, 
    specialist_id   INTEGER,   

    CONSTRAINT fk_prices_animals
        FOREIGN KEY (animal_id) REFERENCES animals(id),
    CONSTRAINT fk_prices_service
        FOREIGN KEY (service_id) REFERENCES services(id),
    CONSTRAINT fk_prices_organisations
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
    CONSTRAINT fk_prices_specialists
        FOREIGN KEY (specialist_id) REFERENCES specialists(id)
); 