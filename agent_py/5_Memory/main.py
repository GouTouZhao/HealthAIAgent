import importlib.util
from pathlib import Path
from typing import Any

from langchain.output_parsers import RetryOutputParser
from langchain_core.output_parsers import JsonOutputParser, PydanticOutputParser
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.runnables import RunnableLambda, RunnablePassthrough, RunnableSequence
from pydantic import BaseModel, Field, model_validator

from common import get_dashscope_llm, log_pipeline_step, read_prompt

MODULE_DIR = Path(__file__).resolve().parent
SYSTEM_PROMPT = read_prompt(MODULE_DIR, "system.txt")
USER_PROMPT = read_prompt(MODULE_DIR, "user.txt")
NORMALIZE_USER_PROMPT = read_prompt(MODULE_DIR, "normalize_user.txt")

tools_spec = importlib.util.spec_from_file_location(
    "memory_tools_module", MODULE_DIR / "tools.py"
)
tools_module = importlib.util.module_from_spec(tools_spec)
tools_spec.loader.exec_module(tools_module)
build_memory_context = tools_module.build_memory_context

qwen_max = get_dashscope_llm("qwen-max", temperature=0.2)


def emit_progress(data: dict, step_name: str):
    callback = data.get("_progress_callback")
    if callable(callback):
        safe_data = {k: v for k, v in data.items() if k != "_progress_callback"}
        callback(step_name, safe_data)


class MemoryItem(BaseModel):
    memory: str = Field(min_length=1, max_length=120)
    tag: str = Field(min_length=1, max_length=30)

    @model_validator(mode="after")
    def normalize_fields(self):
        self.memory = self.memory.strip()
        self.tag = self.tag.strip().lower()
        return self


class MemoryOutput(BaseModel):
    summary: str = Field(default="")
    memories: list[MemoryItem] = Field(default_factory=list)

    @model_validator(mode="after")
    def dedupe_memories(self):
        seen: set[str] = set()
        deduped: list[MemoryItem] = []
        for item in self.memories:
            key = f"{item.tag}:{item.memory.strip().lower()}"
            if key in seen:
                continue
            seen.add(key)
            deduped.append(item)
        self.memories = deduped
        return self


build_prompt = ChatPromptTemplate.from_messages(
    [
        ("system", SYSTEM_PROMPT),
        ("user", USER_PROMPT),
    ]
)

normalize_prompt = ChatPromptTemplate.from_messages(
    [
        ("system", SYSTEM_PROMPT),
        ("user", NORMALIZE_USER_PROMPT),
    ]
)

step1_llm = RunnablePassthrough.assign(
    step1_output=normalize_prompt | qwen_max | JsonOutputParser()
)


def run_step1(data: dict[str, Any]):
    result = step1_llm.invoke(data)
    emit_progress(result, "step1_llm_normalize")
    log_pipeline_step("5_Memory", "step1_llm_normalize", result.get("step1_output", {}))
    return result


step1_runnable = RunnableLambda(run_step1)

step2_tool = RunnablePassthrough.assign(
    context=RunnableLambda(lambda x: build_memory_context(x))
)


def run_step2(data: dict[str, Any]):
    result = step2_tool.invoke(data)
    emit_progress(result, "step2_tool_context")
    log_pipeline_step("5_Memory", "step2_tool_context", result.get("context", {}))
    return result


step2_runnable = RunnableLambda(run_step2)

memory_parser = PydanticOutputParser(pydantic_object=MemoryOutput)
retry_memory_parser = RetryOutputParser.from_llm(
    parser=memory_parser,
    llm=qwen_max,
    max_retries=2,
)


def run_step3(data: dict[str, Any]):
    context = data.get("context", {})
    prompt_value = build_prompt.invoke(
        {
            "user_input": context.get("user_input", ""),
            "existing_memories": context.get("existing_memories", []),
            "user_profile": context.get("user_profile", {}),
            "format_instructions": memory_parser.get_format_instructions(),
        }
    )
    raw_output = qwen_max.invoke(prompt_value).content
    result = {**data, "prompt_value": prompt_value, "raw_output": raw_output}
    emit_progress(result, "step3_llm_generate")
    log_pipeline_step("5_Memory", "step3_llm_generate", {"raw_output": raw_output})
    return result


step3_llm = RunnableLambda(run_step3)


def run_step4(data: dict[str, Any]):
    result = retry_memory_parser.parse_with_prompt(
        data.get("raw_output", ""), data.get("prompt_value")
    ).model_dump(exclude_none=True)
    emit_progress(result, "step4_validator")
    log_pipeline_step("5_Memory", "step4_validator", result)
    return result


step4_validator = RunnableLambda(run_step4)

pipeline = RunnableSequence(step1_runnable, step2_runnable, step3_llm, step4_validator)


def run(payload: dict, progress_callback=None):
    runtime_payload = {**payload, "_progress_callback": progress_callback}
    log_pipeline_step("5_Memory", "pipeline_start", payload)
    result = pipeline.invoke(runtime_payload)
    result.pop("_progress_callback", None)
    log_pipeline_step("5_Memory", "pipeline_end", result)
    return result
