from typing import Any


ALLOWED_PROFILE_FIELDS = {
    "height_cm",
    "weight_kg",
    "age",
    "gender",
    "core_goal",
    "detailed_goal",
    "injury_history",
    "favorite_sports",
    "add_favorite_to_plan",
    "current_exercise_desc",
    "expected_intensity",
    "memories",
}


def _sanitize_user_profile(profile: dict[str, Any]) -> dict[str, Any]:
    if not isinstance(profile, dict):
        return {}
    return {key: profile.get(key) for key in ALLOWED_PROFILE_FIELDS if key in profile}


def build_other_context(payload: dict[str, Any]) -> dict[str, Any]:
    return {
        "user_input": payload.get("user_input", ""),
        "user_profile": _sanitize_user_profile(payload.get("user_profile", {})),
    }
