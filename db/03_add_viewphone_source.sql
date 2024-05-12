alter table Source
    modify column type enum('whatsapp','ivr','property', 'viewphone') NOT NULL;
insert into Source (type) values ('viewphone');

