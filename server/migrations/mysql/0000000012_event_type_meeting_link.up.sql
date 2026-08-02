ALTER TABLE calendar_events ADD type VARCHAR(20) DEFAULT 'call' NOT NULL;
ALTER TABLE calendar_events ADD meeting_link VARCHAR(500);
