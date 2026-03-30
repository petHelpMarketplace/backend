CREATE TABLE IF NOT EXISTS city_areas (
    id              INTEGER PRIMARY KEY,
    city_id         INTEGER,
    area_name       VARCHAR(200),

    CONSTRAINT fk_city_areas_city
        FOREIGN KEY (city_id) REFERENCES cities(id)
);