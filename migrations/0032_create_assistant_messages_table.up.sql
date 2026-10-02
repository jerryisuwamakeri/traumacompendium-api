CREATE TABLE assistant_messages (
    id              CHAR(36)    NOT NULL PRIMARY KEY,
    conversation_id CHAR(36)    NOT NULL,
    role            VARCHAR(20) NOT NULL,
    content         MEDIUMTEXT  NOT NULL,
    created_at      TIMESTAMP   NOT NULL,

    KEY idx_assistant_messages_conversation_created (conversation_id, created_at),
    CONSTRAINT fk_assistant_messages_conversation FOREIGN KEY (conversation_id) REFERENCES assistant_conversations (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
