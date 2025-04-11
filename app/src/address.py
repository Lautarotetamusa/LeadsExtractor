import re

def extract_street_from_address(address: str) -> str:
    parts = [part.strip() for part in address.split(',')]
    
    # Find the first part that contains a 5-digit postal code
    postal_code_index = None
    for i, part in enumerate(parts):
        if re.search(r'\b\d{5}\b', part):
            postal_code_index = i
            break
    
    # If no postal code found, return the entire address
    if postal_code_index is None:
        return ', '.join(parts)
    
    # Extract parts before the postal code
    street_parts = parts[:postal_code_index]
    
    # Find the last part in street_parts that contains a number (street number)
    last_number_index = None
    for i, part in enumerate(street_parts):
        if re.search(r'\d', part):
            last_number_index = i
    
    # Slice street_parts up to and including the last part with a number
    if last_number_index is not None:
        street_parts = street_parts[:last_number_index + 1]
    
    return ', '.join(street_parts)
