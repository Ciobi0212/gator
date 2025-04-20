-- name: CreateFeed :one
INSERT INTO feed(name, url,created_at,updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT * FROM feed;

-- name: FindFeedByURL :one
select * from feed 
WHERE url = $1;

-- name: DeleteAllFeeds :exec
DELETE from feed;

-- name: MarkFeedFetched :exec
UPDATE feed 
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;


-- name: GetNextFeedsToFetch :many
SELECT * from feed 
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;