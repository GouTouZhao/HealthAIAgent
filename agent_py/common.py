import os
import json
import logging
from pathlib import Path

from langchain_openai import ChatOpenAI

BASE_DIR = Path(__file__).resolve().parent
DASHSCOPE_BASE_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

logging.basicConfig(level=logging.DEBUG, format="%(asctime)s [%(levelname)s] %(message)s")
LOGGER = logging.getLogger("agent_py")


def read_prompt(module_dir: Path, prompt_name: str) -> str:
    prompt_path = module_dir / "prompts" / prompt_name
    return prompt_path.read_text(encoding="utf-8")


def get_dashscope_llm(model_name: str, temperature: float = 0.2) -> ChatOpenAI:
    return ChatOpenAI(
        model=model_name,
        api_key=os.getenv("DASHSCOPE_API_KEY", ""),
        base_url=os.getenv("DASHSCOPE_BASE_URL", DASHSCOPE_BASE_URL),
        temperature=temperature,
        request_timeout=180,
    )


def preview_data(data: object, limit: int = 700) -> str:
    try:
        text = json.dumps(data, ensure_ascii=False, default=str)
    except Exception:
        text = str(data)
    if len(text) > limit:
        return text[:limit] + "..."
    return text


def log_pipeline_step(mode_name: str, step_name: str, data: object):
    LOGGER.debug("[%s] %s => %s", mode_name, step_name, preview_data(data))
