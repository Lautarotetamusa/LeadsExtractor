import phonenumbers
from src.logger import Logger

def parse_number(logger: Logger, phone: str, code=None) -> str | None:
    logger.debug("Parseando numero, code: "+str(code))
    try:
        number = phonenumbers.parse(phone, code)
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        logger.success("Numero obtenido: " + parsed_number)
        return parsed_number
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + phone)
        return None
