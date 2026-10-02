CREATE TABLE trusted_contacts (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    user_id      CHAR(36)     NOT NULL,
    name         VARCHAR(200) NOT NULL,
    relationship VARCHAR(100) NOT NULL,
    phone        VARCHAR(20)  NOT NULL DEFAULT '',
    email        VARCHAR(255) NOT NULL DEFAULT '',
    created_at   TIMESTAMP    NOT NULL,

    KEY idx_trusted_contacts_user (user_id),
    CONSTRAINT fk_trusted_contacts_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
