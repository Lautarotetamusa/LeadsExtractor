# Pruebas realizadas para [Inmuebles24.com](https://www.inmuebles24.com)

- [X] Selenium

  - Funciona pero solamente despues de acceder al navegador manualmente, sino se da cuenta que es un bot y no funciona. No serviría para el propósito de crear workers automaticos.
- [X] Hacer requests a la api desde python.

  - Le pasamos exactamente los mismos headers y cookies que aparecen en el navegador cuando se realiza una petición estando loggeados dentro del sitio.
  - No funciona, devuelve un código binario sin ningun sentido aparente. Si no le pasamos los headers nos dice que no estamos autorizados
- [X] Hacer peticiones através de paginas intermedias.

  - [X] scraperapi.com ❌

    - No pudimos nisiquiera obtener la data de la página principal, aun usando proxy resindenciales, headless browser y ubicacion en mx.
  - [X] webscrapping.com

    - No funcionó
  - [X] scraperant.com ❌

    - [X] Hacer request https://www.inmuebles24.com sin la sesion abierta. ✅

      * Sin utilizar proxy residencial ni headless browser option.
    - [X] Request a la API con sesion iniciada. ❌

      * El Browser fue detectado por Cloudflare aun con proxy residencial en mexico y headless browser.
    - [X] Loggearse con usuario y contraseña, para verificar si al obtener los tokens através de la scraperant. ❌

      * Fue detectado por el antibot de cloudflare aun usando proxy residencial y headless browser.
    - [ ] Brightdata
      El scraping browser funciona muy bien y podemos acceder a la pagina correctamente, sin embargo tenemos problemas a la hora de logearnos.

      - [X] Obtener la pagina principal de inmuebles24. ✅
      - [X] Loggearse en el sitio ❌
        * El browser de brighdata nos tira el siguiente error:
          selenium.common.exceptions.JavascriptException: Message: javascript error: Forbidden action: password typing is not allowed	(Session info: chrome=118.0.5993.88)
          Al parecer brightdata no permite ingresar una contraseña através de su web browser.

    - [] ZenRows
      - [X] Obtener pagina principal. ✅
          - Pudimos obtener la pagina principal utilizando JS rendering y AI antibot.
	  - [X] Obtener panel/interesados ❌
    	  - Obtenemos la pagina de cloudflare 
  	  - [X] Hacer una peticion directamente a la API. ✅
    	  - En principio funciono. Logramos obtener el archivo JSON con la data necesaria.
    	  - Usamos los siguientes parametros
          ```json
		  {
			"js_render": "true",
			"antibot": "true",
			"premium_proxy":"true",
			"proxy_country":"mx"
		  }
          ``` 

- [X] Intentar encontrar la IP original del sitio que está detrás de CloudFlare

  - Usamos diversos metodos para intentar encontrar la ip pero no tuvimos exito.
    - [X] [CloudPeler](https://github.com/zidansec/CloudPeler/tree/master)
    - [X] [CloudFlair](https://github.com/zidansec/CloudPeler)
    - [X] [bypass by DNS history](https://github.com/vincentcox/bypass-firewalls-by-DNS-history)
    - [X] [Censys](https://search.censys.io/search?resource=hosts&sort=RELEVANCE&per_page=25&virtual_hosts=EXCLUDE&q=inmuebles24.com)

# loggearse

`POST f"https://www.inmuebles24.com/rp-api/user/{user}/exist"`

## Request

`POST f"https://www.inmuebles24.com/login_login.ajax"`

```json
{
	"email": user,
	"password": password,
	"recordarme": "true",
	"homeSeeker": "true",
	"urlActual": "https://www.inmuebles24.com"
}
```

## Response

```json
{
	"status": {
		"status": 200
	},
	"errorMessage": "",
	"contenido": {
		"result": true,
		"url": "/panel.bum",
		"validation": {
			"errores": {},
			"erroresDetalle": {}
		},
		"idUsuario": 57036554,
		"sessionID": "1d9bfb30-7504-431e-86c0-2a97c109938c",
		"cookieUser": "ventas.rebora@gmail.com",
		"cookieHash": "4IVc5it99duNAigBci/3yyRpCrk5sO32+tC87cOOnZK6Fqv+kZrN0ZYwa2+p8/1XQH0fHfMWtfmM3EkS2Aubz4pi1EyevGasZrabbtZ4H9mLnm20MElt3m//c5IiFA==",
		"domain": "inmuebles24.com",
		"esAnunciante": true
	},
	"ajaxResponseId": 0
}
```

# Enviar mensaje

`POST f"https://www.inmuebles24.com/leads-api/leads/455447182/messages"`

## Esta supongo que no hace falta

`POST "https://www.inmuebles24.com/leads-api/leads/455447182/message_attachements"`

# Obtener la informacion del lead

## Pagina del lead

`GET "https://www.inmuebles24.com/panel/interesados/182835939"`

## Informacion via API

`GET "https://www.inmuebles24.com/leads-api/publisher/contact/182835939"`

### Response

```json
{
	"id": "455447182",
	"last_lead_date": "2023-10-27T13:13:31.000+00:00",
	"lead_user": {
		"id": "0",
		"user_id": "60982056",
		"name": "Alejandro",
		"email": "alejandroorozcoruiz@hotmail.com",
		"phone": "3313367924"
	},
	"actions": [
		{
			"id": "456002212",
			"date": "2023-10-27T13:13:31.000+00:00",
			"type": {
				"id": "10",
				"description": "Contactó por Whatsapp",
				"name": "WHATSAPP"
			}
		}
	],
	"posting": {
		"id": "90565599",
		"title": "Casa en venta en La Toscana. Hermosa, segura y confortable.",
		"address": "Av. De La Toscana 888",
		"location": {
			"name": "Valle Real",
			"label": "ZONA",
			"parent": {
				"name": "Zapopan",
				"label": "CIUDAD",
				"parent": {
					"name": "Jalisco",
					"label": "PROVINCIA",
					"parent": {
						"name": "Mexico",
						"label": "PAIS"
					}
				}
			}
		},
		"posting_type": "PROPERTY",
		"real_estate_type": {
			"name": "Casa",
			"category": {
				"metadata": {
					"translation": {
						"plural": "Residenciales",
						"singular": "Residencial"
					}
				},
				"real_estate_type_category_id": "1",
				"label": "Residencial"
			},
			"metadata": {
				"translation": {
					"plural": "Casas",
					"singular": "Casa"
				}
			},
			"real_estate_type_id": "1"
		},
		"statuses": [
			"ONLINE"
		],
		"reserved": false,
		"operation_type": "venta",
		"price": {
			"amount": 15700000,
			"currency": "MXN"
		},
		"internal_code": "",
		"pictures": {
			"48x48": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/48x48/1099480892.jpg",
			"126x126": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/126x126/1099480892.jpg",
			"94x94": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/94x94/1099480892.jpg",
			"290x230": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/290x230/1099480892.jpg",
			"360x266": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/360x266/1099480892.jpg",
			"328x125": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/328x125/1099480892.jpg",
			"274x125": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/274x125/1099480892.jpg"
		},
		"price_operation_types": [
			{
				"prices": [
					{
						"amount": 15700000,
						"currency": "MXN"
					}
				],
				"operation_type": {
					"name": "Venta",
					"metadata": {
						"connector": "en",
						"translation": {
							"plural": "Ventas",
							"singular": "venta"
						}
					},
					"operation_type_id": "1"
				}
			}
		],
		"image_url": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/290x230/1099480892.jpg",
		"small_image_url": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/48x48/1099480892.jpg",
		"detail_image_url": "https://img10.naventcdn.com/avisos/18/00/90/56/55/99/328x125/1099480892.jpg"
	},
	"last_message": {
		"text": "Hola qué tal, Gracias por comunicarte a Rebora Arquitectos, me presento Brenda Diaz ¿Cómo puedo ayudarte? Enseguida me pongo en contacto contigo!",
		"date": "2023-10-27T20:13:42+0000",
		"from": "57036554",
		"to": "60982056",
		"from_seeker": true,
		"message_id": "423629260",
		"is_comment": false
	},
	"contact_publisher_user_id": "182835939",
	"statuses": [
		"READ"
	],
	"phone": "3313367924",
	"phone_list": [
		"3313367924"
	]
}
```

# Informacion acerca de lo que busca el usuario

`GET "https://www.inmuebles24.com/leads-api/publisher/contact/182835939/user-profile"`

## Response

```json
{
	"site_id": "24MX",
	"user_id": 60982056,
	"lead_info": {
		"search_type": {
			"operation": "Venta",
			"type": "Casa",
			"operation_type_id": "1",
			"realestate_type_id": "1"
		},
		"price": {
			"currency": "MXN",
			"max": 19850000,
			"min": 15700000
		},
		"views": 8,
		"contacts": 3,
		"started_search_days": 0
	},
	"searched_locations": {
		"streets": [
			{
				"city": "Jalisco",
				"name": "avenida  de la toscana",
				"percent": 33
			},
			{
				"city": "Jalisco",
				"name": "coto almendros",
				"percent": 33
			},
			{
				"city": "Jalisco",
				"name": "pº san arturo ote",
				"percent": 33
			}
		],
		"neighborhoods": [
			{
				"amount": 100,
				"max_amount": 85,
				"name": "Valle Real"
			}
		],
		"polygons": {
			"type": "multipolygon",
			"coordinates": [
				[
					[
						[
							-103.4320717,
							20.7272783
						],
						[
							-103.4321678,
							20.7270333
						],
						[
							-103.4323287,
							20.7267882
						],
						[
							-103.4325934,
							20.7265432
						],
						[
							-103.4333682,
							20.7262982
						],
						[
							-103.4470512,
							20.7272864
						],
						[
							-103.447826,
							20.7275314
						],
						[
							-103.4480907,
							20.7277764
						],
						[
							-103.4482516,
							20.7280215
						],
						[
							-103.4483477,
							20.7282665
						],
						[
							-103.4483933,
							20.7285115
						],
						[
							-103.4483933,
							20.7287565
						],
						[
							-103.4483477,
							20.7290015
						],
						[
							-103.4482516,
							20.7292465
						],
						[
							-103.4480907,
							20.7294916
						],
						[
							-103.447826,
							20.7297366
						],
						[
							-103.4470512,
							20.7299816
						],
						[
							-103.4399691,
							20.729689
						],
						[
							-103.4333682,
							20.7289934
						],
						[
							-103.4325934,
							20.7287484
						],
						[
							-103.4323287,
							20.7285034
						],
						[
							-103.4321678,
							20.7282583
						],
						[
							-103.4320717,
							20.7280133
						],
						[
							-103.4320261,
							20.7277683
						],
						[
							-103.4320261,
							20.7275233
						],
						[
							-103.4320717,
							20.7272783
						]
					]
				]
			]
		}
	},
	"property_features": {
		"bedrooms": {
			"max": 5,
			"min": 3
		},
		"covered_area_xm2": {
			"max": 450,
			"min": 409
		},
		"total_area_xm2": {
			"max": 342,
			"min": 290
		},
		"baths": {
			"max": 5,
			"min": 4
		},
		"age": {
			"max": 5,
			"min": 0
		},
		"pct_of_postings_with_expenses": 0,
		"pct_of_postings_with_garage": 100,
		"loan_capable": 0,
		"professional_capable": 0,
		"offers_financing": 0,
		"pet_friendly": 0,
		"commercial_use": 0
	},
	"user_questionary_profiles": []
}
```

# Obtener toda la data necesesaria

* Primero nos loggeamos y obtenemos el session_id haciendo
  * `POST rp-api/user/{user}/exist`
  * `POST f"https://www.inmuebles24.com/login_login.ajax"`
* Obtenemos la data de la persona y de la publicacion a la que consulto
  * `GET "https://www.inmuebles24.com/leads-api/publisher/contact/182835939"`
  * la respuesta la guardaremos en el objecto `data`
* Obtenemos la data de las busquedas de la persona
  * `GET "https://www.inmuebles24.com/leads-api/publisher/contact/182835939/user-profile"`
  * en este caso la respuesta la guardaremos en el object `busqueda`
* Armamos el objecto final de la siguiente manera:

```python

data = requests.get()
busqueda = requests.get()

lead_info = {
	"nombre": data["lead_user"]["name"],
	"link": "https://www.inmuebles24.com/panel/interesados/182835939",
	"telefono_1": data["phone_list"][0],
	"telefono_2": data["phone_list"][1] or "",
	"email": data["lead_user"]["email"],
	"propiedad":{
		"nombre": data["posting"]["title"],
		"link":
		"precio": data["posting"]["price"]["amount"],
		"ubicacion": data["posting"]["address"],
		"tipo": data["posting"]["real_estate_type"]["name"],
	},
	"busquedas":{
		"zonas": [for i["name"] in busqueda["searched_locations"]["streets"]]
		"tipo": busqueda["lead_info"]["search_type"]["type"],
		"total_area": busqueda["property_features"]["total_area_xm2"]["min"] + " " + busqueda["property_features"]["total_area_xm2"]["max"],
		"covered_area": busqueda["property_features"]["covered_area_xm2"]["min"] + " " + busqueda["property_features"]["covered_area_xm2"]["max"],
		"banios": busqueda["property_features"]["baths"]["min"] + " " + busqueda["property_features"]["baths"]["max"],
		"recamaras": busqueda["property_features"]["bedrooms"]["min"] + " " + busqueda["property_features"]["bedrooms"]["max"],
		"presupuesto": busqueda["lead_info"]["price"]["min"] + " " + busqueda["lead_info"]["price"]["max"],
		"cantidad_anuncios": busqueda["lead_info"]["views"]
		"contactos": busqueda["lead_info"]["contacts"],
		"inicio_busqueda": busqueda["lead_info"]["started_search_days"] //Hace cuanto tiempo comenzo a buscar
	}
}
```

* Enviamos un mensaje al lead
  * `POST f"https://www.inmuebles24.com/leads-api/leads/455447182/messages"`
