CREATE TABLE device_tokens (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    user_id    CHAR(36)     NOT NULL,
    token      VARCHAR(255) NOT NULL,
    platform   VARCHAR(20)  NOT NULL DEFAULT '',
    created_at TIMESTAMP    NOT NULL,

    UNIQUE KEY uq_device_tokens_token (token),
    KEY idx_device_tokens_user (user_id),
    CONSTRAINT fk_device_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
