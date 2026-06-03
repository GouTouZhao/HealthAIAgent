import requests
import json
import os
from pathlib import Path
from dotenv import load_dotenv

# Load .env from project root
load_dotenv(Path(__file__).resolve().parent / ".env")

def test_usda_api():
    api_key = os.getenv("USDA_API_KEY", "St8kAmMrb2CTL0tl2dtbRJJedqySwirVELEGYBm7")
    query = "egg"
    # Added dataType restriction for "Foundation" and "SR Legacy"
    url = f"https://api.nal.usda.gov/fdc/v1/foods/search?api_key={api_key}&query={query}&dataType=Foundation&dataType=SR%20Legacy"

    print(f"Requesting USDA API for '{query}' with dataTypes [Foundation, SR Legacy]...")
    
    try:
        response = requests.get(url)
        response.raise_for_status()  # Raise an error for bad status codes
        
        # Output raw JSON response
        print("\nRaw Response Content:")
        print(json.dumps(response.json(), indent=2, ensure_ascii=False))
        
    except requests.exceptions.RequestException as e:
        print(f"An error occurred: {e}")
        if hasattr(e.response, 'text'):
            print(f"Response text: {e.response.text}")

if __name__ == "__main__":
    test_usda_api()


# Requesting USDA API for 'egg' with dataTypes [Foundation, SR Legacy]...

# Raw Response Content:
# {
#   "totalHits": 398,
#   "currentPage": 1,
#   "totalPages": 8,
#   "pageList": [
#     1,
#     2,
#     3,
#     4,
#     5,
#     6,
#     7,
#     8
#   ],
#   "foodSearchCriteria": {
#     "dataType": [
#       "Foundation",
#       "SR Legacy"
#     ],
#     "query": "egg",
#     "generalSearchInput": "egg",
#     "pageNumber": 1,
#     "numberOfResultsPerPage": 50,
#     "pageSize": 50,
#     "requireAllWords": false,
#     "foodTypes": [
#       "Foundation",
#       "SR Legacy"
#     ]
#   },
#   "foods": [
#     {
#       "fdcId": 747997,
#       "description": "Eggs, Grade A, Large, egg white",
#       "commonNames": "",
#       "additionalDescriptions": "",
#       "dataType": "Foundation",
#       "ndbNumber": 1124,
#       "publishedDate": "2019-12-16",
#       "foodCategory": "Dairy and Egg Products",
#       "mostRecentAcquisitionDate": "2019-07-08",
#       "allHighlightFields": "",
#       "score": 337.9996,
#       "microbes": [],
#       "foodNutrients": [
#         {
#           "nutrientId": 1166,
#           "nutrientName": "Riboflavin",
#           "nutrientNumber": "405",
#           "unitName": "MG",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.391,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 6500,
#           "indentLevel": 1,
#           "foodNutrientId": 8526264,
#           "dataPoints": 24,
#           "min": 0.302,
#           "max": 0.504,
#           "median": 0.376
#         },
#         {
#           "nutrientId": 1003,
#           "nutrientName": "Protein",
#           "nutrientNumber": "203",
#           "unitName": "G",
#           "derivationCode": "NC",
#           "derivationDescription": "Calculated",
#           "derivationId": 49,
#           "value": 10.7,
#           "foodNutrientSourceId": 2,
#           "foodNutrientSourceCode": "4",
#           "foodNutrientSourceDescription": "Calculated or imputed",
#           "rank": 600,
#           "indentLevel": 1,
#           "foodNutrientId": 8526265,
#           "min": 9.5,
#           "max": 12.3,
#           "median": 10.6
#         },
#         {
#           "nutrientId": 1007,
#           "nutrientName": "Ash",
#           "nutrientNumber": "207",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.65,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 1000,
#           "indentLevel": 1,
#           "foodNutrientId": 8526266,
#           "dataPoints": 24,
#           "min": 0.48,
#           "max": 0.79,
#           "median": 0.66
#         },
#         {
#           "nutrientId": 1103,
#           "nutrientName": "Selenium, Se",
#           "nutrientNumber": "317",
#           "unitName": "UG",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 17.9,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 6200,
#           "indentLevel": 1,
#           "foodNutrientId": 8526267,
#           "dataPoints": 24,
#           "min": 7.0,
#           "max": 48.8,
#           "median": 14.6
#         },
#         {
#           "nutrientId": 1004,
#           "nutrientName": "Total lipid (fat)",
#           "nutrientNumber": "204",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.0,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 800,
#           "indentLevel": 1,
#           "foodNutrientId": 8526268,
#           "dataPoints": 24,
#           "min": 0.0,
#           "max": 0.0,
#           "median": 0.0
#         },
#         {
#           "nutrientId": 1051,
#           "nutrientName": "Water",
#           "nutrientNumber": "255",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 86.3,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 100,
#           "indentLevel": 1,
#           "foodNutrientId": 8526269,
#           "dataPoints": 24,
#           "min": 84.4,
#           "max": 88.4,
#           "median": 86.3
#         },
#         {
#           "nutrientId": 1062,
#           "nutrientName": "Energy",
#           "nutrientNumber": "268",
#           "unitName": "kJ",
#           "derivationCode": "NC",
#           "derivationDescription": "Calculated",
#           "derivationId": 49,
#           "value": 231,
#           "foodNutrientSourceId": 2,
#           "foodNutrientSourceCode": "4",
#           "foodNutrientSourceDescription": "Calculated or imputed",
#           "rank": 400,
#           "indentLevel": 1,
#           "foodNutrientId": 8526270
#         },
#         {
#           "nutrientId": 1008,
#           "nutrientName": "Energy",
#           "nutrientNumber": "208",
#           "unitName": "KCAL",
#           "derivationCode": "NC",
#           "derivationDescription": "Calculated",
#           "derivationId": 49,
#           "value": 55.0,
#           "foodNutrientSourceId": 2,
#           "foodNutrientSourceCode": "4",
#           "foodNutrientSourceDescription": "Calculated or imputed",
#           "rank": 300,
#           "indentLevel": 1,
#           "foodNutrientId": 8526271
#         },
#         {
#           "nutrientId": 1005,
#           "nutrientName": "Carbohydrate, by difference",
#           "nutrientNumber": "205",
#           "unitName": "G",
#           "derivationCode": "NC",
#           "derivationDescription": "Calculated",
#           "derivationId": 49,
#           "value": 2.36,
#           "foodNutrientSourceId": 2,
#           "foodNutrientSourceCode": "4",
#           "foodNutrientSourceDescription": "Calculated or imputed",
#           "rank": 1110,
#           "indentLevel": 2,
#           "foodNutrientId": 8526272
#         },
#         {
#           "nutrientId": 1002,
#           "nutrientName": "Nitrogen",
#           "nutrientNumber": "202",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 1.71,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 500,
#           "indentLevel": 1,
#           "foodNutrientId": 8526273,
#           "dataPoints": 24,
#           "min": 1.52,
#           "max": 1.97,
#           "median": 1.7
#         },
#         {
#           "nutrientId": 1210,
#           "nutrientName": "Tryptophan",
#           "nutrientNumber": "501",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.188,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 16300,
#           "indentLevel": 1,
#           "foodNutrientId": 8526274,
#           "dataPoints": 18,
#           "min": 0.15,
#           "max": 0.23,
#           "median": 0.19
#         },
#         {
#           "nutrientId": 1218,
#           "nutrientName": "Tyrosine",
#           "nutrientNumber": "509",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.466,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from analytical",      
#           "rank": 17100,
#           "indentLevel": 1,
#           "foodNutrientId": 8526275,
#           "dataPoints": 18,
#           "min": 0.34,
#           "max": 0.54,
#           "median": 0.485
#         },
#         {
#           "nutrientId": 1219,
#           "nutrientName": "Valine",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.779,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.779,
#           "foodNutrientSourceId": 1,
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.779,
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "nutrientNumber": "510",
#           "unitName": "G",
#           "derivationCode": "A",
#           "derivationDescription": "Analytical",
#           "derivationId": 1,
#           "value": 0.779,
#           "foodNutrientSourceId": 1,
#           "foodNutrientSourceCode": "1",
#           "foodNutrientSourceDescription": "Analytical or derived from a