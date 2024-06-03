-- name: EndpointList :many
SELECT *
FROM endpoint
ORDER BY domain, path, method
OFFSET $1 LIMIT $2;


-- name: EndpointListTop :many
SELECT *
FROM endpoint
ORDER BY hits DESC, domain, path, method
LIMIT $1;