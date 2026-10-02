CREATE TABLE quiz_option_translations (
    option_id     CHAR(36)     NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    text          VARCHAR(500) NOT NULL,
    needs_review  BOOLEAN      NOT NULL DEFAULT TRUE,

    PRIMARY KEY (option_id, locale),
    CONSTRAINT fk_quiz_option_translations_option FOREIGN KEY (option_id) REFERENCES quiz_options (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
