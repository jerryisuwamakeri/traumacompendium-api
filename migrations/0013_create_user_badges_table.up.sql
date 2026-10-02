CREATE TABLE user_badges (
    user_id    CHAR(36)  NOT NULL,
    badge_id   CHAR(36)  NOT NULL,
    awarded_at TIMESTAMP NOT NULL,

    PRIMARY KEY (user_id, badge_id),
    CONSTRAINT fk_user_badges_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_badges_badge FOREIGN KEY (badge_id) REFERENCES badges (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
