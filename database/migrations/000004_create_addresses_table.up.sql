DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('pending', 'completed', 'cancelled');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'rating_type') THEN
        CREATE TYPE rating_type AS ENUM ('1', '2', '3', '4', '5');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'location_type') THEN
        CREATE TYPE location_type AS ENUM ('вдома', 'у спеціаліста');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'appointment_status') THEN
        CREATE TYPE appointment_status AS ENUM ('pending', 'confirmed', 'completed', 'cancelled');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'animal_size') THEN
        CREATE TYPE animal_size AS ENUM ('Маленький — 0–7 кг', 'Середній — 7–18 кг', 'Великий — 18–45 кг', 'Гігантський — понад 45 кг');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'author_type') THEN
         CREATE TYPE author_type AS ENUM ('user', 'specialist', 'organisation');
    END IF; 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'recipient_type') THEN
        CREATE TYPE recipient_type AS ENUM ('user', 'specialist', 'organisation');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS addresses (
    id          INTEGER PRIMARY KEY,
    country_id  INTEGER,
    city_id     INTEGER,
    area_id     INTEGER,
    postal_code VARCHAR(50),
    street      VARCHAR(200),
    apt         VARCHAR(50),

    CONSTRAINT fk_addresses_city
        FOREIGN KEY (city_id) REFERENCES cities(id),
    CONSTRAINT fk_addresses_area
        FOREIGN KEY (area_id) REFERENCES city_areas(id),
    CONSTRAINT fk_addresses_country
        FOREIGN KEY (country_id) REFERENCES countries(id)
);