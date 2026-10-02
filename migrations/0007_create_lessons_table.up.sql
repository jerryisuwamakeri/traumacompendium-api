CREATE TABLE lessons (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    level_id    CHAR(36)     NOT NULL,
    title       VARCHAR(200) NOT NULL,
    content     MEDIUMTEXT   NOT NULL,
    order_index INT          NOT NULL,
    created_at  TIMESTAMP    NOT NULL,

    UNIQUE KEY uq_lessons_level_order (level_id, order_index),
    CONSTRAINT fk_lessons_level FOREIGN KEY (level_id) REFERENCES levels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
