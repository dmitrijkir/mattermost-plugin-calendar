CREATE TABLE IF NOT EXISTS calendar_ical_tokens (
    token     VARCHAR(64) PRIMARY KEY,
    user_id   VARCHAR(26) NOT NULL UNIQUE,
    created   TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used TIMESTAMP
);
