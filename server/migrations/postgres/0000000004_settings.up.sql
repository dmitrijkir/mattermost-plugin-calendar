-- calendar_members definition
CREATE TABLE IF NOT EXISTS calendar_settings
(
    "user"   varchar NOT null references users (id),
    is_open_calendar_left_bar boolean NOT NULL DEFAULT true,
    first_day_of_week integer default 1 not null,
    hide_non_working_days boolean NOT NULL DEFAULT false
);
