from zenrows import ZenRowsClient
from dotenv import load_dotenv
import json
import os

from .params import cookies, headers
load_dotenv()

APIKEY = os.getenv('ZENROWS_APIKEY')

client = ZenRowsClient(APIKEY)

params = {
	#"resolve_captcha": "true",
    "js_render": "true",
    "antibot": "true",
    "premium_proxy":"true",
    "proxy_country":"mx",
	#"session_id": 10
}

def get_req(url, logger):
	res = client.get(url, params=params, headers=headers, cookies=cookies)
	if (res.status_code < 200 or res.status_code > 299):
		logger.error(f"no se pudo realizar la request a la url: {url}")
		logger.error(f"STATUS: {res.status_code}")
		print("RESPONSE:", res.text)
		try:
			json.dumps(res.json(), indent=4)
		except Exception as e:
			logger.error(res.text)
		return None
	return res

def post_req(url, data, logger):
	print(data)
	res = client.post(url, data=data, params=params)
	if (res.status_code < 200 or res.status_code > 299):
		logger.error(f"no se pudo realizar la request a la url: {url}")
		logger.error(res.status_code)
		logger.error(res.text)
		try:
			json.dumps(res.json(), indent=4)
		except Exception as e:
			logger.error(res.text)
		return None
	return res

if __name__ == "__main__":
	from params import cookies, headers
	#url = "https://www.inmuebles24.com/leads-api/publisher/leads?offset=0&limit=20&spam=false&status=nondiscarded&sort=last_activity"
	url = "https://www.inmuebles24.com/leads-api/publisher/contact/183569275/user-profile"
	#res = client.get(url, params=params, headers=headers, cookies=cookies)
	res = get_req(url)
	print(res.text)