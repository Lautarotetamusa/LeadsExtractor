/* Lista de LEADS tal cual estaba en el google sheets*/
SELECT C.created_at as "Fecha extraccion", A.name as "Asesor asignado", S.type, P.portal, L.name, L.phone, L.email, P.title, P.url, P.price, P.ubication
FROM Communication C
INNER JOIN Leads L 
    ON C.lead_phone = L.phone
INNER JOIN Source S
    ON C.source_id = S.id
INNER JOIN Asesor A
    ON L.asesor = A.phone
LEFT JOIN Property P
    ON S.property_id = P.id;
