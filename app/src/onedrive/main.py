import requests

from onedrive_authorization_utils import load_token, procure_new_tokens, refresh_access_token 

def download_file_from_id(token: dict, drive_id: str, item_id: str) -> bytes | None:
    headers = {
        "Authorization": f"Bearer {token['access_token']}",
        "Content-Type": "application/json"
    }

    download_url = f"https://graph.microsoft.com/v1.0/drives/{drive_id}/items/{item_id}/content"
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
    drive_id = "b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW"
    item_id = "01VOCEREB3UO5WLXDB6ZHYVWW5BHH2SPSI"
    content = download_file_from_id(token, drive_id, item_id)
    if content is None:
        exit(1)
    with open("image.png", "wb") as f:
        f.write(content)
