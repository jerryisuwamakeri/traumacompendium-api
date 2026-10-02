CREATE TABLE lesson_translations (
    lesson_id     CHAR(36)     NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    title         VARCHAR(200) NOT NULL,
    content       MEDIUMTEXT   NOT NULL,
    needs_review  BOOLEAN      NOT NULL DEFAULT TRUE,

    PRIMARY KEY (lesson_id, locale),
    CONSTRAINT fk_lesson_translations_lesson FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
