CREATE TABLE IF NOT EXISTS calendar_events (
id varchar NOT NULL PRIMARY KEY,
title varchar NOT NULL,
"start" timestamp NOT NULL,
"end" timestamp NOT NULL,
created timestamp NOT NULL,
owner varchar NOT NULL references users(id),
"channel" varchar references channels(id),
processed timestamp,
recurrence jsonb NULL,
recurrent boolean NOT NULL DEFAULT false
);
-- calendar_members definition
CREATE TABLE IF NOT EXISTS calendar_members (
"event" varchar NOT null references calendar_events(id) ON DELETE CASCADE,
"user" varchar NOT null references users(id),
accepted boolean NOT NULL DEFAULT false
);
