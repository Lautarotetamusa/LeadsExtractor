import __init__

from zenrows import ZenRowsClient
from params import cookies, headers
import json
import logger
import os

from dotenv import load_dotenv
load_dotenv()

APIKEY = os.getenv('ZENROWS_APIKEY')

client = ZenRowsClient(APIKEY)

params = {
    #"js_render": "true",
	"resolve_captcha": "true",
    #"antibot": "true",
    "premium_proxy":"true",
    "proxy_country":"mx",
	"session_id": 1
}

def get_req(url):
	res = client.get(url, params=params, headers=headers, cookies=cookies)
	if (res.status_code < 200 or res.status_code > 299):
		logger.log_print("no se pudo realizar la request a la url: "+url, logger.LogType.error)
		logger.log_print("STATUS: "+str(res.status_code), logger.LogType.error)
		print("RESPONSE:", res.text)
		try:
			json.dumps(res.json(), indent=4)
		except Exception as e:
			logger.log_print(res.text, logger.LogType.error)
		return None
	return res

def post_req(url, data):
	print(data)
	res = client.post(url, data=data, params=params)
	if (res.status_code < 200 or res.status_code > 299):
		logger.log_print("no se pudo realizar la request POST a la url: "+url, logger.LogType.error)
		logger.log_print("STATUS: "+str(res.status_code), logger.LogType.error)
		print("RESPONSE:", res.text)
		try:
			json.dumps(res.json(), indent=4)
		except Exception as e:
			logger.log_print(res.text, logger.LogType.error)
		return None
	return res

if __name__ == "__main__":
	from params import cookies, headers
	#url = "https://www.inmuebles24.com/leads-api/publisher/leads?offset=0&limit=20&spam=false&status=nondiscarded&sort=last_activity"
	url = "https://www.inmuebles24.com/leads-api/publisher/contact/183569275/user-profile"
	#res = client.get(url, params=params, headers=headers, cookies=cookies)
	res = get_req(url)
	print(res.text)