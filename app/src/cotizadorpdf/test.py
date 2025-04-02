from datetime import datetime
import os

if __name__ == "__main__":
    with open("src/cotizadorpdf/output.html", "r") as f:
        html_content = f.read()
    timestamp_str = datetime.now().strftime("%Y-%m-%d%H:%M:%S")
    pdf_filename = os.path.join("pdfs", "cotizacion" + timestamp_str +".pdf")

    from pyhtml2pdf import converter

    path = os.path.abspath('index.html')
    converter.convert("file:///home/teti/Archivos/Scrapers/LeadsExtractor/app/src/cotizadorpdf/output.html", 'sample.pdf')
    exit(0)
