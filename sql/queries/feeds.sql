-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFeeds :many
SELECT feeds.name as name, feeds.url as url, users.name as username
FROM feeds
LEFT JOIN users on users.id = feeds.user_id;

-- name: GetFeedByURL :one
SELECT *
FROM feeds
Where url = $1;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = $2, updated_at = $2
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
Select *
FROM feeds
ORDER BY last_fetched_at DESC NULLS FIRST
LIMIT 1;
