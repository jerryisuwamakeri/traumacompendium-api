CREATE TABLE refresh_tokens (
    id           CHAR(36)     NOT NULL PRIMARY KEY, -- matches the issued JWT's jti claim
    user_id      CHAR(36)     NOT NULL,
    token_hash   CHAR(64)     NOT NULL,
    user_agent   VARCHAR(255) NOT NULL DEFAULT '',
    ip_address   VARCHAR(45)  NOT NULL DEFAULT '',
    expires_at   TIMESTAMP    NOT NULL,
    revoked_at   TIMESTAMP    NULL DEFAULT NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_refresh_tokens_user (user_id),
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
