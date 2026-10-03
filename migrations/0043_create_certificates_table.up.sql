CREATE TABLE certificates (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    user_id    CHAR(36)     NOT NULL,
    level_id   CHAR(36)     NOT NULL,
    serial     VARCHAR(40)  NOT NULL,
    issued_at  TIMESTAMP    NOT NULL,

    UNIQUE KEY uq_certificates_user_level (user_id, level_id),
    UNIQUE KEY uq_certificates_serial (serial),
    CONSTRAINT fk_certificates_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_certificates_level FOREIGN KEY (level_id) REFERENCES levels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
