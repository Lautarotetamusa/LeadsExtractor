import requests
from time import time

from src.onedrive.onedrive_authorization_utils import load_token, procure_new_tokens, refresh_access_token, DRIVE_ID

token = load_token() 
if token is None: # If the token file does not exists
    token = procure_new_tokens()

class OneDrive():
    def __init__(self, cache_max_size=20):
        self.token: dict = {}

        self.token = load_token() 
        if self.token is None: # If the token file does not exists
            self.token = procure_new_tokens()
        self.refresh_token()

        # Each link has the content of the image cached
        # dict[url, image content]
        self.cache: dict[str, bytes] = { }

        self.cache_max_size = cache_max_size

    def download_file(self, download_url: str) -> bytes | None:
        print(f"downloading {download_url}")
        self.refresh_token()

        if download_url in self.cache:
            print("image its cached")
            return self.cache[download_url]

        headers = {
            "Authorization": f"Bearer {self.token['access_token']}",
            "Content-Type": "application/json"
        }
        res = requests.get(download_url, headers=headers, stream=True)

        if res.ok:
            self.cache[download_url] = res.content
            if len(self.cache) > self.cache_max_size:
                # This its not probably the first key, but delete some key works fine
                first_key = list(self.cache.keys())[0]
                self.cache.pop(first_key)
            return res.content

    def refresh_token(self):
        if time() >= self.token["expires_at"]:
            print("refreshing OneDrive code..")
            new_token = refresh_access_token(self.token)
            if new_token is None:
                print("cannot refresh the token")
                return None
            self.token = new_token

def download_file(token, download_url: str) -> bytes | None:
    if time() >= token["expires_at"]:
        print("refreshing OneDrive code..")
        new_token = refresh_access_token(token)
        if new_token is None:
            print("cannot refresh the token")
            return None
        token = new_token

    headers = {
        "Authorization": f"Bearer {token['access_token']}",
        "Content-Type": "application/json"
    }
    res = requests.get(download_url, headers=headers, stream=True)

    if res.ok:
        return res.content

def download_url_from_id(item_id: str) -> str:
    return f"https://graph.microsoft.com/v1.0/drives/{DRIVE_ID}/items/{item_id}/content"
