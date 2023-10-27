ALTER TABLE calendar_events ADD description varchar DEFAULT '';
alter table calendar_members
    drop constraint if exists calendar_members_user_fkey;