CREATE TABLE quizzes (
    id         CHAR(36)  NOT NULL PRIMARY KEY,
    lesson_id  CHAR(36)  NOT NULL,
    created_at TIMESTAMP NOT NULL,

    UNIQUE KEY uq_quizzes_lesson (lesson_id),
    CONSTRAINT fk_quizzes_lesson FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
