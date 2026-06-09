ALTER TABLE calendar_events ADD COLUMN updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
UPDATE calendar_events SET updated = created WHERE updated IS NULL;
