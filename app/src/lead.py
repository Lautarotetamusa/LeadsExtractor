import json

class Lead():
    def __init__(self):
        self.fuente = ""
        self.fecha_lead = ""
        self.lead_id = ""
        self.nombre = ""
        self.link = ""
        self.telefono = ""
        self.email = ""
        self.message = ""
        self.estado = ""
        self.cotizacion = ""
        self.propiedad = {
            "id": "",
            "titulo": "",
            "link": "",
            "precio": "",
            "ubicacion": "",
            "tipo": "",
        }
        self.busquedas = {
            "zonas": "",
            "tipo": "",
            "presupuesto": "",
            "cantidad_anuncios": None,
            "contactos": None,
            "inicio_busqueda": None,
            "total_area": "",
            "covered_area": "",
            "banios": "",
            "recamaras": "",
        }
    def set_args(self, args: dict):
        self.__dict__ = {**self.__dict__, **args}
    def set_propiedad(self, args: dict):
        self.propiedad = {**self.propiedad, **args}
    def set_busquedas(self, args: dict):
        self.busquedas = {**self.busquedas, **args}
    def set_asesor(self, args: dict):
        self.asesor = {**self.asesor, **args}

    #Validamos los campos que tienen que estar si o si en infobip
    def validate(self):
        assert self.telefono != None and self.telefono != '', "El lead no tiene telefono"
        assert self.fuente != None and self.fuente != '', f"El lead {self.telefono} no tiene fuente"
        assert self.fecha_lead != None and self.fecha_lead != '', f"El lead {self.telefono} no tiene fecha_lead"
        assert self.asesor['name'] != None and self.asesor['name'] != '', f"El lead {self.telefono} no tiene asesor asignado"
        assert self.asesor['phone'] != None and self.asesor['phone'] != '', f"El lead {self.telefono} no tiene asesor asignado"
        assert self.cotizacion != None and self.cotizacion != '', f"El lead {self.telefono} no tiene cotizacion"

    #No funciona recursivamente, si por ejemplo asesor solo tiene nombre devolverá el asesor con nombre y el telefono vacio.
    def get_no_empty_values(self) -> dict[str, str]:
        return {k: v for k, v in self.__dict__.items() if v != "" and v != None and type(v) is not dict}

    def print(self):
        print(json.dumps(self.__dict__, indent=4))

if __name__ == "__main__":
    lead = Lead()
    lead.set_args({
        "propiedad": {
            "titulo": "test"
        },
        "fuente": "fuentee",
        "nombre": "nombre"
    })
    print(lead.__dict__)
