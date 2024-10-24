CREATE TABLE Utm (
    id int not null auto_increment,
    code varchar(64) not null,
    utm_source varchar(256) default null,
    utm_medium varchar(256) default null,
    utm_campaign varchar(256) default null,
    utm_ad      varchar(256) default null,
    utm_channel enum('ivr', 'inbox', 'whatsapp', 'email', 'flyer'),

    primary key (id),
    unique (code)
);
