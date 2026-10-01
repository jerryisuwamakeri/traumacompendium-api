CREATE TABLE mood_entries (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    user_id     CHAR(36)     NOT NULL,
    score       TINYINT      NOT NULL,
    tags        JSON         NOT NULL,
    note        VARCHAR(2000) NOT NULL DEFAULT '',
    created_at  TIMESTAMP    NOT NULL,
    deleted_at  TIMESTAMP    NULL DEFAULT NULL,

    KEY idx_mood_entries_user_created (user_id, created_at),
    CONSTRAINT fk_mood_entries_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_mood_entries_score CHECK (score BETWEEN 1 AND 5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
