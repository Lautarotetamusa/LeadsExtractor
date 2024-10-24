/*alter table Communication
rename column source_id to channel_id;

rename table Source to Channel;
*/
alter table Communication
    add column utm_source VARCHAR(256),
    add column utm_medium VARCHAR(256),
    add column utm_campaign VARCHAR(256),
    add column utm_ad VARCHAR(256),
    add column utm_channel VARCHAR(256),
