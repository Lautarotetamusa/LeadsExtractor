import requests

url = 'https://www.inmuebles24.com/leads-api/leads/464754951/messages'
apikey = 'f569290d8d48d2d19a288ff2a95cad9b6679ab4d'
params = {
    'url': url,
    'apikey': apikey,
	'premium_proxy': 'true',
	'proxy_country': 'mx',
	'custom_headers': 'true',
	'autoparse': 'true',
}
headers = {
	'sessionId': 'e8264398-9f24-4f19-833b-1ed87b5657b1',
	'idUsuario': '57036554',
	'x-panel-portal': '24MX',
}
data = {
	'is_comment': False,
	'message': 'test',
	'message_attachment':  []
}

response = requests.post('https://api.zenrows.com/v1/', params=params, headers=headers, json=data)
print(response.text)
