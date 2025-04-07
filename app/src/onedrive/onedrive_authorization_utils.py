import json
import os
import time
import msal
import webbrowser
import urllib.parse
import requests
from dotenv import load_dotenv

load_dotenv()
APP_ID = os.environ.get("MS_OPENGRAPH_APP_ID")
CLIENT_SECRET = os.environ.get("MS_OPENGRAPH_CLIENT_SECRET")
DRIVE_ID = os.environ.get("DRIVE_ID")
SCOPES = ["Files.ReadWrite.All"]
AUTHORITY_URL = "https://login.microsoftonline.com/common"
TOKEN_FILE = "./src/onedrive/token.json"

def load_token() -> dict[str, str] | None:
    try:
        with open(TOKEN_FILE, "r") as f:
            return json.load(f)
    except FileNotFoundError:
        return None

def save_token(token: dict):
    # Save the expires_at time
    token["expires_at"] = time.time() + token["expires_in"] 
    with open(TOKEN_FILE, "w") as f:
        json.dump(token, f)

def extract_code_from_url(url: str) -> str | None:
    parsed_url = urllib.parse.urlparse(url)
    query_params = urllib.parse.parse_qs(parsed_url.query)
    
    code = query_params.get("code", [None])[0]
    if code is None:
        print("code dont found on the redirect_url")
        return None
    return code

def procure_new_tokens() -> dict:
    client_instance = msal.ConfidentialClientApplication(
        APP_ID, 
        client_credential=CLIENT_SECRET, 
        authority=AUTHORITY_URL
    )
    auth_url = client_instance.get_authorization_request_url(SCOPES)
    webbrowser.open(auth_url)
    
    redirected_url = input("Enter the url from your browser: ").strip()
    code = extract_code_from_url(redirected_url)
    if code is None: exit(1)

    token = client_instance.acquire_token_by_authorization_code(code, SCOPES)
    
    if "access_token" in token:
        save_token(token)
        print("Success authentication")
        return token
    else:
        print("Authentication error", token.get("error_description"))
        exit(1)

def refresh_access_token(token) -> dict | None:
    """ get the new access token with the refresh token """

    if not "refresh_token" in token:
        print("no refresh token, authenticate manually again")
        return procure_new_tokens()

    url = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
    data = {
        "client_id": APP_ID,
        "client_secret": CLIENT_SECRET,
        "scope": "https://graph.microsoft.com/Files.ReadWrite.All",
        "grant_type": "refresh_token",
        "refresh_token": token["refresh_token"],
    }
    
    res = requests.post(url, data=data)
    if not res.ok: 
        print("Error while refreshing the token", res.json())
        return None
    token = res.json()
    
    if "access_token" in token:
        save_token(token)
        return token
    else:
        print("Authentication error", token.get("error_description"))
        return None
