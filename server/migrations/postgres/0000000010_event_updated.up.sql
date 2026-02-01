ALTER TABLE calendar_events ADD COLUMN IF NOT EXISTS updated TIMESTAMP DEFAULT NOW();
UPDATE calendar_events SET updated = created WHERE updated IS NULL;
