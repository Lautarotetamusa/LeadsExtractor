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
    'Cookie': 'G_ENABLED_IDPS=google; g_state={"i_p":1720379279209,"i_l":3}; _ga_G9RXEV9M1M=GS1.1.1721843204.33.1.1721843892.0.0.0; _ga=GA1.1.267251478.1700080378; _clck=1ulkcfu%7C2%7Cfnq%7C0%7C1433; GCLB=CJWrpeT_67jfMhAD; _clsk=df2o7d%7C1721817234781%7C21%7C1%7Cv.clarity.ms%2Fcollect; __cf_bm=nJvWc1BtBcxgh5cXe25bqWRbj3jumrvkJUDs3h1Zge4-1721854000-1.0.1.1-Il1nlCcOBk5i9MWzBxFwCFWjwy3VIRnCtcm5_d.sSm5ML6XN3cHqyG3ANp_PmpofrYTukATNoVeEzknca_vcZw; cf_clearance=9Me.j0SYTCj8IeDVcwfKKvzZWJWF90j6jHbfGx9ib5I-1721854002-1.0.1.1-YrXBa2_6CVRf1HK97gceUHFVqbMgOeWkDAzLbIM1H22kgE.2DLVmhWozbSAneu3lL6l1ANliDZgmxb4Nfb0xMQ; PHPSESSID=r86if12e8elks501mc9kn3o4r4; user_fingerprint=ae1a7bd443336aa3f26814bd37fe0dbbb39b2836s%3A128%3A%224a2e1c6ae0bd2bec75f46a4e6e4caa1ffe5e222bb78edf34e761cb5bebfdd943f24ee2d833ae7d4b001e2ebd20344c5f95f4bb1a908c0bb3b3db5848b30a3407%22%3B; cf_chl_rc_m=2; source=6bd8da202333c9d69fdf6752342e4b92c7cf7c6ds%3A4%3A%22null%22%3B; campaign=6bd8da202333c9d69fdf6752342e4b92c7cf7c6ds%3A4%3A%22null%22%3B',
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
print(f"Response Text: {response.text}")
