# pip install zenrows
from zenrows import ZenRowsClient

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
				"value": "cf_clearance=7yoBJbwat9LAoVao3s_nXQWA9csObG5gtgsDF4um24A-1699154087-0-1-115e9ea5.ba12bc.465fc145-160.2.1699154087; sessionId=08233694-c999-41ee-b6ea-91fd4f9c27e7; g_state={\"i_p\":1699030291139,\"i_l\":3}; hashKey=4IVc5it99duNGjgDajK2gWBqUKoGvsO609S2qt3Z2c31AfyJxpCUsaJtcAqD1oJORGYcG6pBna3NhUrE7PW0tnDlfTVpJS2R0hY9; usuarioLogeado=ventas.rebora@gmail.com; usuarioFormNombre=Rebora; usuarioFormEmail=ventas.rebora%40gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioFormNombre=Rebora; usuarioFormApellido=Aquitectos; usuarioFormEmail=ventas.rebora@gmail.com; usuarioFormTelefono=3313420733; usuarioFormId=57036554; usuarioPublisher=true; usuarioIdCompany=50796870; hideWelcomeBanner=true; _gcl_au=1.1.805164046.1698609866; _ga_8XFRKTEF9J=GS1.1.1699143157.4.1.1699143794.48.0.0; _ga=GA1.2.1232673381.1698609867; __rtbh.lid=%7B%22eventType%22%3A%22lid%22%2C%22id%22%3A%22mT2RKa8aWnPnBkTSoKky%22%7D; _ga_F842TPK3EE=GS1.2.1699143159.4.1.1699143751.32.0.0; _hjSessionUser_174024=eyJpZCI6IjhkMDg1YTY3LWM5NTAtNWY2Zi1hMDM3LTg2OTViY2YzZTgyYSIsImNyZWF0ZWQiOjE2OTg2MDk4Njg0ODksImV4aXN0aW5nIjp0cnVlfQ==; _fbp=fb.1.1698609869726.1255128104; mousestats_vi=fc98ac1a434c881e4202; __gads=ID=675995fc19c9ffb7:T=1698775996:RT=1699153951:S=ALNI_MZ-BntMcU8J1M_QDn1lvyYKlRrFzg; __gpi=UID=00000a3be27e374b:T=1698775996:RT=1699153951:S=ALNI_Ma4N5plSE60bXXTWOi2C-fxi8-C-w; __rtbh.uid=%7B%22eventType%22%3A%22uid%22%7D; __cf_bm=iKud1bCOTilUNj5w0.PrfxXbTB4zMQAgt0Q7MWc0uTo-1699153950-0-AeLoayc8me3ax6oqOMM97ZCpI3ezQOarp/FGdLl8TUKlYnp16+1vVShoQSIyfSuHZxqUcgPsnkZmxTKJMG0h8XEkB85faGSbD2lwaL0KOf0j; JSESSIONID=E5786A5A25978B9DEC9BB6B97D8EF894; _gid=GA1.2.1506021246.1699143160; _hjIncludedInSessionSample_174024=0; _hjSession_174024=eyJpZCI6IjA3N2I3OTk4LWM0YmMtNGYyZi1hOGZmLTIyMzdmMDMxYmQ1MiIsImNyZWF0ZWQiOjE2OTkxNDMxNjA3MjQsImluU2FtcGxlIjpmYWxzZSwic2Vzc2lvbml6ZXJCZXRhRW5hYmxlZCI6ZmFsc2V9; _hjAbsoluteSessionInProgress=0; _hjHasCachedUserAttributes=true; mousestats_si=61be15dbba4d29f8fe3e; reputationTooltipSeen=true; _dc_gtm_UA-10030475-1=1"
			},
			{
				"name": "email",
				"value": ""
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
				"value": "08233694-c999-41ee-b6ea-91fd4f9c27e7"
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
cookies = {
		"__cf_bm": "iKud1bCOTilUNj5w0.PrfxXbTB4zMQAgt0Q7MWc0uTo-1699153950-0-AeLoayc8me3ax6oqOMM97ZCpI3ezQOarp/FGdLl8TUKlYnp16+1vVShoQSIyfSuHZxqUcgPsnkZmxTKJMG0h8XEkB85faGSbD2lwaL0KOf0j",
		"__gads": "ID=675995fc19c9ffb7:T=1698775996:RT=1699153951:S=ALNI_MZ-BntMcU8J1M_QDn1lvyYKlRrFzg",
		"__gpi": "UID=00000a3be27e374b:T=1698775996:RT=1699153951:S=ALNI_Ma4N5plSE60bXXTWOi2C-fxi8-C-w",
		"__rtbh.lid": "{\"eventType\":\"lid\",\"id\":\"mT2RKa8aWnPnBkTSoKky\"}",
		"__rtbh.uid": "{\"eventType\":\"uid\"}",
		"_fbp": "fb.1.1698609869726.1255128104",
		"_ga": "GA1.2.1232673381.1698609867",
		"_ga_8XFRKTEF9J": "GS1.1.1699143157.4.1.1699143281.58.0.0",
		"_ga_F842TPK3EE": "GS1.2.1699143159.4.0.1699143159.60.0.0",
		"_gcl_au": "1.1.805164046.1698609866",
		"_gid": "GA1.2.1506021246.1699143160",
		"_hjAbsoluteSessionInProgress": "0",
		"_hjHasCachedUserAttributes": "true",
		"_hjIncludedInSessionSample_174024": "0",
		"_hjSession_174024": "eyJpZCI6IjA3N2I3OTk4LWM0YmMtNGYyZi1hOGZmLTIyMzdmMDMxYmQ1MiIsImNyZWF0ZWQiOjE2OTkxNDMxNjA3MjQsImluU2FtcGxlIjpmYWxzZSwic2Vzc2lvbml6ZXJCZXRhRW5hYmxlZCI6ZmFsc2V9",
		"_hjSessionUser_174024": "eyJpZCI6IjhkMDg1YTY3LWM5NTAtNWY2Zi1hMDM3LTg2OTViY2YzZTgyYSIsImNyZWF0ZWQiOjE2OTg2MDk4Njg0ODksImV4aXN0aW5nIjp0cnVlfQ==",
		"cf_clearance": "7yoBJbwat9LAoVao3s_nXQWA9csObG5gtgsDF4um24A-1699154087-0-1-115e9ea5.ba12bc.465fc145-160.2.1699154087",
		"g_state": "{\"i_p\":1699030291139,\"i_l\":3}",
		"hashKey": "4IVc5it99duNGjgDajK2gWBqUKoGvsO609S2qt3Z2c31AfyJxpCUsaJtcAqD1oJORGYcG6pBna3NhUrE7PW0tnDlfTVpJS2R0hY9",
		"hideWelcomeBanner": "true",
		"JSESSIONID": "E5786A5A25978B9DEC9BB6B97D8EF894",
		"mousestats_si": "61be15dbba4d29f8fe3e",
		"mousestats_vi": "fc98ac1a434c881e4202",
		"sessionId": "08233694-c999-41ee-b6ea-91fd4f9c27e7",
		"usuarioFormApellido": "Aquitectos",
		"usuarioFormEmail": "ventas.rebora@gmail.com",
		"usuarioFormId": "57036554",
		"usuarioFormNombre": "Rebora",
		"usuarioFormTelefono": "3313420733",
		"usuarioIdCompany": "50796870",
		"usuarioLogeado": "ventas.rebora@gmail.com",
		"usuarioPublisher": "true"
	}


client = ZenRowsClient("ce23db25f7f551a5043c05c6375c5b251cf494c6")
url = "https://www.inmuebles24.com/leads-api/publisher/leads?offset=0&limit=20&spam=false&status=nondiscarded&sort=last_activity"

params = {
    "js_render": "true",
    "antibot": "true",
    "premium_proxy":"true",
    "proxy_country":"mx"
}

res = client.get(url, params=params, headers=obj_headers, cookies=cookies)

print(res.text)