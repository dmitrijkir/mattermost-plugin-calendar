CREATE TABLE IF NOT EXISTS calendar_ical_tokens (
    token     VARCHAR(64) NOT NULL PRIMARY KEY,
    user_id   VARCHAR(26) NOT NULL,
    created   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP NULL,
    UNIQUE KEY unique_user (user_id)
);
