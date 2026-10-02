CREATE TABLE subscriptions (
    id                 CHAR(36)     NOT NULL PRIMARY KEY,
    user_id            CHAR(36)     NOT NULL,
    plan               VARCHAR(20)  NOT NULL,
    status             VARCHAR(20)  NOT NULL,
    paystack_reference VARCHAR(100) NOT NULL DEFAULT '',
    current_period_end TIMESTAMP    NOT NULL,
    created_at         TIMESTAMP    NOT NULL,
    updated_at         TIMESTAMP    NOT NULL,

    UNIQUE KEY uq_subscriptions_user (user_id),
    KEY idx_subscriptions_paystack_reference (paystack_reference),
    CONSTRAINT fk_subscriptions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
