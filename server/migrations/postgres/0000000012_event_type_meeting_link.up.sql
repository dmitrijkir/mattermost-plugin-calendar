alter table calendar_events
    add type varchar default 'call' not null;

alter table calendar_events
    add meeting_link varchar;
