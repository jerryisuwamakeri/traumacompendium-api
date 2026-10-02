CREATE TABLE community_comments (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    post_id    CHAR(36)     NOT NULL,
    user_id    CHAR(36)     NOT NULL,
    body       TEXT         NOT NULL,
    status     VARCHAR(20)  NOT NULL,
    created_at TIMESTAMP    NOT NULL,
    deleted_at TIMESTAMP    NULL DEFAULT NULL,

    KEY idx_community_comments_post_created (post_id, created_at),
    CONSTRAINT fk_community_comments_post FOREIGN KEY (post_id) REFERENCES community_posts (id) ON DELETE CASCADE,
    CONSTRAINT fk_community_comments_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
