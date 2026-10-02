CREATE TABLE safereach_location_updates (
    id           CHAR(36)  NOT NULL PRIMARY KEY,
    alert_id     CHAR(36)  NOT NULL,
    lat          DOUBLE    NOT NULL,
    lng          DOUBLE    NOT NULL,
    recorded_at  TIMESTAMP NOT NULL,

    KEY idx_safereach_location_updates_alert_recorded (alert_id, recorded_at),
    CONSTRAINT fk_safereach_location_updates_alert FOREIGN KEY (alert_id) REFERENCES safereach_alerts (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
