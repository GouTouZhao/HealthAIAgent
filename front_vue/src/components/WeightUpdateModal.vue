<template>
  <div class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-2 md:p-4">
    <div class="w-full max-w-5xl max-h-[95vh] overflow-hidden rounded-3xl border border-border bg-card shadow-2xl flex flex-col">
      <div class="px-4 md:px-6 py-4 border-b border-border flex items-center justify-between">
        <div>
          <h3 class="text-lg md:text-xl font-bold text-foreground">更新体重与统计</h3>
          <p class="text-xs text-muted-foreground mt-1">显示体重基于最近7天加权计算</p>
        </div>
        <button @click="$emit('close')" class="px-3 py-1 rounded-lg border border-border hover:bg-muted/60 transition-colors">关闭</button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-5">
        <div class="grid grid-cols-1 lg:grid-cols-4 gap-4 p-4 rounded-2xl border border-border bg-muted/20">
          <div class="lg:col-span-2">
            <label class="text-xs text-muted-foreground block mb-2">输入体重（{{ weightUnit === 'jin' ? '斤' : 'kg' }}）</label>
            <input
              v-model.number="newWeight"
              type="number"
              min="0"
              step="0.1"
              class="w-full rounded-xl px-4 py-3 bg-muted border border-border focus:border-primary/50"
              placeholder="例如 70"
            />
          </div>
          <div>
            <label class="text-xs text-muted-foreground block mb-2">状态</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="state in weightStates"
                :key="state.value"
                @click="weightState = state.value"
                :class="[
                  'rounded-xl px-2 py-3 text-sm border transition-colors',
                  weightState === state.value
                    ? 'bg-primary text-white border-primary'
                    : 'bg-muted border-border text-muted-foreground hover:text-foreground'
                ]"
              >
                {{ state.label }}
              </button>
            </div>
          </div>
          <div class="flex items-end">
            <button
              @click="submitWeight"
              :disabled="saving"
              class="w-full rounded-xl px-4 py-3 bg-primary text-white font-semibold disabled:opacity-50"
            >
              {{ saving ? '提交中...' : '提交体重' }}
            </button>
          </div>
        </div>

        <div class="rounded-2xl border border-border p-4 bg-muted/20 flex flex-wrap items-center justify-between gap-4">
          <div>
            <div class="text-xs text-muted-foreground">当前显示体重（7日加权）</div>
            <div class="text-2xl md:text-3xl font-black mt-1 text-foreground">{{ displayWeightText }}</div>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <button
              @click="toggleHistory"
              class="px-3 py-1.5 rounded-lg text-xs md:text-sm border border-border text-muted-foreground hover:text-foreground"
            >
              {{ showHistory ? '收起历史' : '历史记录' }}
            </button>
            <button
              v-for="mode in viewModes"
              :key="mode.value"
              @click="setView(mode.value)"
              :class="[
                'px-3 py-1.5 rounded-lg text-xs md:text-sm border transition-colors',
                chartView === mode.value ? 'bg-primary/20 border-primary text-primary' : 'border-border text-muted-foreground hover:text-foreground'
              ]"
            >
              {{ mode.label }}
            </button>
          </div>
        </div>

        <div v-if="showHistory" class="rounded-2xl border border-border p-4 bg-muted/10">
          <div class="flex items-center justify-between mb-3 text-foreground">
            <div class="text-sm font-semibold">历史体重记录</div>
            <button
              @click="loadWeightRecords"
              :disabled="historyLoading"
              class="px-3 py-1 text-xs rounded-md border border-border text-muted-foreground hover:text-foreground disabled:opacity-50"
            >
              {{ historyLoading ? '刷新中...' : '刷新' }}
            </button>
          </div>

          <div v-if="historyLoading" class="text-sm text-muted-foreground py-6 text-center">加载中...</div>
          <div v-else-if="!weightRecords.length" class="text-sm text-muted-foreground py-6 text-center">暂无记录</div>
          <div v-else class="space-y-2 max-h-64 overflow-y-auto pr-1">
            <div
              v-for="item in weightRecords"
              :key="item.id"
              class="rounded-xl border border-border bg-card p-3 flex items-center justify-between gap-3"
            >
              <div class="min-w-0">
                <div class="text-sm font-semibold">{{ item.record_date }} · {{ item.state_label }}</div>
                <div class="text-xs text-muted-foreground mt-1">
                  空腹换算 {{ formatWeight(item.fasting_weight) }} {{ weightUnit === 'jin' ? '斤' : 'kg' }}
                  <span class="mx-1">|</span>
                  原始 {{ formatWeight(item.raw_weight) }} {{ weightUnit === 'jin' ? '斤' : 'kg' }}
                </div>
              </div>
              <button
                @click="removeWeightRecord(item)"
                :disabled="deletingRecordId === item.id"
                class="px-2.5 py-1 text-xs rounded-md border border-red-500/40 text-red-300 hover:bg-red-500/10 disabled:opacity-50"
              >
                {{ deletingRecordId === item.id ? '删除中...' : '删除' }}
              </button>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-border p-4 bg-card/50">
          <div class="overflow-x-auto pb-2" @mouseleave="hoveredIndex = -1">
            <svg :width="svgWidth" height="320" class="min-w-full" @mousemove="handleChartMouseMove" @mouseleave="hoveredIndex = -1">
              <line :x1="leftPad" :x2="leftPad" :y1="topPad" :y2="bottomPad" stroke="var(--border)" />
              <line :x1="leftPad" :x2="svgWidth - rightPad" :y1="bottomPad" :y2="bottomPad" stroke="var(--border)" />

              <line
                :x1="leftPad"
                :x2="svgWidth - rightPad"
                :y1="yForWeight(referenceWeight)"
                :y2="yForWeight(referenceWeight)"
                stroke="#f59e0b"
                stroke-dasharray="6 6"
                opacity="0.85"
              />

              <g v-for="tick in yTicks" :key="tick">
                <line :x1="leftPad" :x2="svgWidth - rightPad" :y1="yForWeight(tick)" :y2="yForWeight(tick)" stroke="#27272a" />
                <text :x="leftPad - 10" :y="yForWeight(tick) + 4" text-anchor="end" fill="#a1a1aa" font-size="11">{{ tick.toFixed(1) }}</text>
              </g>

              <path :d="linePath(points)" stroke="#93c5fd" stroke-width="2" fill="none" />
              <path :d="linePath(trendPoints)" stroke="#f97316" stroke-width="2.5" fill="none" />

              <g v-for="(p, idx) in points" :key="`pt-${p.date}-${idx}`">
                <circle
                  :cx="xForTimelineIndex(timelineIndexOfPoint(p, idx))"
                  :cy="yForWeight(p.weight)"
                  r="4"
                  fill="#93c5fd"
                />
              </g>

              <line
                v-if="hoveredIndex >= 0 && points[hoveredIndex]"
                :x1="xForTimelineIndex(timelineIndexOfPoint(points[hoveredIndex], hoveredIndex))"
                :x2="xForTimelineIndex(timelineIndexOfPoint(points[hoveredIndex], hoveredIndex))"
                :y1="topPad"
                :y2="bottomPad"
                stroke="#d4d4d8"
                stroke-dasharray="4 4"
              />

              <g v-if="hoveredIndex >= 0 && points[hoveredIndex]">
                <rect
                  :x="Math.max(leftPad + 8, xForTimelineIndex(timelineIndexOfPoint(points[hoveredIndex], hoveredIndex)) - 52)"
                  :y="topPad + 4"
                  width="132"
                  height="38"
                  rx="8"
                  fill="#111827"
                  stroke="#374151"
                />
                <text
                  :x="Math.max(leftPad + 14, xForTimelineIndex(timelineIndexOfPoint(points[hoveredIndex], hoveredIndex)) - 46)"
                  :y="topPad + 20"
                  fill="#f9fafb"
                  font-size="12"
                >{{ points[hoveredIndex].date }}</text>
                <text
                  :x="Math.max(leftPad + 14, xForTimelineIndex(timelineIndexOfPoint(points[hoveredIndex], hoveredIndex)) - 46)"
                  :y="topPad + 34"
                  fill="#93c5fd"
                  font-size="12"
                >{{ points[hoveredIndex].weight.toFixed(1) }} {{ weightUnit === 'jin' ? '斤' : 'kg' }}</text>
              </g>

              <text
                v-for="(date, idx) in timelineDates"
                :key="`date-${idx}`"
                :x="xForTimelineIndex(idx)"
                :y="bottomPad + 18"
                text-anchor="middle"
                fill="#71717a"
                font-size="10"
              >{{ date.slice(5) }}</text>
            </svg>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { deleteWeightRecord, getWeightDashboard, getWeightRecords, saveWeightRecord } from '../api'
import { useToastStore } from '../api/toast'
import { useDialogStore } from '../api/dialog'

const props = defineProps({
  userId: {
    type: Number,
    required: true
  },
  defaultUnit: {
    type: String,
    default: 'kg'
  }
})

const emit = defineEmits(['close', 'updated'])
const toast = useToastStore()
const dialog = useDialogStore()

const newWeight = ref(null)
const weightUnit = ref('kg')
const weightState = ref('fasting')
const chartView = ref('day')
const points = ref([])
const trendPoints = ref([])
const timelineDates = ref([])
const displayWeight = ref(0)
const referenceWeight = ref(0)
const yMin = ref(50)
const yMax = ref(80)
const saving = ref(false)
const hoveredIndex = ref(-1)
const showHistory = ref(false)
const historyLoading = ref(false)
const weightRecords = ref([])
const deletingRecordId = ref(null)

const leftPad = 58
const rightPad = 24
const topPad = 16
const bottomPad = 286

const viewModes = [
  { value: 'day', label: '日' },
  { value: '3day', label: '三日' },
  { value: 'week', label: '周' }
]

const weightStates = [
  { value: 'fasting', label: '空腹' },
  { value: 'normal', label: '正常' },
  { value: 'full', label: '饱腹' }
]

const timelinePointCount = computed(() => Math.max(1, timelineDates.value.length))
const svgWidth = computed(() => Math.max(720, timelinePointCount.value * 72 + leftPad + rightPad))

const displayWeightText = computed(() => {
  if (!displayWeight.value) return '--'
  return `${displayWeight.value.toFixed(1)} ${weightUnit.value === 'jin' ? '斤' : 'kg'}`
})

const yTicks = computed(() => {
  const min = yMin.value
  const max = yMax.value
  const step = (max - min) / 4
  return [min, min + step, min + step * 2, min + step * 3, max]
})

function yForWeight(weight) {
  const min = yMin.value
  const max = yMax.value
  if (max <= min) return bottomPad
  const ratio = (weight - min) / (max - min)
  return bottomPad - ratio * (bottomPad - topPad)
}

function xForTimelineIndex(index) {
  if (timelinePointCount.value <= 1) return leftPad + 10
  const usable = svgWidth.value - leftPad - rightPad
  return leftPad + (usable * index) / (timelinePointCount.value - 1)
}

function timelineIndexOfPoint(point, fallbackIndex = 0) {
  const idx = Number(point?.timeline_index)
  if (Number.isFinite(idx) && idx >= 0) {
    return Math.min(idx, timelinePointCount.value - 1)
  }
  return Math.min(fallbackIndex, timelinePointCount.value - 1)
}

function linePath(arr) {
  if (!arr.length) return ''
  return arr
    .map((p, idx) => `${idx === 0 ? 'M' : 'L'} ${xForTimelineIndex(timelineIndexOfPoint(p, idx))} ${yForWeight(p.weight)}`)
    .join(' ')
}

function handleChartMouseMove(event) {
  if (!points.value.length) {
    hoveredIndex.value = -1
    return
  }
  const rect = event.currentTarget?.getBoundingClientRect?.()
  if (!rect) {
    hoveredIndex.value = -1
    return
  }
  const pointerX = event.clientX - rect.left
  let bestIndex = -1
  let bestDistance = Number.POSITIVE_INFINITY

  points.value.forEach((p, idx) => {
    const x = xForTimelineIndex(timelineIndexOfPoint(p, idx))
    const distance = Math.abs(pointerX - x)
    if (distance < bestDistance) {
      bestDistance = distance
      bestIndex = idx
    }
  })

  const step = timelinePointCount.value > 1
    ? (svgWidth.value - leftPad - rightPad) / (timelinePointCount.value - 1)
    : (svgWidth.value - leftPad - rightPad)
  hoveredIndex.value = bestDistance <= step / 2 ? bestIndex : -1
}

async function loadDashboard() {
  try {
    const res = await getWeightDashboard(props.userId, chartView.value)
    const data = res.data || {}
    if (data.weight_unit) {
      weightUnit.value = data.weight_unit
    }
    displayWeight.value = Number(data.display_weight || 0)
    referenceWeight.value = Number(data.reference_latest_weight || data.display_weight || 0)
    yMin.value = Number(data.y_axis_min || 40)
    yMax.value = Number(data.y_axis_max || 100)
    points.value = Array.isArray(data.points) ? data.points : []
    trendPoints.value = Array.isArray(data.trend_points) ? data.trend_points : []
    timelineDates.value = Array.isArray(data.timeline_dates)
      ? data.timeline_dates
      : points.value.map(item => item.date)
  } catch (e) {
    toast.add({ type: 'error', title: '读取失败', message: e.response?.data?.error || e.message })
  }
}

async function loadWeightRecords() {
  if (!showHistory.value) return
  historyLoading.value = true
  try {
    const res = await getWeightRecords(props.userId, 100)
    const data = res.data || {}
    if (data.weight_unit) {
      weightUnit.value = data.weight_unit
    }
    weightRecords.value = Array.isArray(data.items) ? data.items : []
  } catch (e) {
    toast.add({ type: 'error', title: '读取失败', message: e.response?.data?.error || e.message })
  } finally {
    historyLoading.value = false
  }
}

function toggleHistory() {
  showHistory.value = !showHistory.value
  if (showHistory.value) {
    loadWeightRecords()
  }
}

function setView(mode) {
  chartView.value = mode
}

async function submitWeight() {
  if (!newWeight.value || newWeight.value <= 0) {
    toast.add({ type: 'warning', message: '请输入有效体重' })
    return
  }
  saving.value = true
  try {
    await saveWeightRecord({
      user_id: props.userId,
      weight: Number(newWeight.value),
      weight_unit: weightUnit.value,
      state: weightState.value
    })
    toast.add({ type: 'success', message: '体重更新成功' })
    newWeight.value = null
    await loadDashboard()
    await loadWeightRecords()
    emit('updated')
  } catch (e) {
    toast.add({ type: 'error', title: '提交失败', message: e.response?.data?.error || e.message })
  } finally {
    saving.value = false
  }
}

function formatWeight(value) {
  const num = Number(value)
  if (!Number.isFinite(num)) return '--'
  return num.toFixed(1)
}

async function removeWeightRecord(item) {
  if (!item?.id) {
    return
  }
  const ok = await dialog.confirm('删除确认', '确认删除该条体重记录吗？这将影响 7 日加权体重的计算结果。')
  if (!ok) {
    return
  }
  deletingRecordId.value = item.id
  try {
    await deleteWeightRecord(props.userId, item.id)
    toast.add({ type: 'success', message: '体重记录已删除' })
    await Promise.all([loadDashboard(), loadWeightRecords()])
    emit('updated')
  } catch (e) {
    toast.add({ type: 'error', title: '删除失败', message: e.response?.data?.error || e.message })
  } finally {
    deletingRecordId.value = null
  }
}

watch(chartView, () => {
  loadDashboard()
})

onMounted(() => {
  weightUnit.value = props.defaultUnit || 'kg'
  loadDashboard()
})
</script>
