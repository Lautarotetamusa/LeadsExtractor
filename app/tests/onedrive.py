import sys
sys.path.append('.')
from src.onedrive.main import OneDrive

if __name__ == "__main__":
    od = OneDrive()
    images = [
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCEREDHBDIRUER655CJO6IWNRCWUHJC/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCERECIJVTA2HVZYRCLLVA7UU2X7YBF/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCEREHY5Q7JYUIJWFD2M3DI3IJWZNFA/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCERECXXO4SDRSEU5E3FT7OPNXE2TWP/content",

        # The same urls to test the cache
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCEREDHBDIRUER655CJO6IWNRCWUHJC/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCERECIJVTA2HVZYRCLLVA7UU2X7YBF/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCEREHY5Q7JYUIJWFD2M3DI3IJWZNFA/content",
        "https://graph.microsoft.com/v1.0/drives/b!zdgDJiBvHkW3w7RM8tHO3op6N0Vfd0pPq7JLkfwb0V1lb38lwGanTb49NytJD2PW/items/01VOCERECXXO4SDRSEU5E3FT7OPNXE2TWP/content",
    ]

    for image in images:
        content = od.download_file(image)
        if content is None:
            print("error downloading the image")
            continue

        with open("image.png", "wb") as f:
            f.write(content)
