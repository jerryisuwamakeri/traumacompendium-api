CREATE TABLE quiz_question_translations (
    question_id   CHAR(36)     NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    text          VARCHAR(500) NOT NULL,
    needs_review  BOOLEAN      NOT NULL DEFAULT TRUE,

    PRIMARY KEY (question_id, locale),
    CONSTRAINT fk_quiz_question_translations_question FOREIGN KEY (question_id) REFERENCES quiz_questions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
