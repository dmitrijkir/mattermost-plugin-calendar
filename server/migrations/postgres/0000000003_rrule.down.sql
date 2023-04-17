alter table calendar_events
    alter column recurrence type jsonb using recurrence::jsonb;

