CREATE TABLE IF NOT EXISTS organisations ( 
    id SERIAL PRIMARY KEY, 
    name VARCHAR(255) NOT NULL, 
    description TEXT, 
    address_id INTEGER REFERENCES addresses(id), 
    phone VARCHAR(20), 
    email VARCHAR(255), 
    password_hash VARCHAR(255),
    avatar TEXT,
    website VARCHAR(255), 
    image_id TEXT[],
    is_banned BOOLEAN,
    is_deleted BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP 
);
