from dataclasses import dataclass
from datetime import datetime
from typing import Optional
from enum import Enum

class PropertyType(Enum):
    HOUSE = "house"
    APARTMENT = "apartment"

class OperationType(Enum):
    SALE = "sale"
    RENT = "rent"

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

@dataclass
class Property:
    id: int
    title: str
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
    images: list[dict[str, str]]
    parking_lots: Optional[int] = None
    bathrooms: Optional[int] = None
    half_bathrooms: Optional[int] = None
    video_url: Optional[str] = None
    virtual_route: Optional[str] = None
    neighborhood: Optional[str] = None
