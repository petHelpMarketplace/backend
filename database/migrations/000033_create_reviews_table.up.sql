CREATE TABLE IF NOT EXISTS reviews (
    id              INTEGER PRIMARY KEY,
    text            TEXT,
    author_id       INTEGER NOT NULL,
    specialist_id   INTEGER,
    organisation_id INTEGER,
    user_id         INTEGER,
    rating          rating_type,
    order_id        INTEGER,
    is_visible      BOOLEAN,
    author_type     author_type,
    recipient_type  recipient_type,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,

    CONSTRAINT fk_reviews_order
        FOREIGN KEY (order_id) REFERENCES orders(id),
        
    CONSTRAINT fk_reviews_specialist
        FOREIGN KEY (specialist_id) REFERENCES specialists(id),
        
    CONSTRAINT fk_reviews_organisation
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
        
    CONSTRAINT fk_reviews_user
        FOREIGN KEY (user_id) REFERENCES users(id),

    CONSTRAINT chk_only_one_recipient
        CHECK (
            (
                specialist_id IS NOT NULL
                AND organisation_id IS NULL
                AND user_id IS NULL
            )
            OR (
                specialist_id IS NULL
                AND organisation_id IS NOT NULL
                AND user_id IS NULL
            )
            OR (
                specialist_id IS NULL
                AND organisation_id IS NULL
                AND user_id IS NOT NULL
            )
        )
); 