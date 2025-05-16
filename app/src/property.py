from dataclasses import dataclass
from datetime import datetime
from typing import Optional, Type, TypedDict
from enum import Enum

class PropertyType(Enum):
    HOUSE = "house"
    APARTMENT = "apartment"

class OperationType(Enum):
    SALE = "sell"
    RENT = "rent"

class PlanType(Enum):
    SIMPLE = "simple"
    HIGHLIGHTED = "highlighted"
    SUPER = "super"

@dataclass
class Image:
    url: str

@dataclass
class Location:
    lat: float
    lng: float

@dataclass
class Ubication:
    address: str
    location: Location

class Internal(TypedDict):
    state: str
    state_id: int
    city: str
    city_id: int
    colony: str
    colony_id: int
    zipcode: str | None

@dataclass
class Property:
    id: int
    title: str
    zone: str
    price: str  # Stored as string to preserve formatting (e.g., "300000")
    currency: str
    description: str
    type: PropertyType
    antiquity: int
    rooms: int
    operation_type: OperationType
    m2_total: int
    m2_covered: int
    ubication: Ubication # Full GOOGLE address
    updated_at: datetime
    created_at: datetime
    plan: PlanType 

    internal: Internal

    images: list[dict[str, str]]
    parking_lots: Optional[int] = None
    bathrooms: Optional[int] = None
    half_bathrooms: Optional[int] = None
    video_url: Optional[str] = None
    virtual_route: Optional[str] = None
