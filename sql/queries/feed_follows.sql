-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows(created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4)
    RETURNING *
)

SELECT inserted_feed_follows.*, users.name, feed.name 
FROM inserted_feed_follows 
JOIN users on users.id = inserted_feed_follows.user_id
JOIN feed on feed.id = inserted_feed_follows.feed_id;

-- name: GetFeedFollowsForUser :many

WITH feed_follows_entries AS (
    SELECT * from feed_follows
    WHERE user_id = $1
)

SELECT feed.name FROM feed_follows_entries
JOIN feed ON feed.id = feed_follows_entries.feed_id;

-- name: DeleteFeedFollowsEntry :exec
DELETE FROM feed_follows
WHERE user_id = $1 and feed_id = $2;

-- name: DeleteAllFeedFollows :exec
DELETE FROM feed_follows;