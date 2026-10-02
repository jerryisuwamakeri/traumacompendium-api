CREATE TABLE quiz_options (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    question_id CHAR(36)     NOT NULL,
    text        VARCHAR(500) NOT NULL,
    is_correct  BOOLEAN      NOT NULL DEFAULT FALSE,
    order_index INT          NOT NULL,

    UNIQUE KEY uq_quiz_options_question_order (question_id, order_index),
    CONSTRAINT fk_quiz_options_question FOREIGN KEY (question_id) REFERENCES quiz_questions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
