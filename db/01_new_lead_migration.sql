alter table Communication add column new_lead BOOLEAN NOT NULL DEFAULT true;

DROP PROCEDURE IF EXISTS communicationList;
DELIMITER //
CREATE PROCEDURE communicationList ()
    BEGIN
        SELECT 
            DATE_FORMAT(C.created_at, "%d/%m/%Y") as "Fecha extraccion", 
            C.lead_date as "Fecha lead",
            IF(C.new_lead, "", "Duplicado") as "Estado",
            A.name as "Asesor asignado", S.type, P.portal, L.name, C.url, L.phone, L.email,
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
    END;
//
DELIMITER ;

update Communication
set new_lead = false
where lead_phone IN (
    select lead_phone from Communication
    group by lead_phone
    having count(lead_phone) > 1
);
