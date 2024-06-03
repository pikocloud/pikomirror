-- +migrate Up

CREATE TABLE request
(
    ID              BIGSERIAL   NOT NULL PRIMARY KEY,
    CREATED_AT      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    REQUEST_ID      TEXT        NOT NULL,
    DOMAIN          TEXT        NOT NULL,
    METHOD          TEXT        NOT NULL,
    PATH            TEXT        NOT NULL,
    URL             TEXT        NOT NULL,
    IP              INET        NOT NULL,
    HEADERS         JSONB       NOT NULL,
    CONTENT         BYTEA       NOT NULL,
    PARTIAL_CONTENT BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX request_domain_method_path ON request (DOMAIN, PATH, METHOD);
CREATE INDEX request_create_at ON request (CREATED_AT);
CREATE INDEX request_request_id ON request (REQUEST_ID); -- non unique


CREATE VIEW endpoint AS
SELECT "domain", method, path, count(1) as hits, max(CREATED_AT) ::timestamptz as last_update
FROM request
GROUP BY "domain", path, method;

CREATE VIEW headline AS
SELECT ID, request.CREATED_AT, request.REQUEST_ID, request.DOMAIN, request.METHOD, request.PATH, IP
FROM request
ORDER BY ID;