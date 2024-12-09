CREATE TABLE IF NOT EXISTS calendar_events
(
    id          VARCHAR(50)                NOT NULL PRIMARY KEY,
    title       VARCHAR(255)               NOT NULL,
    dt_start       TIMESTAMP                  NOT NULL,
    dt_end         TIMESTAMP                  NOT NULL,
    created     TIMESTAMP                  NOT NULL,
    owner       VARCHAR(50)                NOT NULL,
    channel     VARCHAR(50),
    processed   TIMESTAMP,
    recurrence  VARCHAR(50),
    recurrent   BOOLEAN      DEFAULT FALSE NOT NULL,
    color       VARCHAR(7),
    description VARCHAR(255) DEFAULT ''    NOT NULL
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
