alter table calendar_events
    alter column recurrence type varchar using recurrence::varchar;
