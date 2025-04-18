-- +goose Up
CREATE TABLE feed_follows (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    feed_id INTEGER NOT NULL,
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,
    FOREIGN KEY (feed_id)
    REFERENCES feed(id)
    ON DELETE CASCADE,
    UNIQUE(user_id, feed_id)
);

ALTER TABLE feed
DROP COLUMN user_id;

-- +goose Down
DROP TABLE feed_follows;

ALTER TABLE feed 
ADD COLUMN user_id uuid;

ALTER table feed 
ADD CONSTRAINT feed_user_id_fkey FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;