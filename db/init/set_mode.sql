/*Esto es necesario porque sino al intentar insertar un string que es un UUID, mysql lo toma como un int y no funciona*/
SET GLOBAL sql_mode='';
