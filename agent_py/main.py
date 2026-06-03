import importlib.util
from pathlib import Path
import sys
import os
import json
from queue import Queue
from threading import Thread
from types import ModuleType
from typing import Any

from fastapi import FastAPI
from pydantic import BaseModel
from starlette.responses import StreamingResponse
from dotenv import load_dotenv

from common import log_pipeline_step

BASE_DIR = Path(__file__).resolve().parent
# Load .env from project root or current dir
if (BASE_DIR.parent / ".env").exists():
    load_dotenv(BASE_DIR.parent / ".env")
    print(f"Loaded .env from {BASE_DIR.parent / '.env'}")
else:
    load_dotenv()
    print("Attempted to load .env from current directory")

sys.path.insert(0, str(BASE_DIR))


def _load_module(module_name: str, file_path: Path) -> ModuleType:
    spec = importlib.util.spec_from_file_location(module_name, file_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


food_module = _load_module(
    "food_pipeline",
    BASE_DIR / "1_FoodRecognition" / "main.py",
)
make_plan_module = _load_module(
    "make_plan_pipeline",
    BASE_DIR / "2_MakePlan" / "main.py",
)
change_plan_module = _load_module(
    "change_plan_pipeline",
    BASE_DIR / "3_ChangePlan" / "main.py",
)
other_module = _load_module(
    "other_pipeline",
    BASE_DIR / "4_OtherQuestion" / "main.py",
)
memory_module = _load_module(
    "memory_pipeline",
    BASE_DIR / "5_Memory" / "main.py",
)

PIPELINE_RUNNERS = {
    "1_FoodRecognition": food_module.run,
    "2_MakePlan": make_plan_module.run,
    "3_ChangePlan": change_plan_module.run,
    "4_OtherQuestion": other_module.run,
    "5_Memory": memory_module.run,
}

app = FastAPI(title="AI Agent Pipeline Service")


class PipelineRequest(BaseModel):
    mode: str = "4_OtherQuestion"
    user_input: str = ""
    image_url: str | None = None
    user_profile: dict[str, Any] | None = None
    existing_plan: dict[str, Any] | None = None


def _to_json_safe(value: Any):
    if value is None or isinstance(value, (str, int, float, bool)):
        return value
    if isinstance(value, dict):
        return {str(k): _to_json_safe(v) for k, v in value.items()}
    if isinstance(value, (list, tuple, set)):
        return [_to_json_safe(v) for v in value]
    return str(value)


def _build_gateway_payload(request: PipelineRequest) -> dict[str, Any]:
    payload = request.model_dump(exclude_none=True)
    if request.mode != "3_ChangePlan":
        return payload

    existing_plan = payload.get("existing_plan")
    if isinstance(existing_plan, dict) and existing_plan:
        return payload

    fallback_plan: dict[str, Any] = {}
    user_profile = payload.get("user_profile")
    if isinstance(user_profile, dict):
        raw_latest_plan = user_profile.get("latest_plan_json")
        if isinstance(raw_latest_plan, str):
            trimmed_latest_plan = raw_latest_plan.strip()
            if trimmed_latest_plan:
                try:
                    parsed_latest_plan = json.loads(trimmed_latest_plan)
                    if isinstance(parsed_latest_plan, dict):
                        fallback_plan = parsed_latest_plan
                except json.JSONDecodeError:
                    pass

    payload["existing_plan"] = fallback_plan
    return payload


@app.post("/pipeline/run")
def run_pipeline(request: PipelineRequest):
    payload = _build_gateway_payload(request)
    runner = PIPELINE_RUNNERS.get(request.mode, other_module.run)
    log_pipeline_step("gateway", "pipeline_request", {"mode": request.mode, "payload": payload})
    result = runner(payload)
    log_pipeline_step("gateway", "pipeline_response", {"mode": request.mode, "result": result})
    return result


@app.post("/pipeline/run_stream")
def run_pipeline_stream(request: PipelineRequest):
    payload = _build_gateway_payload(request)
    runner = PIPELINE_RUNNERS.get(request.mode, other_module.run)
    log_pipeline_step("gateway", "pipeline_request", {"mode": request.mode, "payload": payload})

    event_queue: Queue[dict[str, Any]] = Queue()

    def on_progress(step_name: str, step_data: Any = None):
        event_queue.put(
            {
                "type": "step",
                "step": step_name,
                "data": _to_json_safe(step_data),
            }
        )

    def worker():
        try:
            result = runner(payload, progress_callback=on_progress)
            event_queue.put({"type": "result", "data": result})
            log_pipeline_step("gateway", "pipeline_response", {"mode": request.mode, "result": result})
        except Exception as exc:
            log_pipeline_step("gateway", "pipeline_error", {"mode": request.mode, "error": str(exc)})
            event_queue.put({"type": "error", "error": str(exc)})
        finally:
            event_queue.put({"type": "done"})

    Thread(target=worker, daemon=True).start()

    def event_generator():
        while True:
            event = event_queue.get()
            yield json.dumps(event, ensure_ascii=False) + "\n"
            if event.get("type") == "done":
                break

    return StreamingResponse(event_generator(), media_type="application/x-ndjson")

