CREATE TABLE otp_codes (
    id             CHAR(36)     NOT NULL PRIMARY KEY,
    user_id        CHAR(36)     NOT NULL,
    purpose        ENUM('email_verification', 'phone_verification', 'password_reset') NOT NULL,
    code_hash      CHAR(64)     NOT NULL,
    expires_at     TIMESTAMP    NOT NULL,
    consumed_at    TIMESTAMP    NULL DEFAULT NULL,
    attempt_count  INT UNSIGNED NOT NULL DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_otp_codes_user_purpose (user_id, purpose, consumed_at),
    CONSTRAINT fk_otp_codes_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
