import requests
from api_request import post

APIKEY = "8fc62499ac97fc62f0d5081f478d5ede"

def webscrappingapi(url):
    site_url = "https://www.inmuebles24.com/"

    cookies = {
		#"__cf_bm": "o3ZbLYEzjLlsH6f38XdNJRCjoKibAUSUwGWp9E3ZxCw-1698443451-0-AQdx4izAxcREtaKwakfE3l7AWYswCjZL97xdRwf3uuCoY4MC7ZRlUtw59rlKfTBkEGbYhJcD+1eOf5sLo/dg2gvXYd7nyepMnnNZBzSRAC2v",
		#"cf_clearance": "za9dVMWY0XVeHHhQdFyv60WmEOlY_ogtjaeAa8pt2BA-1698436283-0-1-115e9ea5.ba12bc.465fc145-160.2.1698436283",
		"g_state": "{\"i_p\":1699030291139,\"i_l\":3}",
		#"hashKey": "4IVc5it99duNGjgDajK2gWBqUKoYpMO62d6jqt/b2c7zAfyJwpCUsaJvdQqD1oJORGYcG6pBna3NhUqFx6GGZZuoCpoWFh3Mvx9c",
		"JSESSIONID": "7C3CE6AC396608531EC1FD0309BDC85A",
		#"sessionId": "1d9bfb30-7504-431e-86c0-2a97c109938c",
		"usuarioFormApellido": "Aquitectos",
		"usuarioFormEmail": "ventas.rebora@gmail.com",
		"usuarioFormTelefono": "3313420733",
		"usuarioFormId": "57036554",
		"usuarioPublisher": "true",
		"usuarioIdCompany": "50796870",
		"hideWelcomeBanner": "true",
		"usuarioFormNombre": "Rebora",
		"usuarioLogeado": "ventas.rebora@gmail.com"
	}

    payload = {
        'api_key': "EBY0g5y21tU7AdVvSEGDWWg5Tok9llLQ",
        'url': site_url,
        'session': 1,
        'render_js': 1,
        #'keep_headers': 1,
        #'cookies': cookies,
        'stealth_mode': 1,
        'device': 'desktop',

    }

    headers = [
			{
				"name": "Accept",
				"value": "*/*"
			},
			{
				"name": "Accept-Encoding",
				"value": "gzip, deflate, br"
			},
			{
				"name": "Accept-Language",
				"value": "en-US,en;q=0.5"
			},
			{
				"name": "Connection",
				"value": "keep-alive"
			},
			{
				"name": "content-type",
				"value": "application/json;charset=UTF-8"
			},
			{
				"name": "Cookie",
				"value": "cf_clearance=za9dVMWY0XVeHHhQdFyv60WmEOlY_ogtjaeAa8pt2BA-1698436283-0-1-115e9ea5.ba12bc.465fc145-160.2.1698436283; sessionId=1d9bfb30-7504-431e-86c0-2a97c109938c; g_state={\"i_p\":1699030291139,\"i_l\":3}; __cf_bm=o3ZbLYEzjLlsH6f38XdNJRCjoKibAUSUwGWp9E3ZxCw-1698443451-0-AQdx4izAxcREtaKwakfE3l7AWYswCjZL97xdRwf3uuCoY4MC7ZRlUtw59rlKfTBkEGbYhJcD+1eOf5sLo/dg2gvXYd7nyepMnnNZBzSRAC2v; JSESSIONID=7C3CE6AC396608531EC1FD0309BDC85A; hashKey=4IVc5it99duNGjgDajK2gWBqUKoYpMO62d6jqt/b2c7zAfyJwpCUsaJvdQqD1oJORGYcG6pBna3NhUqFx6GGZZuoCpoWFh3Mvx9c; usuarioLogeado=ventas.rebora@gmail.com; usuarioFormNombre=Rebora; usuarioFormEmail=ventas.rebora%40gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioFormNombre=Rebora; usuarioFormApellido=Aquitectos; usuarioFormEmail=ventas.rebora@gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioPublisher=true; usuarioIdCompany=50796870; hideWelcomeBanner=true"
			},
			{
				"name": "email",
				"value": ""
			},
			{
				"name": "Host",
				"value": "www.inmuebles24.com"
			},
			{
				"name": "Referer",
				"value": "https://www.inmuebles24.com/panel/interesados/182835939"
			},
			{
				"name": "Sec-Fetch-Dest",
				"value": "empty"
			},
			{
				"name": "Sec-Fetch-Mode",
				"value": "cors"
			},
			{
				"name": "Sec-Fetch-Site",
				"value": "same-origin"
			},
			{
				"name": "sessionId",
				"value": "1d9bfb30-7504-431e-86c0-2a97c109938c"
			},
			{
				"name": "TE",
				"value": "trailers"
			},
			{
				"name": "User-Agent",
				"value": "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/111.0"
			},
			{
				"name": "x-panel-portal",
				"value": "24MX"
			}
		]
    obj_headers = {h["name"]: h["value"] for h in headers}

    res = requests.get('https://api.webscrapingapi.com/v1', params=payload)
    if res.status_code != 200:
        print("Status code is not 200: ", res.status_code)
        print(res.text)
        return None

    print(res.text)
    return res.json()

def get_req(url):
    site_url = "https://www.inmuebles24.com/"

    payload = {
        'api_key': APIKEY, 
        'url': site_url,
        'session_number': 1,
        #'keep_headers': True,
        #'render': True
        #'ultra_premium': True
    }

    cookies = {
		#"__cf_bm": "o3ZbLYEzjLlsH6f38XdNJRCjoKibAUSUwGWp9E3ZxCw-1698443451-0-AQdx4izAxcREtaKwakfE3l7AWYswCjZL97xdRwf3uuCoY4MC7ZRlUtw59rlKfTBkEGbYhJcD+1eOf5sLo/dg2gvXYd7nyepMnnNZBzSRAC2v",
		#"cf_clearance": "za9dVMWY0XVeHHhQdFyv60WmEOlY_ogtjaeAa8pt2BA-1698436283-0-1-115e9ea5.ba12bc.465fc145-160.2.1698436283",
		"g_state": "{\"i_p\":1699030291139,\"i_l\":3}",
		#"hashKey": "4IVc5it99duNGjgDajK2gWBqUKoYpMO62d6jqt/b2c7zAfyJwpCUsaJvdQqD1oJORGYcG6pBna3NhUqFx6GGZZuoCpoWFh3Mvx9c",
		"JSESSIONID": "7C3CE6AC396608531EC1FD0309BDC85A",
		#"sessionId": "1d9bfb30-7504-431e-86c0-2a97c109938c",
		"usuarioFormApellido": "Aquitectos",
		"usuarioFormEmail": "ventas.rebora@gmail.com",
		"usuarioFormTelefono": "3313420733",
		"usuarioFormId": "57036554",
		"usuarioPublisher": "true",
		"usuarioIdCompany": "50796870",
		"hideWelcomeBanner": "true",
		"usuarioFormNombre": "Rebora",
		"usuarioLogeado": "ventas.rebora@gmail.com"
	}
    
    headers = [
			{
				"name": "Accept",
				"value": "*/*"
			},
			{
				"name": "Accept-Encoding",
				"value": "gzip, deflate, br"
			},
			{
				"name": "Accept-Language",
				"value": "en-US,en;q=0.5"
			},
			{
				"name": "Connection",
				"value": "keep-alive"
			},
			{
				"name": "content-type",
				"value": "application/json;charset=UTF-8"
			},
			{
				"name": "Cookie",
				"value": "cf_clearance=za9dVMWY0XVeHHhQdFyv60WmEOlY_ogtjaeAa8pt2BA-1698436283-0-1-115e9ea5.ba12bc.465fc145-160.2.1698436283; sessionId=1d9bfb30-7504-431e-86c0-2a97c109938c; g_state={\"i_p\":1699030291139,\"i_l\":3}; __cf_bm=o3ZbLYEzjLlsH6f38XdNJRCjoKibAUSUwGWp9E3ZxCw-1698443451-0-AQdx4izAxcREtaKwakfE3l7AWYswCjZL97xdRwf3uuCoY4MC7ZRlUtw59rlKfTBkEGbYhJcD+1eOf5sLo/dg2gvXYd7nyepMnnNZBzSRAC2v; JSESSIONID=7C3CE6AC396608531EC1FD0309BDC85A; hashKey=4IVc5it99duNGjgDajK2gWBqUKoYpMO62d6jqt/b2c7zAfyJwpCUsaJvdQqD1oJORGYcG6pBna3NhUqFx6GGZZuoCpoWFh3Mvx9c; usuarioLogeado=ventas.rebora@gmail.com; usuarioFormNombre=Rebora; usuarioFormEmail=ventas.rebora%40gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioFormNombre=Rebora; usuarioFormApellido=Aquitectos; usuarioFormEmail=ventas.rebora@gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioPublisher=true; usuarioIdCompany=50796870; hideWelcomeBanner=true"
			},
			{
				"name": "email",
				"value": ""
			},
			{
				"name": "Host",
				"value": "www.inmuebles24.com"
			},
			{
				"name": "Referer",
				"value": "https://www.inmuebles24.com/panel/interesados/182835939"
			},
			{
				"name": "Sec-Fetch-Dest",
				"value": "empty"
			},
			{
				"name": "Sec-Fetch-Mode",
				"value": "cors"
			},
			{
				"name": "Sec-Fetch-Site",
				"value": "same-origin"
			},
			{
				"name": "sessionId",
				"value": "1d9bfb30-7504-431e-86c0-2a97c109938c"
			},
			{
				"name": "TE",
				"value": "trailers"
			},
			{
				"name": "User-Agent",
				"value": "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/111.0"
			},
			{
				"name": "x-panel-portal",
				"value": "24MX"
			}
		]
    obj_headers = {h["name"]: h["value"] for h in headers}

    res = requests.get('http://api.scraperapi.com', params=payload)
    if res.status_code != 200:
        print("Status code is not 200: ", res.status_code)
        print(res.text)
        return None

    print(res.text)
    return res.json()

#get_req("leads-api/publisher/contact/182835939")
webscrappingapi("leads-api/publisher/contact/182835939")