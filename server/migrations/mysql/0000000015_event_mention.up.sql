-- Existing events default to no mention on purpose: turning one on for rows
-- that predate the setting would start pinging channels that never asked.
ALTER TABLE calendar_events ADD mention VARCHAR(20) DEFAULT '' NOT NULL;
