alter table calendar_events
    rename column dt_start to 'start';

alter table calendar_events
    rename column 'dt_end' to 'end';