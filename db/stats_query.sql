SELECT 
    COUNT(*) AS total,
    SUM(CASE WHEN new_lead is True THEN 1 ELSE 0 END) AS new_leads,
    SUM(CASE WHEN S.type = 'whatsapp' THEN 1 ELSE 0 END) AS whatsapp,
    SUM(CASE WHEN S.type = 'ivr' THEN 1 ELSE 0 END) AS ivr,
    SUM(CASE WHEN portal = 'inmuebles24' THEN 1 ELSE 0 END) AS inmuebles24,
    SUM(CASE WHEN portal = 'lamudi' THEN 1 ELSE 0 END) AS lamudi,
    SUM(CASE WHEN portal = 'casasyterrenos' THEN 1 ELSE 0 END) AS casasyterrenos,
    SUM(CASE WHEN portal = 'propiedades' THEN 1 ELSE 0 END) AS propiedades
SELECT C.*, S.* 
FROM Communication C
INNER JOIN Source S
    ON C.source_id = S.id
LEFT JOIN Property P
    ON S.property_id = P.id
WHERE created_at BETWEEN '2025-01-01' AND '2025-06-03'
AND S.type = "whatsapp";


select count(*), new_lead, IF(S.type = "property", P.portal, S.type) as "fuente" 
FROM Communication C
INNER JOIN Source S
    ON C.source_id = S.id
LEFT JOIN Property P
    ON S.property_id = P.id
WHERE created_at BETWEEN '2025-01-01' AND '2025-06-03'
group by fuente, new_lead;

/* IF(S.type = "property", P.portal, S.type) as "fuente" */
