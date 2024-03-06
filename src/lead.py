class Lead():
    def __init__(self):
        self.fuente = ""
        self.fecha_lead = ""
        self.id = ""
        self.fecha = ""
        self.nombre = ""
        self.link = ""
        self.telefono = ""
        self.email = ""
        self.message = ""
        self.asesor = {
            "name": "",
            "phone": ""
        }
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
            "cantidad_anuncios": "",
            "contactos": "",
            "inicio_busqueda": "",
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
