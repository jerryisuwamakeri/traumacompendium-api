CREATE TABLE badges (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    level_id    CHAR(36)     NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    icon_key    VARCHAR(100) NOT NULL,

    UNIQUE KEY uq_badges_level (level_id),
    CONSTRAINT fk_badges_level FOREIGN KEY (level_id) REFERENCES levels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
