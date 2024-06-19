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

/* Lista de telefono unicos (leads distintos) */
select distinct lead_phone from Communications; /* 17 jun -> 500 */

select count(*) from Leads; /*17 jun 2024 -> 637 */
