-- +goose Up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID NOT NULL,
    feed_id UUID NOT NULL,
    UNIQUE (user_id, feed_id),
    FOREIGN KEY (feed_id) REFERENCES feeds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;
