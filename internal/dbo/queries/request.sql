-- name: RequestCreate :batchexec
INSERT INTO request(CREATED_AT, DOMAIN, METHOD, PATH, URL, IP, REQUEST_ID, HEADERS, CONTENT, PARTIAL_CONTENT)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: RequestGet :one
SELECT *
FROM request
WHERE id = $1;