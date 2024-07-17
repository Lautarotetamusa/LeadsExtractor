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
        IFNULL(P.portal_id, "") as "Portal_id", 
        IFNULL(P.title, "") as "Title", 
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


