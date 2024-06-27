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
