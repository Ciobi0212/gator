-- +goose Up
CREATE TABLE feed (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    url VARCHAR NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    FOREIGN KEY (user_id)
    REFERENCES users(id) 
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed;