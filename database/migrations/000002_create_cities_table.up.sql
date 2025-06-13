CREATE TABLE IF NOT EXISTS cities (
    id              INTEGER PRIMARY KEY,
    country_id      INTEGER,
    city            VARCHAR(200),

    CONSTRAINT fk_cities_country
        FOREIGN KEY (country_id) REFERENCES countries(id)
);