from typing import Any


def _extract_existing_memories(user_profile: dict[str, Any]) -> list[dict[str, Any]]:
    memories = user_profile.get("memories", []) if isinstance(user_profile, dict) else []
    if not isinstance(memories, list):
        return []

    out: list[dict[str, Any]] = []
    for item in memories:
        if not isinstance(item, dict):
            continue
        memory_text = str(item.get("memory", "")).strip()
        tag_text = str(item.get("tag", "")).strip().lower()
        if not memory_text or not tag_text:
            continue
        out.append({"memory": memory_text, "tag": tag_text})
    return out


def build_memory_context(payload: dict[str, Any]) -> dict[str, Any]:
    step1_output = payload.get("step1_output", {})
    if not isinstance(step1_output, dict):
        step1_output = {}

    user_profile = payload.get("user_profile", {})
    if not isinstance(user_profile, dict):
        user_profile = {}

    normalized_user_input = str(step1_output.get("user_input", "")).strip()
    if not normalized_user_input:
        normalized_user_input = str(payload.get("user_input", "")).strip()

    return {
        "user_input": normalized_user_input,
        "user_profile": user_profile,
        "existing_memories": _extract_existing_memories(user_profile),
    }
