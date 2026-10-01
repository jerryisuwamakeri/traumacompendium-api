CREATE TABLE users (
    id                 CHAR(36)     NOT NULL PRIMARY KEY,
    full_name          VARCHAR(150) NOT NULL,
    email              VARCHAR(255) NOT NULL,
    email_verified_at  TIMESTAMP    NULL DEFAULT NULL,
    phone              VARCHAR(20)  NOT NULL,
    phone_verified_at  TIMESTAMP    NULL DEFAULT NULL,
    password_hash      VARCHAR(255) NOT NULL,
    role               ENUM('user', 'admin') NOT NULL DEFAULT 'user',
    status             ENUM('active', 'suspended') NOT NULL DEFAULT 'active',
    created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at         TIMESTAMP    NULL DEFAULT NULL,

    UNIQUE KEY uq_users_email (email),
    UNIQUE KEY uq_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
