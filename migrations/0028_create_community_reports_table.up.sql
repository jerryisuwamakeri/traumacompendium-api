CREATE TABLE community_reports (
    id                  CHAR(36)     NOT NULL PRIMARY KEY,
    target_type         VARCHAR(20)  NOT NULL,
    target_id           CHAR(36)     NOT NULL,
    reporter_user_id    CHAR(36)     NULL,
    reason              VARCHAR(500) NOT NULL DEFAULT '',
    status              VARCHAR(20)  NOT NULL,
    resolved_by_user_id CHAR(36)     NULL,
    resolved_at         TIMESTAMP    NULL DEFAULT NULL,
    created_at          TIMESTAMP    NOT NULL,

    KEY idx_community_reports_status_created (status, created_at),
    KEY idx_community_reports_target (target_type, target_id),
    CONSTRAINT fk_community_reports_reporter FOREIGN KEY (reporter_user_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_community_reports_resolver FOREIGN KEY (resolved_by_user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
