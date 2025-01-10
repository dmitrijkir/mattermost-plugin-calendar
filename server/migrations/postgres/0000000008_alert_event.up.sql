ALTER TABLE calendar_events ADD alert varchar DEFAULT '' NOT NULL;
ALTER TABLE calendar_events ADD alert_time timestamp DEFAULT NULL;
