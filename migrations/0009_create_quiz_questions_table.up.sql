CREATE TABLE quiz_questions (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    quiz_id     CHAR(36)     NOT NULL,
    question    VARCHAR(500) NOT NULL,
    order_index INT          NOT NULL,

    UNIQUE KEY uq_quiz_questions_quiz_order (quiz_id, order_index),
    CONSTRAINT fk_quiz_questions_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
