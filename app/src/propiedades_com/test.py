import requests

sender = {
    "name": "juani",
    "phone": "3415854221",
    "email": "juanipozzi@gmail.com"
}
import requests

# URL de destino
url = "https://propiedades.com/messages/send?pagina=1"

# Datos del formulario
data = {
    "lead_type": "2",
    "is_ficha": "0",
    "id": "26813758",
    "referenceAddress": "",
    "referral": "null",
    "ssr": "1",
    "ContactForm[name]": "test",
    "ContactForm[lastname]": "",
    "ContactForm[phone]": "3415854221",
    "ContactForm[email]": "test@test.com",
    "ContactForm[acceptTerms]": "1",
    "ContactForm[registerUser]": "false",
    "ContactForm[lead_source]": "5",
    "ContactForm[body]": "test"
}

import requests

url = "https://propiedades.com/messages/send?pagina=1"

# Datos del formulario
data = {
    'lead_type': '2',
    'is_ficha': '0',
    'id': '26813758',
    'referenceAddress': '',
    'referral': 'null',
    'ssr': '1',
    'ContactForm[name]': 'test',
    'ContactForm[lastname]': '',
    'ContactForm[phone]': '3415854221',
    'ContactForm[email]': 'test@test.com',
    'ContactForm[acceptTerms]': '1',
    'ContactForm[registerUser]': 'false',
    'ContactForm[lead_source]': '5',
    'ContactForm[body]': 'test'
}

# Encabezados de la petición
headers = {
    'Accept': 'application/json, text/plain, */*',
    'Accept-Encoding': 'gzip, deflate, br',
    'Accept-Language': 'en-US,en;q=0.5',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'Content-Type': 'multipart/form-data; boundary=---------------------------42269429314279474156221563304',
    'Cookie': 'G_ENABLED_IDPS=google; g_state={"i_p":1720379279209,"i_l":3}; _ga_G9RXEV9M1M=GS1.1.1732879179.40.1.1732880400.0.0.0; _ga=GA1.1.267251478.1700080378; _clck=1ulkcfu%7C2%7Cfra%7C0%7C1433; __cf_bm=tbGKmoI6STuIdpr8JHIcSaZ6ULdYnfSbaT8FXdZRdf8-1732891018-1.0.1.1-t8WjhqvtDQTXP1h2GzFYXpgianSNAJrsN5xqdmgiwp7FJYHcvJW.L4lXYHiufYhTDOymOx8IeJ0Y5tKjmL0MpA; cf_clearance=im3ffTkT4vKIedCJRw8S5Orylh5VVIFPHhV3zM.EAtU-1732889982-1.2.1.1-_4Q3Y6nyAoy.FExJ1sd84gGiagm35CfSn3OcA8U8U9k8.TeDMLAtR.IP8HgGJJnIsZnZvjo12XKnsypDMS9iRLxD_NVHUmDwMKlIFcNQjgalTc7mk7Aby4wO8BInkrOpguQiHYzN0J3dy5jr2mHn.Xsz9ryqMg7DWU8NNxdGBsjsY0r0c4DPDR4uKaQdRBDDAnu2FnUue4YYyPV7Kv5aHUmk9QS4JnBiF03kdvcDEN5q8UPLDKe2790RhlrbQ7NVIe8ewFIlqwvTMJTrrEW4caE4E8MD.xEU6DVSef5UhpAK4OrRqaumsiOGjb6tCz5LbeBioPi6mvXUb.Gg306qN1yN_93quqZGDhvNoEM5ZrbNWtazmM6VWDLtOUKD2_1m; _clsk=1p7ldsg%7C1732879181563%7C1%7C1%7Ci.clarity.ms%2Fcollect; PHPSESSID=11hb3ml2iatv3s5as58tb6tps7; user_fingerprint=a905ee53977ef33ff2730d5b3b2eeab556915d2ds%3A128%3A%227f0d0fb2396a94d23ed5c8dc516c53f2da971e330df3a09520006c3c29cda7d7fcf653378115d9cf9bf4fa517d1b7299d476a5f173d20c2c872fc945667ca5c9%22%3B; source=6bd8da202333c9d69fdf6752342e4b92c7cf7c6ds%3A4%3A%22null%22%3B; campaign=6bd8da202333c9d69fdf6752342e4b92c7cf7c6ds%3A4%3A%22null%22%3B; ContactActivity=2196b383d58d3d564f817582a025d4fc6c1790a1s%3A136%3A%22%7B%22name%22%3Anull%2C%22lastname%22%3Anull%2C%22email%22%3Anull%2C%22phone%22%3Anull%2C%22properties%22%3A%5B%5D%2C%22typePropertiesContact%22%3A%5B%5D%2C%22zones%22%3A%5B%22%402024-11-29%22%2C%22%402024-11-29%22%5D%7D%22%3B',
    'cucu': '1',
    'Host': 'propiedades.com',
    'Origin': 'https://propiedades.com',
    'Pragma': 'no-cache',
    'Referer': 'https://propiedades.com/zapopan/terrenos-comerciales-venta?pagina=1',
    'Sec-Fetch-Dest': 'empty',
    'Sec-Fetch-Mode': 'cors',
    'Sec-Fetch-Site': 'same-origin',
    'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0'
}

# Enviar la solicitud POST
response = requests.post(url, data=data, headers=headers)

# Imprimir el resultado
print(f"Status Code: {response.status_code}")
print(f"Response Text: {response.json()}")
