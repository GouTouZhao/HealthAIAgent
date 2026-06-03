from typing import Any


def build_change_context(payload: dict[str, Any]) -> dict[str, Any]:
    return {
        "user_input": payload.get("user_input", ""),
        "user_profile": payload.get("user_profile", {}),
        "existing_plan": payload.get("existing_plan", {}),
        "forbidden_activities": payload.get("forbidden_activities", []),
    }
