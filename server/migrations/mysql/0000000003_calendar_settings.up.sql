CREATE TABLE IF NOT EXISTS calendar_settings
(
    owner                      VARCHAR(50)           NOT NULL,
    is_open_calendar_left_bar BOOLEAN DEFAULT TRUE  NOT NULL,
    first_day_of_week         INTEGER DEFAULT 1     NOT NULL,
    hide_non_working_days     BOOLEAN DEFAULT FALSE NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
