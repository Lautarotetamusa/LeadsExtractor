if __name__ == "__main__":
    from src.sheets import Sheet
    from src.sheets import Gmail
    from src.logger import Logger
    logger = Logger("Auth")
    
    gmail = Gmail({'email': 'Prueba'}, logger)
    gmail.send_message('mensaje de prueba', 'soypiki@gmail.com', 'Subject de prueba')

    sheets = Sheet(logger, "mapping.json")
    sheets.write(["prueba", "prueba"])
