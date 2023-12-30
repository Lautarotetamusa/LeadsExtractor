if __name__ == "__main__":
    from src.sheets import Sheet
    from src.sheets import Gmail
    from src.logger import Logger
    from email.mime.application import MIMEApplication
    with open('messages/attachment.pdf', 'rb') as pdf_file:
        attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
        attachment.add_header('Content-Disposition', 'attachment', 
              filename='Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'
        )

    logger = Logger("Auth")

    #sheets = Sheet(logger, "mapping.json")
    #sheets.write(["prueba", "prueba"])
    
    gmail = Gmail({'email': 'Prueba'}, logger)
    gmail.send_message('mensaje de prueba', 'Subject de prueba', 'soypiki@gmail.com', attachment)

    """
https://accounts.google.com/o/oauth2/auth?response_type=code&client_id=744314508953-pi57n5vmm4kjjdktcr9abad4f8gd64qo.apps.googleusercontent.com&redirect_uri=https://d733-2800-40-32-3f9-ab99-cd0a-3c87-fe4a.ngrok-free.app&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fspreadsheets+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.send&state=LbT6jsCmE9y0NEubF09MW4ISURkwBT&access_type=offline
    """
