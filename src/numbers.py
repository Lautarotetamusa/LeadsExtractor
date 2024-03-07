import phonenumbers
from src.logger import Logger

def parse_number(logger: Logger, phone: str) -> None | str:
    logger.debug("Parseando numero")
    try:
        number = phonenumbers.parse(phone, "MX")
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        logger.success("Numero obtenido: " + parsed_number)
        return parsed_number
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + phone)
        return None

def parse_wpp_number(logger: Logger, phone: str) -> str:
    logger.debug("Parseando numero")
    try:
        number = phonenumbers.parse(phone, None)
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        logger.success("Numero obtenido: " + parsed_number)
        return parsed_number
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + phone)
        return phone
