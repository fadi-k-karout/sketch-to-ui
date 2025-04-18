CREATE TABLE authboss_user_attributes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES authboss_user(id),
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    UNIQUE (user_id, key)
);