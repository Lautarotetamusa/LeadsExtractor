alter table Portal
add column base_url VARCHAR(64) NOT NULL DEFAULT " _ ",
add check (base_url <> "");

update Portal 
set base_url = "https://propiedades.com/inmuebles/-"
where name = "propiedades";

update Portal 
set base_url = "https://www.lamudi.com.mx/detalle/"
where name = "lamudi";

update Portal 
set base_url = "https://www.casasyterrenos.com/propiedad/-"
where name = "casasyterrenos";
