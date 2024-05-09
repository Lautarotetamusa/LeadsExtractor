alter table Asesor
    add column email varchar(128) not null;

update Asesor set email="diego.rubio@rebora.com.mx" where phone = "5213317186543";
update Asesor set email="onder.sotomayor@rebora.com.mx" where phone="5213318940377";
update Asesor set email="maggie.escobedo@rebora.com.mx" where phone="5213314299454";
update Asesor set email="aldo.salcido@rebora.com.mx" where phone="5213322563353";
update Asesor set email="brenda.diaz@rebora.com.mx" where phone="5213313420733";
