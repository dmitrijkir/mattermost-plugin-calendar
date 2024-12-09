alter table calendar_events
    rename column "start" to dt_start;

alter table calendar_events
    rename column "end" to dt_end;

alter table calendar_members
    rename column "user" to member;

alter table calendar_settings
    rename column "user" to owner;