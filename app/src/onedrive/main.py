import requests

from src.onedrive.onedrive_authorization_utils import load_token, procure_new_tokens, refresh_access_token, DRIVE_ID

# TODO: dont do this global i know, but i have to finish today.
token = load_token() 
if token is None: # If the token file does not exists
    token = procure_new_tokens()

def download_url_from_id(item_id: str) -> str:
    return f"https://graph.microsoft.com/v1.0/drives/{DRIVE_ID}/items/{item_id}/content"

def download_file(token: dict, download_url: str) -> bytes | None:
    headers = {
        "Authorization": f"Bearer {token['access_token']}",
        "Content-Type": "application/json"
    }

    if token["expires_in"] <= 0:
        print("refreshing OneDrive code..")
        new_token = refresh_access_token(token)
        if new_token is None:
            print("cannot refresh the token")
            return None

    res = requests.get(download_url, headers=headers, stream=True)

    if res.ok:
        return res.content

    return None

if __name__ == "__main__":
    token = load_token()
    if token is None: # If the token file does not exists
        token = procure_new_tokens()

    # At this point the token its valid not matter whats
    item_id = "01VOCEREB3UO5WLXDB6ZHYVWW5BHH2SPSI"
    content = download_file(token, download_url_from_id(item_id))
    if content is None:
        exit(1)
    with open("image.png", "wb") as f:
        f.write(content)
