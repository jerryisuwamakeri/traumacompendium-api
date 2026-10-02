CREATE TABLE levels (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    number      TINYINT      NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    created_at  TIMESTAMP    NOT NULL,

    UNIQUE KEY uq_levels_number (number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
