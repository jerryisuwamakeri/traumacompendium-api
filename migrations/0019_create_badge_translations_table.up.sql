CREATE TABLE badge_translations (
    badge_id      CHAR(36)     NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    name          VARCHAR(100) NOT NULL,
    description   VARCHAR(500) NOT NULL,
    needs_review  BOOLEAN      NOT NULL DEFAULT TRUE,

    PRIMARY KEY (badge_id, locale),
    CONSTRAINT fk_badge_translations_badge FOREIGN KEY (badge_id) REFERENCES badges (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
