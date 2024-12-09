CREATE TABLE IF NOT EXISTS calendar_members
(
    event    VARCHAR(50)           NOT NULL,
    member     VARCHAR(50)           NOT NULL,
    accepted BOOLEAN DEFAULT FALSE NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
