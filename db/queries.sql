/* Lista de LEADS tal cual estaba en el google sheets*/
SELECT 
    C.created_at as "Fecha extraccion", C.lead_date as "Fecha lead", A.name as "Asesor asignado", S.type, P.portal, L.name, C.url, L.phone, L.email,
    P.*,
    C.zones, C.mt2_terrain, C.mt2_builded, C.baths, C.rooms
FROM Communication C
INNER JOIN Leads L 
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id
ORDER BY C.id DESC;
/* 6728 */

/* Lista de comunicaciones sin duplicar (en teoria) (saco la fecha porque ese campo es unico) */
SELECT phone FROM ( 
    SELECT distinct 
        C.lead_date as "Fecha lead", A.name as "Asesor asignado", S.type, L.name, C.url as "link", L.phone, L.email,
        P.*,
        C.zones, C.mt2_terrain, C.mt2_builded, C.baths, C.rooms
    FROM Communication C
    INNER JOIN Leads L 
        ON C.lead_phone = L.phone
    INNER JOIN Source S
        ON C.source_id = S.id
    INNER JOIN Asesor A
        ON L.asesor = A.phone
    LEFT JOIN Property P
        ON S.property_id = P.id
    ORDER BY C.id DESC
) as A;
/* 3577 */

select distinct lead_phone, lead_date, source_id, url from Communication; 
/* 2433 */

/* Lista de telefono unicos (leads distintos) */
select distinct lead_phone from Communication; 
/* 17 jun 2024 -> 500 */
/* 26 jun 2024 -> 650 | 150 mas*/
/* dsp del inmuebles24, lamudi, casasyterrenos first_run -> 2040 | 1390 mas */

select count(*) from Leads; 
/*17 jun 2024 -> 637 */
/*26 jun 2024 -> 786 | 131 mas*/ 
/* dsp del inmuebles24 y lamudi, casasyterrenos first_run -> 2119 | 1333 mas */

select count(*) from Communication;
/*26 jun 2024 -> 4439*/ 
/* dsp del inmuebles24 y lamudi first_run -> 6726  | 2287 mas */
/* 2287 - 1139 = 1148 */


/* Lista negra, no está ninguno xd*/
select * from Communication where lead_phone IN
(
"5213332072880",
"5213333610908",
"5213316035907",
"5213336765454",
"5213310088151",
"5213111227357",
"5213319167138",
"5213318943811",
"5213310258112"
);

/* Communicaciones del ultimo dia */
SELECT 
    C.created_at as "Fecha extraccion", C.lead_date as "Fecha lead", A.name as "Asesor asignado", S.type, P.portal, L.name, C.url, L.phone, L.email,
    P.*,
    C.zones, C.mt2_terrain, C.mt2_builded, C.baths, C.rooms
FROM Communication C
INNER JOIN Leads L 
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id
WHERE C.created_at >= date_add(curdate(), interval -1 day)
ORDER BY C.id DESC;


/* Communicaciones del ultimo dia por tipo */
SELECT S.type, P.portal, COUNT(*)
FROM Communication C
INNER JOIN Leads L 
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id
WHERE C.created_at >= date_add(curdate(), interval -1 day)
AND C.new_lead = true
GROUP BY S.type, P.portal;


/* Into outfile */
        SELECT 
            C.created_at, 
            C.lead_date,
            C.new_lead,
            L.cotizacion,
            A.name as "asesor.name", A.phone as "asesor.phone", A.email as "asesor.email",
            IF(S.type = "property", P.portal, S.type) as "fuente",
            L.name, C.url, L.phone, L.email,
                IFNULL(P.portal_id, "") as "propiedad.portal_id", 
                IFNULL(P.title, "") as "propiedad.title", 
                IFNULL(P.price, "") as "propiedad.price", 
                IFNULL(P.ubication, "") as "propiedad.ubication", 
                IFNULL(P.url, "") as "propiedad.url", 
                IFNULL(P.tipo, "") as "propiedad.tipo",
            C.zones as "busquedas.zones", C.mt2_terrain as "busquedas.mt2_terrain", C.mt2_builded as "busquedas.mt2_builded", C.baths as "busquedas.baths", C.rooms as "busquedas.rooms"
        FROM Communication C
        INNER JOIN Leads L
            ON C.lead_phone = L.phone
        INNER JOIN Source S
            ON C.source_id = S.id
        INNER JOIN Asesor A
            ON L.asesor = A.phone
        LEFT JOIN Property P
            ON S.property_id = P.id
        ORDER BY C.id DESC
        INTO OUTFILE '/home/leads.csv';

/*
Necesito los siguientes datos de cada lead: fecha en la que llegó, portal por el que llegó (o medio como WhatsApp), precio de la propiedad por la que pregunto y zona de la propiedad (si está disponible), la cantidad de mensajes duplicados que tenemos de ese lead (interacciones que ha tenido con nosotros), número de WhatsApp o teléfono, asesor asignado y número del asesor asignado.
*/

select lead_phone, count(*) from Communication group by lead_phone;

SELECT "Telefono", "Mensajes", "Asesor telefono", "Asesor nombre", "Cotizacion", "Fecha", "Fecha lead", "Fuente", "Portal_id", "Title", "Price", "Ubicacion", "Link", "Tipo"
UNION
SELECT
    L.phone as "Telefono",
    count(L.phone) as "Mensajes",
    A.phone as "Asesor telefono",
    A.name as "Asesor nombre",
    L.cotizacion as "Cotizacion",
    C.created_at as "Fecha",
    C.lead_date as "Fecha lead",
    IF(S.type = "property", P.portal, S.type) as "Fuente",
        IFNULL(P.portal_id, "") as "Portal_id", IFNULL(P.title, "") as "Title", 
        IFNULL(P.price, "") as "Price", 
        IFNULL(P.ubication, "") as "Ubicacion", 
        IFNULL(P.url, "") as "Link", 
        IFNULL(P.tipo, "") as "Tipo"
    FROM Communication C
    INNER JOIN Leads L
        ON C.lead_phone = L.phone
    INNER JOIN Source S
        ON C.source_id = S.id
    INNER JOIN Asesor A
        ON L.asesor = A.phone
    LEFT JOIN Property P
        ON S.property_id = P.id
    GROUP BY L.phone
    INTO OUTFILE '/tmp/leads.csv';


/* Cambiar la zona horaria a mexico
vamos a restar -6hs a todos los leads */
select * from Communication
SET created_at = DATE_SUB(created_at, INTERVAL 6 HOUR);

update Communication
set created_at="2024-07-18T00:00:01"
where id=38;
select * from Communication where id=38;

/*maggie*/
select * from Leads
where asesor="5213314299454";

/*Lucy*/
select * from Leads
where asesor="5213317186543";

/* Agregar asesor */
insert into Asesor
(email, phone, name, active)
values
("mmichel.marti@gmail.com", "+523321608647", "Marti Michel", true);

select * from Leads
where asesor="5213313291761";
select * from Leads
where asesor="5213317186543"; /*87 leads, 270 comms */

delete from Asesor
where phone="5213313291761";

update Asesor
set name="Eduardo Jordan",
email="eduardo.jordan@rebora.com.mx "
where phone="5213317186543";

update Leads 
set asesor="5213317186543"
where asesor="5213313291761";

update Asesor
set phone="5213317186543"
where phone="5213313291761";

update Asesor
set active=false
where phone="5213322563353";


/* update phones to E.164 */
insert into Leads (name, phone, asesor)
values ("wppTest", "5493415854220", "+5493415854220");

select * from Leads
where phone not like "+%";

update Leads 
set phone = concat("+", phone)
where phone not like "+%";


select L1.name, L1.phone, L2.phone,
IF(L1.email is NULL, L2.email, L1.email) as email,
IF(L1.cotizacion is NULL, L2.cotizacion, L1.cotizacion) as cotizacion,
L1.asesor
from Leads L1
inner join Leads L2
    on  L1.phone like "+%"
    and L2.phone not like "+%"
    and L1.phone = concat("+", L2.phone);

UPDATE Leads L1
INNER JOIN Leads L2
    ON L1.phone LIKE '+%' 
    AND L2.phone NOT LIKE '+%' 
    AND L1.phone = CONCAT('+', L2.phone)
SET 
    L1.email = IF(L1.email IS NULL, L2.email, L1.email),
    L1.cotizacion = IF(L1.cotizacion IS NULL, L2.cotizacion, L1.cotizacion),
    L1.asesor = L2.asesor;


select L.phone, L2.phone from Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
inner join Leads L2
    on L.phone not LIKE '+%'
    and L2.phone like "+%"
    and L2.phone = CONCAT("+", L.phone);

UPDATE Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
INNER JOIN Leads L2
    ON L.phone NOT LIKE '+%'
    AND L2.phone LIKE '+%'
    AND L2.phone = CONCAT('+', L.phone)
SET C.lead_phone = L2.phone;

delete L2
FROM Leads L1
INNER JOIN Leads L2
    ON L1.phone = CONCAT('+', L2.phone)
    AND L1.phone LIKE '+%'
    AND L2.phone NOT LIKE '+%';

select * from Communication where lead_phone not like "+%";

select * from Leads where phone not like "+%";

/* Actualizar los telefonos de todos los leads, aunque no estenrepetidos */
SET FOREIGN_KEY_CHECKS = 0;
UPDATE Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
SET C.lead_phone = CONCAT('+', C.lead_phone)
WHERE L.phone NOT LIKE '+%';

UPDATE Leads
SET phone = CONCAT('+', phone)
WHERE phone NOT LIKE '+%';
SET FOREIGN_KEY_CHECKS = 1;

/*Leads por día
Esta query no pone en 0 los días que no tienen nada*/
SELECT 
    DATE(C.created_at) AS date,
    COUNT(C.id) AS new_leads_count
FROM 
    Communication C
WHERE 
    C.new_lead = TRUE
GROUP BY 
    DATE(C.created_at)
ORDER BY 
    DATE(C.created_at);

/* agregar el 1 y el 9 dsp de +52 y +54 */
select L.phone, L2.phone from Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
inner join Leads L2
    on L.phone not LIKE '+521%'
    and L2.phone like "+52%"
    and L2.phone = CONCAT("+", L.phone);

SELECT 
    name, 
    phone,
    CASE
            WHEN phone LIKE '+54%' THEN CONCAT('+549', SUBSTRING(phone, 4))
            WHEN phone LIKE '+52%' THEN CONCAT('+521', SUBSTRING(phone, 4))
            ELSE phone
    END AS normalized_phone
FROM Leads
WHERE 
    phone LIKE '+54%' AND SUBSTRING(phone, 4, 1) != '9' OR 
    phone LIKE '+52%' AND SUBSTRING(phone, 4, 1) != '1';

select lead_phone, phone, SUBSTRING(phone, 1, 3) as xd
FROM Communication
inner join Leads
    ON SUBSTRING(phone, 5) = SUBSTRING(lead_phone, 4)
    AND SUBSTRING(phone, 1, 3) = SUBSTRING(lead_phone, 1, 3)
    AND phone <> lead_phone;

UPDATE Communication
inner join Leads
    ON SUBSTRING(phone, 5) = SUBSTRING(lead_phone, 4)
    AND SUBSTRING(phone, 1, 3) = SUBSTRING(lead_phone, 1, 3)
    AND phone <> lead_phone
SET lead_phone = phone;


select L1.phone, L2.phone
from Leads L1
inner join Leads L2
    ON SUBSTRING(L1.phone, 5) = SUBSTRING(L2.phone, 4)
    AND SUBSTRING(L1.phone, 1, 3) = SUBSTRING(L2.phone, 1, 3)
    AND L1.phone <> L2.phone;

delete L2
from Leads L1
inner join Leads L2
    ON SUBSTRING(L1.phone, 5) = SUBSTRING(L2.phone, 4)
    AND SUBSTRING(L1.phone, 1, 3) = SUBSTRING(L2.phone, 1, 3)
    AND L1.phone <> L2.phone;

SET FOREIGN_KEY_CHECKS = 0;
UPDATE Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
SET C.lead_phone = CONCAT('+', C.lead_phone)
WHERE L.phone NOT LIKE '+%';
UPDATE Leads
SET phone = (
    CASE 
        WHEN phone LIKE '+54%' AND SUBSTRING(phone, 4, 1) != '9' THEN
            CONCAT("+549", SUBSTRING(phone, 4))
        WHEN phone LIKE '+52%' AND SUBSTRING(phone, 4, 1) != '1' THEN
            CONCAT("+521", SUBSTRING(phone, 4))
        ELSE phone -- Mantiene el número de teléfono sin cambios si no coincide con ninguna condición
    END
);
SET FOREIGN_KEY_CHECKS = 1;


/* Lista de leads (agregado mensaje)*/
SELECT 
    C.id,
    DATE_SUB(C.created_at, INTERVAL 6 HOUR) as "created_at",
    C.new_lead,
    IF(S.type = "property", P.portal, S.type) as "fuente",
    L.name, 
    L.phone,
    L.email,
    M.text
FROM Communication C
INNER JOIN Leads L
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id
INNER Join Message M
    ON M.id_communication = C.id
ORDER BY C.id DESC LIMIT 1;




/* Leads que recibieron broadcast 2ab6ccd7-0527-4592-9da8-e9380ff22c8a*/
select * from Action 
where flow_uuid = "2ab6ccd7-0527-4592-9da8-e9380ff22c8a";
/* Leads que se les envio un template dsp de recibir el broadcast de clientes*/
select * from (
    select firstAction.lead_phone
    from Action as firstAction
    inner join Action otherAction
        on firstAction.lead_phone = otherAction.lead_phone
        and otherAction.sended_at > firstAction.sended_at
        and firstAction.flow_uuid = "2ab6ccd7-0527-4592-9da8-e9380ff22c8a"
    group by firstAction.lead_phone
) as Interesados
inner join Leads 
on phone = Interesados.lead_phone
INTO OUTFILE '/data/clientes.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

/* leads que recibieron broadcast e25b6a30-f25d-4386-85e7-f8a2c6b8c35f*/
select * from action 
where flow_uuid = "e25b6a30-f25d-4386-85e7-f8a2c6b8c35f";
/* Leads que se les envio un template dsp de recibir el broadcast de asesores*/
select * from (
    select firstAction.lead_phone
    from Action as firstAction
    inner join Action otherAction
        on firstAction.lead_phone = otherAction.lead_phone
        and otherAction.sended_at > firstAction.sended_at
        and firstAction.flow_uuid = "e25b6a30-f25d-4386-85e7-f8a2c6b8c35f"
    group by firstAction.lead_phone
) as Interesados
inner join Leads 
on phone = Interesados.lead_phone
INTO OUTFILE '/data/asesores.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

/* leads que recibieron broadcast 18da18ac-d065-4404-9945-e7363719311d*/
select * from Action 
where flow_uuid = "18da18ac-d065-4404-9945-e7363719311d";

select * from Action 
where flow_uuid = "526ea665-a204-4333-99e7-02f89a404f53";

select * from Message
where text like "Quiero más información";

select * from Message
where text like "No estoy interesado";

/* Personas que enviaron un mensaje luego de recibir el broadcast */
select M.created_at, id_communication, C.lead_phone, text as 'primer mensaje'
from Message M
inner join Communication C
    on C.id = M.id_communication
inner join Action A
    on M.created_at > A.sended_at
    and C.lead_phone = A.lead_phone
    and flow_uuid = "18da18ac-d065-4404-9945-e7363719311d"
group by M.created_at, id_communication
INTO OUTFILE '/data/respuestas-broadcast-video.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

/* Campos de busquedas que no son null */
select lead_phone, zones, mt2_terrain, mt2_builded, baths, rooms 
from Communication
where 
(zones is not NULL and zones <> "") OR
(mt2_builded is not NULL and mt2_builded <> "") OR
(mt2_terrain is not NULL and mt2_terrain <> "") OR
(baths is not NULL and baths <> "") OR
(rooms is not NULL and rooms <> "");


    id INT NOT NULL AUTO_INCREMENT,

    title VARCHAR(256) NOT NULL,
    price VARCHAR(32) NOT NULL,
    currency CHAR(5) NOT NULL,
    description VARCHAR(512) NOT NULL, 
    type VARCHAR(32) NOT NULL,
    antiquity INT NOT NULL,
    parkinglots INT DEFAULT NULL,
    bathrooms INT DEFAULT NULL,
    half_bathrooms INT DEFAULT NULL,
    rooms INT DEFAULT NULL,
    operation_type ENUM("sale", "rent") NOT NULL,
    m2_total INT,
    m2_covered INT,
    video_url VARCHAR(256) DEFAULT NULL,
    virtual_route VARCHAR(256) DEFAULT NULL,

    /* Ubication fields */
    state VARCHAR(128) NOT NULL,
    municipality VARCHAR(128) NOT NULL,
    colony VARCHAR(128) NOT NULL,
    neighborhood VARCHAR(128) DEFAULT NULL,
    street VARCHAR(256) NOT NULL,
    number VARCHAR(32) NOT NULL,
    zip_code VARCHAR(32) NOT NULL,


/* Query to migrate Communications to new properties */
select P.*, C.zones
from Communication C
inner join Source S
    on S.id = C.source_id
    and S.type = "property"
inner join Property P
    on P.id = S.property_id
group by P.portal_id;


/* Obtain the published properties with the publication ids */
select 'id','title','price','currency','type','antiquity','parkinglots','bathrooms','half_bathrooms','rooms','operation_type','m2_total','m2_covered','video_url','virtual_route','address','lat','lng','url'
UNION ALL
select id,title,price,currency,type,antiquity,parkinglots,bathrooms,half_bathrooms,rooms,operation_type,m2_total,m2_covered,video_url,virtual_route,address,lat,lng,
CONCAT(Portal.base_url, Published.publication_id) as url
from PublishedProperty Published
inner join PortalProp Prop
on Prop.id = Published.property_id
inner join Portal
on Portal.name = Published.portal
where status = 'published'
INTO OUTFILE '/data/properties.csv';

/* Get all the properties in queue */
select * from PublishedProperty
where status = 'in_queue';

update PublishedProperty
set status = 'in_queue'
where status = 'failed';

update PublishedProperty
set status = 'in_queue'
where status = 'failed';

update Asesor
set email = "control.general@rebora.com.mx", name = "Control General"
where phone = "5213314299454";

/* Publicar todo en inmuebles24 */ 
select * from PublishedProperty
where portal = "inmuebles24";

delete from PublishedProperty
where portal = "inmuebles24";

insert into PublishedProperty (property_id, status, portal)
select id, "in_queue", "inmuebles24" from PortalProp;

select id from PortalProp
where address = "La Rioja, Los Gavilanes, 45645 Los Gavilanes, Jal.";

/* obtener cantidad de propiedades correctamente publicadas */
select COUNT(*)
from PublishedProperty
where portal = "inmuebles24"
and status = "in_progress";

/* UPDATE */
/* UPDATE */
update PublishedProperty
set status = "in_queue"
where status = "in_progress"
and portal = "inmuebles24";
/* UPDATE */
/* UPDATE */

EL PALOMAR
LA RIOJA
LAGO NOGAL
LAS VILLAS
LOS GAVILANES
PROVENZA
PUNTO SUR
MARTIN DEL TAJO

SELECT id 
FROM PortalProp 
WHERE title LIKE '%PUNTO SUR%';

UPDATE PublishedProperty
SET status = 'in_queue'
WHERE portal = 'inmuebles24'
AND property_id IN (
    SELECT id 
    FROM PortalProp 
    WHERE title LIKE '%MARTIN DEL TAJO%'
);

select * from PublishedProperty
where property_id = 3
and portal = "casasyterrenos";

select * from PublishedProperty
where portal = "casasyterrenos";

update PublishedProperty
select * from PublishedProperty
where status = "failed"
and portal = "casasyterrenos";

select * from PublishedProperty
where publication_id is not null
and portal = "casasyterrenos";
