import os.path

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

# Define your app's authorization scopes.
# When modifying these scopes, delete the file token.json, if it exists.
SCOPES = ["https://www.googleapis.com/auth/chat.messages.create"]

def main():
    '''
    Authenticates with Chat API via user credentials,
    then creates a text message in a Chat space.
    '''

    # Start with no credentials.
    creds = None
    if os.path.exists('token.json'):
        creds = Credentials.from_authorized_user_file('token.json', SCOPES)

    # Authenticate with Google Workspace
    # and get user authorization.
    flow = InstalledAppFlow.from_client_secrets_file('credentials.json', SCOPES)
    creds = flow.run_local_server()

    # Build a service endpoint for Chat API.
    chat = build('chat', 'v1', credentials=creds)

    # Use the service endpoint to call Chat API.
    result = chat.spaces().messages().create(

        # The space to create the message in.
        #
        # Replace SPACE with a space name.
        # Obtain the space name from the spaces resource of Chat API,
        # or from a space's URL.
        parent='spaces/AAAAaFdSObU',

        # The message to create.
        body={'text': 'Hello, world!'}

    ).execute()

    # Prints details about the created membership.
    print(result)

if __name__ == '__main__':
    main()
