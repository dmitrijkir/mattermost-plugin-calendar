-- Existing events default to no mention on purpose: turning one on for rows
-- that predate the setting would start pinging channels that never asked.
alter table calendar_events
    add mention varchar default '' not null;
