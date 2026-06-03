import os
from typing import Any

import requests

USDA_SEARCH_API = "https://api.nal.usda.gov/fdc/v1/foods/search"


def _nutrient_value(
    nutrients: list[dict[str, Any]],
    name: str,
    unit: str = "",
) -> float | None:
    nutrient_map = dict(
        map(
            lambda item: (
                (item.get("nutrientName", ""), str(item.get("unitName", "")).upper()),
                item.get("value"),
            ),
            nutrients,
        )
    )
    value = nutrient_map.get((name, unit.upper()))
    return ({True: lambda x: float(x), False: lambda _: None}[value is not None])(value)


def _clean_candidate(food_item: dict[str, Any]) -> dict[str, Any]:
    nutrients = food_item.get("foodNutrients", [])
    return {
        "description": food_item.get("description", ""),
        "fdc_id": food_item.get("fdcId"),
        "food_category": food_item.get("foodCategory", ""),
        "calories_per_100g": _nutrient_value(nutrients, "Energy", "KCAL"),
        "protein_per_100g": _nutrient_value(nutrients, "Protein", "G"),
        "fat_per_100g": _nutrient_value(nutrients, "Total lipid (fat)", "G"),
        "carbs_per_100g": _nutrient_value(
            nutrients,
            "Carbohydrate, by difference",
            "G",
        ),
    }


def search_food_from_usda(food_name: str) -> dict[str, Any]:
    response = requests.get(
        USDA_SEARCH_API,
        params={
            "query": food_name,
            "pageSize": 5,
            "dataType": ["Foundation", "SR Legacy"],
            "api_key": os.getenv("USDA_API_KEY", "DEMO_KEY"),
        },
        timeout=20,
    )
    data = response.json()
    candidates = data.get("foods", [])
    cleaned_candidates = list(map(_clean_candidate, candidates[:5]))
    return {
        "food_name": food_name,
        "candidate_count": len(cleaned_candidates),
        "candidates": cleaned_candidates,
    }
