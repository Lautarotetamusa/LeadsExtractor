from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common import exceptions

def get_by_selenium():
    options = Options()
    options.add_argument(f"--user-data-dir=session") #Session

    driver = webdriver.Chrome(options=options)

    driver.get("https://www.inmuebles24.com/panel/interesados/")

    WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[3]/table")))
    
    links = driver.find_elements(By.XPATH, '//td[@class="sc-ifAKCX sc-cTtDJl kdACLV postingColumn"]')

    print(links)


get_by_selenium()


##ENVIAR MENSAJE CON SELENIUM
"""
def send_message(driver, contact_id, msg):
	url = f"https://www.inmuebles24.com/panel/interesados/{contact_id}"
	
	logger.debug(f"Enviando mensaje al lead {contact_id}")
	try:
		driver.get(url)

		WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/textarea")))

		input = driver.find_element(By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/textarea")
		input.send_keys(msg)

		button = driver.find_element(By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/button")
		button.click()
		logger.success("Mensaje enviado correctamente")
	except Exception as e:
		logger.error("Ocurrio un error enviando el mensaje")
		logger.error(str(e))
"""