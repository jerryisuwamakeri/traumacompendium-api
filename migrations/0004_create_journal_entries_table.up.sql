CREATE TABLE journal_entries (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    user_id     CHAR(36)     NOT NULL,
    title       VARCHAR(200) NOT NULL,
    content     TEXT         NOT NULL,
    created_at  TIMESTAMP    NOT NULL,
    updated_at  TIMESTAMP    NOT NULL,
    deleted_at  TIMESTAMP    NULL DEFAULT NULL,

    KEY idx_journal_entries_user_created (user_id, created_at),
    CONSTRAINT fk_journal_entries_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
