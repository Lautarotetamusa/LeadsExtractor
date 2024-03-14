import phonenumbers

if __name__ == "__main__":
    number = "+5493415854220"
    number = "+523324944591"
    number = "+5217441036217"
    number = "3314156138" #numero que llega desde inmuebles24, lo tengo que parsear como "MX" para que ande
    number = "+5213319466986" #numero wa_id mexicano
    number = "523327919473" #Propiedades.com
    number = "523411234567" #IVR CALL

    #Para poder parsear los numeros bien tienen que tener un + adelante
    #        +52 3327919473  # -> E164
    x = phonenumbers.parse(number, "MX")
    #x = phonenumbers.parse(number, None)
    print(x)
    parsed_number = phonenumbers.format_number(x, phonenumbers.PhoneNumberFormat.E164)
    print(parsed_number)
