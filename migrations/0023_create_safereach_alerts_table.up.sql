CREATE TABLE safereach_alerts (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    user_id      CHAR(36)     NOT NULL,
    level        VARCHAR(20)  NOT NULL,
    status       VARCHAR(20)  NOT NULL,
    message      VARCHAR(500) NOT NULL DEFAULT '',
    share_token  CHAR(43)     NOT NULL,
    triggered_at TIMESTAMP    NOT NULL,
    resolved_at  TIMESTAMP    NULL DEFAULT NULL,

    UNIQUE KEY uq_safereach_alerts_share_token (share_token),
    KEY idx_safereach_alerts_user_triggered (user_id, triggered_at),
    CONSTRAINT fk_safereach_alerts_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
