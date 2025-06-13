CREATE TABLE IF NOT EXISTS animal_sizes (
    id SERIAL PRIMARY KEY,
    animal_category_id INTEGER REFERENCES animal_categories(id),
    name VARCHAR(100) NOT NULL,
    min_weight DECIMAL(5,2),
    max_weight DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
); 