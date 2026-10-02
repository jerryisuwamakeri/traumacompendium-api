CREATE TABLE assistant_conversations (
    id         CHAR(36)  NOT NULL PRIMARY KEY,
    user_id    CHAR(36)  NOT NULL,
    created_at TIMESTAMP NOT NULL,

    KEY idx_assistant_conversations_user_created (user_id, created_at),
    CONSTRAINT fk_assistant_conversations_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
