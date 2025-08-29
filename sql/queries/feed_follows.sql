-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT inserted_feed_follows.*, users.name AS username, feeds.name AS feedname
FROM inserted_feed_follows
INNER JOIN users ON inserted_feed_follows.user_id = users.id
INNER JOIN feeds ON inserted_feed_follows.feed_id = feeds.id;
