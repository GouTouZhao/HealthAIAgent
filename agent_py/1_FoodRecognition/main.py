import importlib.util
from pathlib import Path
from typing import Literal

from langchain.output_parsers import RetryOutputParser
from langchain_core.output_parsers import JsonOutputParser, PydanticOutputParser
from langchain_core.messages import HumanMessage, SystemMessage
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.runnables import RunnableLambda, RunnablePassthrough, RunnableSequence
from pydantic import BaseModel, RootModel

from common import get_dashscope_llm, log_pipeline_step, read_prompt

MODULE_DIR = Path(__file__).resolve().parent
SYSTEM_PROMPT = read_prompt(MODULE_DIR, "system.txt")
USER_PROMPT = read_prompt(MODULE_DIR, "user.txt")
EXTRACT_USER_PROMPT = read_prompt(MODULE_DIR, "extract_user.txt")

tools_spec = importlib.util.spec_from_file_location(
    "food_tools_module", MODULE_DIR / "tools.py"
)
tools_module = importlib.util.module_from_spec(tools_spec)
tools_spec.loader.exec_module(tools_module)
search_food_from_usda = tools_module.search_food_from_usda

qwen_vl_plus = get_dashscope_llm("qwen-vl-plus", temperature=0)
qwen_max = get_dashscope_llm("qwen-max", temperature=0.2)


def emit_progress(data: dict, step_name: str):
    callback = data.get("_progress_callback")
    if callable(callback):
        safe_data = {k: v for k, v in data.items() if k != "_progress_callback"}
        callback(step_name, safe_data)


class FoodOutput(BaseModel):
    food_name_en: str
    food_name_zh: str
    calories_per_100g: float
    protein_per_100g: float
    fat_per_100g: float
    carbs_per_100g: float


class FoodErrorOutput(BaseModel):
    error: Literal["FOOD_NOT_FOUND"]


class FoodResult(RootModel[FoodOutput | FoodErrorOutput]):
    pass


json_parser = JsonOutputParser()


def summarize_image_url(image_url: str) -> str:
    normalized = str(image_url or "").strip()
    if not normalized:
        return ""
    if normalized.startswith("data:image/"):
        return "[inline_image_data_url]"
    if len(normalized) > 180:
        return normalized[:180] + "..."
    return normalized


def build_step1_messages(data: dict):
    user_input = str(data.get("user_input", "") or "")
    image_url = str(data.get("image_url", "") or "")
    image_url_preview = summarize_image_url(image_url)
    text_prompt = EXTRACT_USER_PROMPT.format(user_input=user_input, image_url=image_url_preview)
    content = [{"type": "text", "text": text_prompt}]
    if image_url:
        content.append({"type": "image_url", "image_url": {"url": image_url}})
    return [
        SystemMessage(content=SYSTEM_PROMPT),
        HumanMessage(content=content),
    ]


def run_step1(data: dict):
    messages = build_step1_messages(data)
    llm_output = qwen_vl_plus.invoke(messages)
    raw_content = llm_output.content
    try:
        step1_output = json_parser.parse(raw_content)
    except Exception:
        fallback_food_name = str(data.get("user_input", "") or "").strip()
        step1_output = {"food_name_en": fallback_food_name, "food_name_zh": fallback_food_name}
    result = {**data, "step1_output": step1_output}
    emit_progress(result, "step1_llm_extract_food")
    log_pipeline_step("1_FoodRecognition", "step1_llm_extract_food", result.get("step1_output", {}))
    return result


step1_runnable = RunnableLambda(run_step1)

step2_tool = RunnablePassthrough.assign(
    food_name_en=RunnableLambda(lambda x: x.get("step1_output", {}).get("food_name_en", "")),
    food_name_zh=RunnableLambda(lambda x: x.get("step1_output", {}).get("food_name_zh", "")),
    usda_data=RunnableLambda(
        lambda x: search_food_from_usda(x.get("step1_output", {}).get("food_name_en", ""))
    ),
)


def run_step2(data: dict):
    result = step2_tool.invoke(data)
    emit_progress(result, "step2_tool_usda")
    log_pipeline_step(
        "1_FoodRecognition",
        "step2_tool_usda",
        {
            "food_name": result.get("food_name", ""),
            "usda_data": result.get("usda_data", {}),
        },
    )
    return result


step2_runnable = RunnableLambda(run_step2)

food_parser = PydanticOutputParser(pydantic_object=FoodResult)
retry_food_parser = RetryOutputParser.from_llm(
    parser=food_parser,
    llm=qwen_max,
    max_retries=2,
)
food_result_prompt = ChatPromptTemplate.from_messages(
    [
        ("system", SYSTEM_PROMPT),
        ("user", USER_PROMPT),
    ]
)


def run_step3(data: dict):
    image_url_preview = summarize_image_url(data.get("image_url", ""))
    prompt_data = {
        "user_input": data.get("user_input", ""),
        "image_url": image_url_preview,
        "food_name_en": data.get("food_name_en", ""),
        "food_name_zh": data.get("food_name_zh", ""),
        "usda_data": data.get("usda_data", {}),
        "format_instructions": food_parser.get_format_instructions(),
    }
    prompt_value = food_result_prompt.invoke(prompt_data)
    raw_output = qwen_max.invoke(prompt_value).content
    result = {**data, "prompt_value": prompt_value, "raw_output": raw_output}
    emit_progress(result, "step3_llm_generate")
    log_pipeline_step("1_FoodRecognition", "step3_llm_generate", {"raw_output": raw_output})
    return result


step3_llm = RunnableLambda(run_step3)


def run_step4(data: dict):
    result = retry_food_parser.parse_with_prompt(
        data.get("raw_output", ""),
        data.get("prompt_value"),
    ).root.model_dump(exclude_none=True)
    emit_progress(result, "step4_validator")
    log_pipeline_step("1_FoodRecognition", "step4_validator", result)
    return result


step4_validator = RunnableLambda(run_step4)

pipeline = RunnableSequence(step1_runnable, step2_runnable, step3_llm, step4_validator)


def run(payload: dict, progress_callback=None):
    runtime_payload = {**payload, "_progress_callback": progress_callback}
    log_pipeline_step("1_FoodRecognition", "pipeline_start", payload)
    result = pipeline.invoke(runtime_payload)
    result.pop("_progress_callback", None)
    log_pipeline_step("1_FoodRecognition", "pipeline_end", result)
    return result
