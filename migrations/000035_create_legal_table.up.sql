CREATE TABLE legal (
    id    INTEGER PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 
    document_type  VARCHAR(100),
    document_url TEXT,
    title VARCHAR(200)
);