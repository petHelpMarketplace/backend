CREATE TABLE IF NOT EXISTS unauthorized_users_emails (
    id              INTEGER PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    email           VARCHAR(200),
);

