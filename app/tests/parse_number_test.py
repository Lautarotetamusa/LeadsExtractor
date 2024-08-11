import phonenumbers

def parse(num: str) -> str:
    try:
        x = phonenumbers.parse(number, None)
    except Exception as e:
        x = phonenumbers.parse(number, "MX")

    print(x)
    parsed_number = phonenumbers.format_number(x, phonenumbers.PhoneNumberFormat.E164)
    print(parsed_number)

if __name__ == "__main__":
    number = "+5493415854220"
    number = "+523324944591"
    number = "+5217441036217"
    number = "3314156138" #numero que llega desde inmuebles24, lo tengo que parsear como "MX" para que ande
    parse(number)
    number = "+5213314156138" #numero wa_id mexicano
    parse(number)
    #number = "523327919473" #Propiedades.com
    #number = "523411234567" #IVR CALL Mexico
    #number = "525493415854220" #IVR CALL Argentina
    #number = "3344556677" #Casas y terrenos
    #number = "+524151510540" #Cassa y terrenos

    #Para poder parsear los numeros bien tienen que tener un + adelante
    #        +52 3327919473  # -> E164
