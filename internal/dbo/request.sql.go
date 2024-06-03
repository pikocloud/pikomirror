// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: request.sql

package dbo

import (
	"context"
)

const requestGet = `-- name: RequestGet :one
SELECT id, created_at, request_id, domain, method, path, url, ip, headers, content, partial_content
FROM request
WHERE id = $1
`

func (q *Queries) RequestGet(ctx context.Context, id int64) (Request, error) {
	row := q.db.QueryRow(ctx, requestGet, id)
	var i Request
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.RequestID,
		&i.Domain,
		&i.Method,
		&i.Path,
		&i.URL,
		&i.IP,
		&i.Headers,
		&i.Content,
		&i.PartialContent,
	)
	return i, err
}