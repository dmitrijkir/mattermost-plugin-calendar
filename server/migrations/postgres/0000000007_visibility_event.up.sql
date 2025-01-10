ALTER TABLE calendar_events ADD visibility varchar DEFAULT 'private' NOT NULL;
ALTER TABLE calendar_events ADD team varchar DEFAULT '' NOT NULL;