from selenium.webdriver import Remote, ChromeOptions
from selenium.webdriver.chromium.remote_connection import ChromiumRemoteConnection
from selenium.webdriver.common.by import By

EMAIL = "ventas.rebora@gmail.com"
PASS = "i242023raf"

AUTH = 'brd-customer-hl_f0e8b04a-zone-inmuebles24:7brv9tiqsfw5'
SBR_WEBDRIVER = f'https://{AUTH}@brd.superproxy.io:9515'

options = ChromeOptions()
options.add_argument(f"--user-data-dir=session") #Session

def main():
    print('Connecting to Scraping Browser...')
    sbr_connection = ChromiumRemoteConnection(SBR_WEBDRIVER, 'goog', 'chrome')
    with Remote(sbr_connection, options=options) as driver:
        print('Connected! Navigating...')

        driver.get('https://inmuebles24.com')

        print('Getting logged')
        login_btn = driver.find_element(By.XPATH, '//button[@class="StyledButton-sc-1b3blmr-0 kigrxY"]')
        login_btn.click()

        print('Submiting email')
        input_email = driver.find_element(By.XPATH, '//input[@id="input-email"]')
        input_email.send_keys(EMAIL)

        driver.get_screenshot_as_file('1.png')

        continuar_btn = driver.find_element(By.XPATH, '//button[@data-qa="SUBMIT_EMAIL"]')
        continuar_btn.click()

        driver.get_screenshot_as_file('2.png')

        print('Submiting password')

        js = """
            document.getElementById('input-password').type = "text";
        """
        driver.execute_script(js)
        print("js executed")
        input_pass = driver.find_element(By.XPATH, '//input[@id="input-password"]')
        print("input pass selected")
        input_pass.send_keys(PASS)
        print("password sended")

        driver.get_screenshot_as_file('3.png')

        continuar_btn = driver.find_element(By.XPATH, '//button[@data-qa="PASSWORD_SUBMIT"]')
        continuar_btn.click()

        print('logged succesfully')

        driver.get_screenshot_as_file('4.png')

        driver.get("https://www.inmuebles24.com/panel/interesados")

        driver.get_screenshot_as_file('5.png')

        html = driver.page_source
        print(html)

if __name__ == '__main__':
    main()