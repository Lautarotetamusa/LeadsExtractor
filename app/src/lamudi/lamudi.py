from datetime import date, datetime
import json
import os
from typing import Iterator
import uuid

import requests

from src.api import download_file
from src.scraper import SENDER_PHONE
from src.property import Property
from src.portal import Mode, Portal
from src.lead import Lead
from src.lamudi.amenities import amenities, nearby_locations

API_URL = "https://api.proppit.com"
DATE_FORMAT = os.getenv("DATE_FORMAT")
CONTACT_EMAIL = os.getenv("SENDER_EMAIL")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"


class Lamudi(Portal):
    def __init__(self):
        super().__init__(
            name="lamudi",
            contact_id_field="id",
            send_msg_field="id",
            username_env="LAMUDI_USERNAME",
            password_env="LAMUDI_PASSWORD",
            params_type="cookies",
            unauthorized_codes=[401],
            filename=__file__
        )

    def login(self):
        self.logger.debug("Iniciando sesion")
        login_url = f"{API_URL}/login"

        data = {
            "email": self.username,
            "password": self.password
        }
        res = requests.post(login_url, json=data)
        if res is None:
            return

        self.request.cookies = {
            "authToken": res.cookies["authToken"]
        }
        with open(self.params_file, "w") as f:
            json.dump(self.request.cookies, f, indent=4)
        self.logger.success("Sesion iniciada con exito")

    def get_leads(self, mode=Mode.NEW) -> Iterator[list[dict]]:
        status = "new lead"
        page = 1
        limit = 25
        total = 0
        cant = -1

        self.logger.debug("Extrayendo leads")

        while cant != 0:
            if mode == Mode.NEW:
                url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
            else:
                url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity"

            res = self.request.make(url, 'GET')
            if res is None:
                break

            data = res.json()["data"]
            cant = len(data["rows"])
            total += cant
            page += 1
            self.logger.debug(f"len: {cant}")

            yield data["rows"]

        self.logger.success(f"Se encontraron {total} nuevos Leads")

    def get_lead_property(self, lead_id):
        property_url = f"{API_URL}/leads/{lead_id}/properties"
        res = self.request.make(property_url)
        if res is None:
            return {}
        props = res.json().get("data", [])

        formatted_props = []
        for p in props:
            format_prop = {
                "titulo": p.get("title", ""),
                "id": p.get("id", ""),
                "link": f"https://www.lamudi.com.mx/detalle/{p['id']}",
                "precio": str(p.get("price", {}).get("amount", "")),
                "ubicacion": p.get("address", ""),  # Direccion completa
                "tipo": p.get("propertyType", ""),
                # Solamente el municipio, lo usamos para generar el mensaje
                "municipio": p["geoLevels"][0]["name"] if len(p["geoLevels"]) > 0 else "",
                "bedrooms": str(p.get("bedrooms", "")),
                "bathrooms": str(p.get("bathrooms", "")),
                "covered_area": str(p.get("floorArea", "")),
                "total_area": ""
            }

            plot_area = p.get("plotArea", [])
            if len(plot_area) > 0:
                format_prop["total_area"] = str(plot_area[0].get("value", ""))

            formatted_props.append(format_prop)

        return formatted_props

    def get_lead_info(self, raw_lead: dict) -> Lead:
        lead = Lead()
        lead.set_args({
            "lead_id": raw_lead["id"],
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["lastActivity"], "%Y-%m-%dT%H:%M:%SZ").strftime(DATE_FORMAT),
            "nombre": raw_lead.get("name"),
            "link": f"https://proppit.com/leads/{raw_lead['id']}",
            "email": raw_lead.get("email"),
            "telefono": raw_lead.get("phone")
        })

        lead.set_propiedad(self.get_lead_property(raw_lead["id"])[0])
        return lead

    def send_message(self, id, message):
        self.logger.debug(f"Enviando mensaje a lead {id}")
        # Tenemos que pasarle un id del mensaje, lo generamos nosotros con uuid()
        msg_id = str(uuid.uuid4())
        msg_url = f"{API_URL}/leads/{id}/notes/{msg_id}"

        data = {
            "id": msg_id,
            "message": message
        }
        self.request.make(msg_url, 'POST', json=data)
        self.logger.success(f"Mensaje enviado correctamente a lead {id}")

    def make_contacted(self, lead):
        id = lead[self.contact_id_field]
        self.logger.debug(f"Marcando como contactacto a lead {id}")
        read_url = f"{API_URL}/leads/{id}/status"

        data = {
            "status": "contacted"
        }
        self.request.make(read_url, 'PUT', json=data)
        self.logger.success(f"Se contacto correctamente a lead {id}")

    def get_location(self, address) -> dict | None: 
        prediction_url = f"{API_URL}/address-suggestions?query={address}"
        address_url = f"{API_URL}/property-geolocations/address?address={address}"

        res = self.request.make(prediction_url, "GET")
        if res is None or not res.ok:
            return None

        predictions = res.json().get("predictions", [])
        if len(predictions) == 0: 
            self.logger.error("cannot get address predictions: " + str(res.json()))
            return None
        place_id = predictions[0].get("place_id")
        if place_id is None: return None

        res = self.request.make(address_url + f"&place_id={place_id}", "GET")
        if res is None or not res.ok:
            return None

        # {
        #     geoLevels
        #     latitude
        #     longitude
        #     address
        # }
        return res.json().get("data")

    def unpublish(self, publication_id: str) -> Exception | None:
        url = f"{API_URL}/properties/{publication_id}"

        res = self.request.make(url, "DELETE")
        if res is None or not res.ok:
            return Exception(f"cannot delete the property with id: {publication_id}")

    def publish(self, property: Property):
        id = str(uuid.uuid4())

        self.logger.debug("getting the address geo location")
        location_data = self.get_location(property.ubication.address)
        if location_data is None: 
            return Exception("cannot get the location data"), None
        self.logger.success("geolocation data getted successfully")

        ad_payload = {
            "address": property.ubication.address,
            "coordinates": {
                "latitude": property.ubication.location.lat,
                "longitude": property.ubication.location.lng
            },
            ## This field its required
            "geoLevels": location_data["geoLevels"],
            "locationVisibility": "accurate",
            "floorPlans": [],
            "propertyImages": [],
            "mainImageIndex": 0,
            "titleMultiLanguage": [{
                "text": property.title,
                "locale": "es-MX"
            }],
            "descriptionMultiLanguage": [{
                "text": property.description,
                "locale": "es-MX"
            }],
            "id": id,
            "bathrooms": property.bathrooms,
            "bedrooms": property.rooms,
            "toilets": property.half_bathrooms,
            "parkingSpaces": property.parking_lots,
            "bankProperty": False,
            "condition": "new",
            "furnished": "unfurnished",
            "constructionYear": date.today().year - (property.antiquity if property.antiquity is not None else 0),
            "floorArea": property.m2_covered, # Built area
            "floorAreaUnit": "sqm",
            "usableArea": property.m2_covered, # Usable = Built area
            "usableAreaUnit": "sqm",
            "plotArea": [ # Terrain
                {
                    "value": property.m2_total,
                    "unit": "m2"
                }
            ],
            "amenities": amenities,
            "nearbyLocations": nearby_locations,
            "contactEmails": [CONTACT_EMAIL],
            "contactWhatsApp": SENDER_PHONE,
            "contactPhone": SENDER_PHONE,
            "propertyType": property.type.__str__(),
            "operations": [{
                "type": property.operation_type.__str__(),
                "price": {
                    "amount": property.price,
                    "currency": property.currency
                }
            }],
            "propertyImagesSortedWithAi": False
        }

        if property.video_url is not None:
            ad_payload["video"] = property.video_url
        if property.virtual_route is not None:
            ad_payload["virtualTours"] = [property.virtual_route]

        i = 0
        # Construct the propertyImages ad_payload field
        for image in property.images:
            ad_payload["propertyImages"].append({
                "ref": i,
                "isProjectImage": False
            })
            i += 1
        json_ad_payload = json.dumps(ad_payload)
        files: list[tuple] = [
            ("ad", (None, json_ad_payload))
        ]

        # Insert the images files for the request
        i = 0
        for image in property.images:
            self.logger.debug(f"downloading image {image['url']}")
            img_data = download_file(image["url"])
            if img_data is None:
                return Exception("cannot download the image"), None
            img_type = "png" if "png" in image["url"] else "jpeg"

            files.append(
                # field name, (name, data, img_type)
                (f"propertyImagesToBeUploaded[{i}]", (f"image.{img_type}", img_data, f"image/{img_type}"))
            )
            i += 1

        self.logger.debug("creating property")
        res = self.request.make(f"{API_URL}/properties/{id}", "POST", files=files)
        if res is None or not res.ok:
            return Exception("cannot create the property"), None
        self.logger.success("property created successfully")

        self.logger.debug("publishing property")
        publish_url = f"{API_URL}/properties/{id}/published"
        res = self.request.make(publish_url, "PUT", json={
            "published": True
        })
        if res is None or not res.ok:
            return Exception("cannot publish the property"), None
        self.logger.success("property published successfully")

        return None, id
