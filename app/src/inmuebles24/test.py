import json
import requests

if __name__ == "__main__":

    url = 'https://www.inmuebles24.com/panel/avisos'
    apikey = 'f569290d8d48d2d19a288ff2a95cad9b6679ab4d'
    js_instructions = [
        {"wait": 2500},
    ]
    params = {
        'url': url,
        'apikey': apikey,
        'js_render': 'true',
        'js_instructions': json.dumps(js_instructions),
        'wait_for': '.price',
        'screenshot': 'true',
    }
    cookies = {
		"hideWelcomeBanner": "true",
		"IDusuario": "57036554",
		"JSESSIONID": "E82EF349012EC53227E0B8B3C8A4E213",
		"mousestats_vi": "fc98ac1a434c881e4202",
		"owneremail": "sergio.esqueda@inmuebles24.com",
		"ownerIdempresa": "50796870",
		"ownername": "Sergio Esqueda",
		"phoneCall": "2025-03-25T02:36:42-03:00",
		"reputationModalLevelSeen": "true",
		"reputationModalLevelSeen2": "true",
		"reputationTourLevelSeen": "true",
		"sessionId": "4fa2004f-57bd-4a0c-8601-9048d0e62771",
		"showWelcomePanelStatus": "true",
		"usuarioFormApellido": "Residences",
		"usuarioFormEmail": "control.general@rebora.com.mx",
		"usuarioFormId": "57036554",
		"usuarioFormNombre": "RBA",
		"usuarioFormTelefono": "523341690109",
		"usuarioIdCompany": "50796870",
		"usuarioLogeado": "control.general@rebora.com.mx",
		"usuarioPublisher": "true",
		"whatsapp": "2025-03-24T23:36:39-03:00"
	}
    res = requests.get('https://api.zenrows.com/v1/', params=params, cookies=cookies)
    print(res.status_code)
    if not res.ok:
        print("error")
        print(res.json())
        exit(1)

    with open("test.png", "wb") as f:
        f.write(res.content)
