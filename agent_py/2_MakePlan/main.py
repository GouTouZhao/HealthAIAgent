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
    "make_plan_tools_module", MODULE_DIR / "tools.py"
)
tools_module = importlib.util.module_from_spec(tools_spec)
tools_spec.loader.exec_module(tools_module)
build_plan_context = tools_module.build_plan_context

qwen_max = get_dashscope_llm("qwen-max", temperature=0.2)


def emit_progress(data: dict, step_name: str):
    callback = data.get("_progress_callback")
    if callable(callback):
        safe_data = {k: v for k, v in data.items() if k != "_progress_callback"}
        callback(step_name, safe_data)


AEROBIC_KEYWORDS = (
    "跑",
    "慢跑",
    "快跑",
    "跑步",
    "冲刺",
    "骑行",
    "单车",
    "动感单车",
    "游泳",
    "椭圆机",
    "划船机",
    "跳绳",
    "有氧",
    "run",
    "jog",
    "cardio",
    "bike",
    "cycling",
    "swim",
)

RECOVERY_KEYWORDS = (
    "拉伸",
    "瑜伽",
    "普拉提",
    "恢复",
    "放松",
    "mobility",
    "stretch",
    "recovery",
)

STRENGTH_KEYWORDS = (
    "卧推",
    "深蹲",
    "硬拉",
    "推举",
    "弯举",
    "飞鸟",
    "引体",
    "下拉",
    "划船",
    "臀桥",
    "箭步蹲",
    "腿举",
    "腿弯举",
    "平板支撑",
    "卷腹",
    "俯卧撑",
    "器械",
    "杠铃",
    "哑铃",
    "bench",
    "squat",
    "deadlift",
    "press",
    "curl",
    "row",
    "pull",
    "push",
    "plank",
)

MUSCLE_GROUP_KEYWORDS = {
    "chest": ("卧推", "飞鸟", "俯卧撑", "bench", "chest"),
    "back": ("划船", "引体", "下拉", "硬拉", "row", "pull", "lat", "deadlift"),
    "legs": ("深蹲", "腿举", "腿弯举", "箭步", "提踵", "硬拉", "squat", "lunge", "leg", "deadlift"),
    "shoulders": ("推举", "侧平举", "前平举", "肩", "press", "shoulder"),
    "arms": ("弯举", "臂屈伸", "三头", "二头", "curl", "triceps", "biceps"),
    "core": (
        "平板",
        "卷腹",
        "核心",
        "俄罗斯转体",
        "死虫",
        "死虫式",
        "抗旋转",
        "pallof",
        "dead bug",
        "plank",
        "core",
        "crunch",
    ),
}


def _normalize_text(text: str) -> str:
    return str(text or "").strip().lower()


def _contains_keyword(text: str, keywords: tuple[str, ...]) -> bool:
    return any(word in text for word in keywords)


def _is_rest_activity(name: str) -> bool:
    normalized = _normalize_text(name)
    return normalized in {"rest", "休息", "恢复"}


def _is_aerobic_activity(name: str) -> bool:
    normalized = _normalize_text(name)
    return _contains_keyword(normalized, AEROBIC_KEYWORDS)


def _is_recovery_activity(name: str) -> bool:
    normalized = _normalize_text(name)
    return _contains_keyword(normalized, RECOVERY_KEYWORDS)


def _is_strength_activity(name: str) -> bool:
    normalized = _normalize_text(name)
    if not normalized or _is_rest_activity(normalized) or _is_recovery_activity(normalized):
        return False
    if _contains_keyword(normalized, STRENGTH_KEYWORDS):
        return True
    return not _is_aerobic_activity(normalized)


def _infer_muscle_groups(name: str) -> set[str]:
    normalized = _normalize_text(name)
    groups: set[str] = set()
    for group, keywords in MUSCLE_GROUP_KEYWORDS.items():
        if _contains_keyword(normalized, keywords):
            groups.add(group)
    return groups


class Activity(BaseModel):
    activity: str
    sets_reps: str | None = None


class DayPlan(BaseModel):
    activities: list[Activity] = Field(min_length=1)
    day_intensity: str = Field(min_length=1)
    estimated_duration: str = Field(min_length=1)
    rest_interval: str | None = None

    @model_validator(mode="after")
    def validate_strength_day_rules(self):
        non_rest_activities = [item for item in self.activities if not _is_rest_activity(item.activity)]
        if not non_rest_activities:
            return self

        has_aerobic = any(_is_aerobic_activity(item.activity) for item in non_rest_activities)
        strength_activities = [item for item in non_rest_activities if _is_strength_activity(item.activity)]
        if strength_activities and not has_aerobic:
            if len(strength_activities) < 4:
                raise ValueError("力量训练日至少需要4个力量动作")

            covered_groups: set[str] = set()
            for item in strength_activities:
                covered_groups.update(_infer_muscle_groups(item.activity))
            if covered_groups and len(covered_groups) < 2:
                raise ValueError("力量训练日动作需要覆盖至少2个不同肌群")
        return self


class WeeklyPlan(BaseModel):
    monday: DayPlan
    tuesday: DayPlan
    wednesday: DayPlan
    thursday: DayPlan
    friday: DayPlan
    saturday: DayPlan
    sunday: DayPlan


class MonthPlan(BaseModel):
    month: int
    goal: str = Field(min_length=1)
    diet_advice: str = Field(min_length=1)
    weekly_plan: WeeklyPlan


class PlanOutput(BaseModel):
    duration: str = Field(min_length=1)
    total_goal: str = Field(min_length=1)
    months: list[MonthPlan]

    @model_validator(mode="after")
    def validate_month_goals_not_fixed(self):
        month_goals = [item.goal.strip().lower() for item in self.months if item.goal.strip()]
        if len(month_goals) > 1 and len(set(month_goals)) == 1:
            raise ValueError("每个月的goal不能固定重复")
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
    log_pipeline_step("2_MakePlan", "step1_llm_normalize", result.get("step1_output", {}))
    return result


step1_runnable = RunnableLambda(run_step1)

step2_tool = RunnablePassthrough.assign(
    context=RunnableLambda(lambda x: build_plan_context(x.get("step1_output", {})))
)


def run_step2(data: dict[str, Any]):
    result = step2_tool.invoke(data)
    emit_progress(result, "step2_tool_context")
    log_pipeline_step("2_MakePlan", "step2_tool_context", result.get("context", {}))
    return result


step2_runnable = RunnableLambda(run_step2)

plan_parser = PydanticOutputParser(pydantic_object=PlanOutput)
retry_plan_parser = RetryOutputParser.from_llm(
    parser=plan_parser,
    llm=qwen_max,
    max_retries=2,
)


def run_step3(data: dict[str, Any]):
    context = data.get("context", {})
    prompt_value = build_prompt.invoke(
        {
            "user_input": context.get("user_input", ""),
            "user_profile": context.get("user_profile", {}),
            "format_instructions": plan_parser.get_format_instructions(),
        }
    )
    raw_output = qwen_max.invoke(prompt_value).content
    result = {**data, "prompt_value": prompt_value, "raw_output": raw_output}
    emit_progress(result, "step3_llm_generate")
    log_pipeline_step("2_MakePlan", "step3_llm_generate", {"raw_output": raw_output})
    return result


step3_llm = RunnableLambda(run_step3)
def run_step4(data: dict[str, Any]):
    result = retry_plan_parser.parse_with_prompt(
        data.get("raw_output", ""), data.get("prompt_value")
    ).model_dump(exclude_none=True)
    emit_progress(result, "step4_validator")
    log_pipeline_step("2_MakePlan", "step4_validator", result)
    return result


step4_validator = RunnableLambda(run_step4)

pipeline = RunnableSequence(step1_runnable, step2_runnable, step3_llm, step4_validator)


def run(payload: dict, progress_callback=None):
    runtime_payload = {**payload, "_progress_callback": progress_callback}
    log_pipeline_step("2_MakePlan", "pipeline_start", payload)
    result = pipeline.invoke(runtime_payload)
    result.pop("_progress_callback", None)
    log_pipeline_step("2_MakePlan", "pipeline_end", result)
    return result
