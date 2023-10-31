import requests
import json
from bs4 import BeautifulSoup

api_calls = 0
succes_api_calls = 0

#Send a POST request with webscrapingapi
def post(url, request, params, log=""):

    api_url = "https://api.webscrapingapi.com/v1"
    params["url"] = url

    global api_calls
    global succes_api_calls

    while True:
        res = requests.post(api_url, params=params, json=request)
        api_calls += 1
        print(f"{log} - [status {res.status_code}] - [time {res.elapsed.total_seconds()}]")
        if res.status_code == 200:
            break
        elif res.status_code == 500:
            return {"publisherOutput": "mailerror"}
            print("Error in the page system")
            print("Probably the sender gmail is detected as spam domain")

    #Get the <pre> </pre> tag
    try:
        soup = BeautifulSoup(res.text, 'html.parser')
        succes_api_calls += 1
        return json.loads(soup.find('pre').text)
    except Exception as e:
        print(e)
        print('CANT LOAD BS4 FROM PAGE')
        print("url:", url)
        print("request:", json.dumps(request, indent=4))
        exit()