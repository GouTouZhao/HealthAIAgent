<template>
  <div class="space-y-6 p-4">
    <div v-for="month in data.months" :key="month.month" class="space-y-4">
      <div class="flex items-center justify-between border-b border-border pb-2">
        <h3 class="text-lg font-bold text-primary">第 {{ month.month }} 个月</h3>
        <span class="text-xs text-muted-foreground uppercase tracking-widest">{{ month.goal }}</span>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="p-3 bg-muted/30 rounded-xl border border-border">
          <div class="text-xs font-semibold text-muted-foreground mb-1 uppercase">本月目标</div>
          <div class="text-sm text-foreground">{{ month.goal }}</div>
        </div>
        <div class="p-3 bg-muted/30 rounded-xl border border-border">
          <div class="text-xs font-semibold text-muted-foreground mb-1 uppercase">饮食建议</div>
          <div class="text-sm text-foreground">{{ month.diet_advice }}</div>
        </div>
      </div>

      <div class="overflow-x-auto rounded-xl border border-border">
        <table class="w-full text-xs text-left">
          <thead class="bg-muted text-muted-foreground">
            <tr>
              <th class="px-3 py-2 min-w-[88px] border-r border-border">周几</th>
              <th class="px-3 py-2 min-w-[180px] border-r border-border">Activity</th>
              <th class="px-3 py-2 min-w-[140px] border-r border-border">组次</th>
              <th class="px-3 py-2 min-w-[120px] border-r border-border">强度</th>
              <th class="px-3 py-2 min-w-[120px] border-r border-border">时间</th>
              <th class="px-3 py-2 min-w-[140px]">组间间歇</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="day in weekdays" :key="day.id">
              <tr
                v-for="(task, idx) in getTasksByDay(month, day.id)"
                :key="`${day.id}-${idx}`"
                class="border-t border-border"
              >
                <td
                  v-if="idx === 0"
                  :rowspan="getTasksByDay(month, day.id).length"
                  class="px-3 py-2 align-top border-r border-border font-semibold text-foreground"
                >
                  {{ day.name }}
                </td>
                <td class="px-3 py-2 border-r border-border text-foreground">
                  {{ task.activity === 'rest' ? 'rest' : (task.activity || '-') }}
                </td>
                <td class="px-3 py-2 border-r border-border text-muted-foreground">
                  {{ formatSetsReps(task) }}
                </td>
                <td
                  v-if="idx === 0"
                  :rowspan="getTasksByDay(month, day.id).length"
                  class="px-3 py-2 align-top border-r border-border text-primary/80"
                >
                  {{ getDayPlan(month, day.id).dayIntensity }}
                </td>
                <td
                  v-if="idx === 0"
                  :rowspan="getTasksByDay(month, day.id).length"
                  class="px-3 py-2 align-top border-r border-border text-muted-foreground"
                >
                  {{ getDayPlan(month, day.id).estimatedDuration }}
                </td>
                <td
                  v-if="idx === 0"
                  :rowspan="getTasksByDay(month, day.id).length"
                  class="px-3 py-2 align-top text-muted-foreground"
                >
                  {{ getDayPlan(month, day.id).restInterval }}
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  data: {
    type: Object,
    required: true
  }
})

const weekdays = [
  { id: 'monday', name: '周一' },
  { id: 'tuesday', name: '周二' },
  { id: 'wednesday', name: '周三' },
  { id: 'thursday', name: '周四' },
  { id: 'friday', name: '周五' },
  { id: 'saturday', name: '周六' },
  { id: 'sunday', name: '周日' },
]

function getTasksByDay(month, dayId) {
  const dayPlan = getDayPlan(month, dayId)
  if (dayPlan.activities.length > 0) {
    return dayPlan.activities
  }
  return [{ activity: '-', sets_reps: '', sets: null, reps: null, duration_min: null }]
}

function getDayPlan(month, dayId) {
  const raw = month?.weekly_plan?.[dayId]
  if (!raw) {
    return {
      activities: [],
      dayIntensity: '-',
      estimatedDuration: '-',
      restInterval: '-'
    }
  }

  if (Array.isArray(raw)) {
    return {
      activities: raw,
      dayIntensity: raw[0]?.day_intensity || raw[0]?.intensity || '-',
      estimatedDuration: raw[0]?.estimated_duration || formatDuration(raw[0]),
      restInterval: raw[0]?.rest_interval || '-'
    }
  }

  const activities = Array.isArray(raw.activities) ? raw.activities : []
  return {
    activities,
    dayIntensity: raw.day_intensity || '-',
    estimatedDuration: raw.estimated_duration || '-',
    restInterval: raw.rest_interval || '-'
  }
}

function formatSetsReps(task) {
  if (typeof task?.sets_reps === 'string' && task.sets_reps.trim()) {
    return task.sets_reps
  }
  const sets = Number.isFinite(task?.sets) ? `${task.sets}组` : null
  const reps = Number.isFinite(task?.reps) ? `${task.reps}次` : null
  const duration = Number.isFinite(task?.duration_min) ? `${task.duration_min}分钟` : null
  if (sets && reps) {
    return `${reps}*${sets}`
  }
  if (duration && sets) {
    return `${duration}*${sets}`
  }
  return sets || reps || duration || '-'
}

function formatDuration(task) {
  if (Number.isFinite(task?.duration_min)) {
    return `约${task.duration_min}分钟`
  }
  return '-'
}
</script>
