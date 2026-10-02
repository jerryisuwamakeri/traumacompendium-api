CREATE TABLE webhook_events (
    dedupe_key   CHAR(64)    NOT NULL PRIMARY KEY,
    provider     VARCHAR(50) NOT NULL,
    processed_at TIMESTAMP   NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
