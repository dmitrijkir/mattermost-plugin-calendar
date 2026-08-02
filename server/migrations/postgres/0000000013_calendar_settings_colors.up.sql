alter table calendar_settings
    add call_color varchar(7) default '#B3E1F7' not null;

alter table calendar_settings
    add event_color varchar(7) default '#B6D9C7' not null;
