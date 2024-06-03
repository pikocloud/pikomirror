-- name: HeadlineList :many
SELECT *
FROM headline
ORDER BY ID desc
OFFSET $1 LIMIT $2;

-- name: HeadlineListByEndpoint :many
SELECT *
FROM headline
WHERE "domain" = $1
  AND method = $2
  AND path = $3
ORDER BY ID DESC
OFFSET $4 LIMIT $5;