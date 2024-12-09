alter table calendar_events
    rename column dt_start to "start";

alter table calendar_events
    rename column dt_end to "end";

alter table calendar_members
    rename column member to "user";

alter table calendar_settings
    rename column owner to "user";