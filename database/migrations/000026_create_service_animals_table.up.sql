CREATE TABLE IF NOT EXISTS service_animals (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id),
    animal_category_id INTEGER REFERENCES animal_categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, animal_category_id)
); 