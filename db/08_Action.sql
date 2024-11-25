CREATE TABLE Action (
    name varchar(256) not null,
    nro int not null,
    flow_uuid char(37) not null,
    wamid   char(63) not null,
    sended_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    lead_phone CHAR(16) NOT NULL,
    on_response CHAR(37) DEFAULT NULL,

    foreign key (lead_phone) references Leads(phone)
);
