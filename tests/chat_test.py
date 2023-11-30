from httplib2 import Http
from oauth2client.service_account import ServiceAccountCredentials
#from apiclient.discovery import build
from googleapiclient.discovery import build

# Specify required scopes.
SCOPES = ['https://www.googleapis.com/auth/chat.bot']

# Specify service account details.
CREDENTIALS = ServiceAccountCredentials.from_json_keyfile_name('credentials.json', SCOPES)

# Build the URI and authenticate with the service account.
chat = build('chat', 'v1', http=CREDENTIALS.authorize(Http()))

# Use the service endpoint to call Chat API.
result = chat.spaces().list(
    # An optional filter that returns named spaces or unnamed
    # group chats, but not direct messages (DMs).
    filter='SPACE'

).execute()

print(result)