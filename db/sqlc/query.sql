-- name: CreateThread :one
INSERT INTO threads (id, title, content, category, created_at)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetThread :one
SELECT * FROM threads WHERE id = ?;

-- name: ListThreads :many
SELECT * FROM threads 
WHERE category = sqlc.arg(category)
ORDER BY created_at DESC;

-- name: ListAllThreads :many
SELECT * FROM threads ORDER BY created_at DESC;

-- name: CreateComment :one
INSERT INTO comments (id, thread_id, created_at)
VALUES (?, ?, ?) RETURNING *;

-- name: CreateCommentTranslation :one
INSERT INTO comment_translations (id, comment_id, language, content)
VALUES (?, ?, ?, ?) RETURNING *;

-- name: CreateCommentImage :one
INSERT INTO comment_images (id, comment_id, filename, filepath, created_at)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetThreadComments :many
SELECT 
    c.id,
    c.thread_id,
    c.created_at,
    ct.content,
    ct.language,
    ci.id as image_id,
    ci.filename,
    ci.filepath
FROM comments c
LEFT JOIN comment_translations ct ON c.id = ct.comment_id
LEFT JOIN comment_images ci ON c.id = ci.comment_id
WHERE c.thread_id = ?
ORDER BY c.created_at ASC; 