-- The text-to-speech cache: populated lazily on first request per
-- (lesson_id, locale), not seeded. See internal/notification/tts and
-- learningservice.Service.GetLessonAudio.
CREATE TABLE lesson_audio (
    lesson_id     CHAR(36)    NOT NULL,
    locale        VARCHAR(10) NOT NULL,
    audio_data    MEDIUMBLOB  NOT NULL,
    content_type  VARCHAR(100) NOT NULL,
    generated_at  TIMESTAMP   NOT NULL,

    PRIMARY KEY (lesson_id, locale),
    CONSTRAINT fk_lesson_audio_lesson FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
