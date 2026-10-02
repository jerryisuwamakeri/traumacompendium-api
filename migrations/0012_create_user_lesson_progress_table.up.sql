CREATE TABLE user_lesson_progress (
    user_id      CHAR(36)  NOT NULL,
    lesson_id    CHAR(36)  NOT NULL,
    completed_at TIMESTAMP NOT NULL,

    PRIMARY KEY (user_id, lesson_id),
    CONSTRAINT fk_user_lesson_progress_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_lesson_progress_lesson FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
