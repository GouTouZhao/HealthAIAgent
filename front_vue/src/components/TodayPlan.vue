<template>
  <div class="h-full flex flex-col max-w-4xl mx-auto px-4 py-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-foreground">今日计划</h1>
        <p class="text-muted-foreground mt-1 text-xs md:text-sm">{{ todayFormatted }}</p>
      </div>
      <div class="text-right">
        <div class="text-sm font-medium text-muted-foreground mb-2 hidden md:block">今日完成度: {{ completedCount }}/{{ totalCount }}</div>
        <div class="w-32 md:w-48 h-2 md:h-2.5 bg-muted rounded-full overflow-hidden border border-border">
          <div 
            class="h-full bg-primary transition-all duration-500 ease-out"
            :style="{ width: `${progress}%` }"
          ></div>
        </div>
      </div>
    </div>

    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="flex flex-col items-center gap-4 opacity-50">
        <div class="w-10 h-10 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
        <p>正在获取您的计划...</p>
      </div>
    </div>

    <div v-else-if="tasks.length === 0" class="flex-1 flex flex-col items-center justify-center text-center opacity-40 gap-4">
      <div class="w-20 h-20 rounded-3xl bg-muted flex items-center justify-center">
        <Calendar class="w-10 h-10" />
      </div>
      <div>
        <h3 class="text-xl font-semibold">今日无安排</h3>
        <p class="mt-2">去 AI 对话模式下制定一个计划吧！</p>
      </div>
    </div>

    <div v-else class="flex-1 space-y-4">
      <div v-if="dayMeta.intensity || dayMeta.duration || dayMeta.restInterval" class="p-4 rounded-2xl border border-border bg-card shadow-sm">
        <div class="flex flex-col gap-2">
          <!-- First Row: Intensity -->
          <div class="flex items-center justify-between p-3 md:p-4 rounded-xl bg-primary/5 border border-primary/10">
            <div class="text-[10px] md:text-xs text-muted-foreground uppercase tracking-widest font-bold">今日强度</div>
            <div class="text-lg md:text-2xl font-black text-primary">{{ dayMeta.intensity || '-' }}</div>
          </div>
          <!-- Second Row: Duration and Rest -->
          <div class="grid grid-cols-2 gap-2">
            <div class="flex flex-col items-start gap-0.5 p-3 md:p-4 rounded-xl bg-muted/20">
              <div class="text-[10px] md:text-xs text-muted-foreground uppercase tracking-widest font-bold">预计时长</div>
              <div class="text-sm md:text-lg font-bold text-foreground truncate">{{ dayMeta.duration || '-' }}</div>
            </div>
            <div class="flex flex-col items-start gap-0.5 p-3 md:p-4 rounded-xl bg-muted/20">
              <div class="text-[10px] md:text-xs text-muted-foreground uppercase tracking-widest font-bold">组间间歇</div>
              <div class="text-sm md:text-lg font-bold text-foreground truncate">{{ dayMeta.restInterval || '-' }}</div>
            </div>
          </div>
        </div>
      </div>

      <div 
        v-for="(task, idx) in tasks" 
        :key="`${task.activity}-${idx}`"
        class="group p-4 rounded-2xl border border-border bg-card hover:border-primary/30 transition-all duration-300"
      >
        <div class="flex items-center gap-4">
          <button 
            @click="toggleTaskStatus(task)"
            :class="[
              'w-8 h-8 rounded-xl flex items-center justify-center border-2 transition-all duration-300',
              task.completed ? 'bg-primary border-primary text-white' : 'border-border text-transparent hover:border-primary/50'
            ]"
          >
            <Check class="w-5 h-5" />
          </button>
          
          <div class="flex-1">
            <h4 :class="['font-bold text-base md:text-lg transition-all', task.completed ? 'text-muted-foreground line-through' : 'text-foreground']">
              {{ task.activity }}
            </h4>
            <div class="flex gap-4 mt-1 text-sm text-muted-foreground">
              <span v-if="formatTaskSetsReps(task)" class="flex items-center gap-1"><Layers class="w-3.5 h-3.5" /> {{ formatTaskSetsReps(task) }}</span>
              <span v-else-if="task.sets" class="flex items-center gap-1"><Layers class="w-3.5 h-3.5" /> {{ task.sets }} 组</span>
              <span v-if="!formatTaskSetsReps(task) && task.reps" class="flex items-center gap-1"><Repeat class="w-3.5 h-3.5" /> {{ task.reps }} 次</span>
              <span v-if="!formatTaskSetsReps(task) && task.duration_min" class="flex items-center gap-1"><Clock class="w-3.5 h-3.5" /> {{ task.duration_min }} 分</span>
            </div>
          </div>

          <div v-if="task.completed" class="text-xs font-bold text-primary uppercase tracking-widest animate-in fade-in zoom-in">
            已完成
          </div>
        </div>
      </div>

      <div class="pt-3 flex justify-center">
        <button
          @click="showWeightModal = true"
          class="px-5 py-2 rounded-xl bg-primary text-white border border-primary shadow-lg shadow-primary/20 hover:bg-primary/90 text-sm font-semibold"
        >
          更新今日体重
        </button>
      </div>
    </div>

    <WeightUpdateModal
      v-if="showWeightModal && Number(userId) > 0"
      :user-id="Number(userId)"
      :default-unit="weightUnit"
      @close="showWeightModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Calendar, Check, Clock, Layers, Repeat } from 'lucide-vue-next'
import { getTodayPlan, getProfile, toggleTask } from '../api'
import WeightUpdateModal from './WeightUpdateModal.vue'

const tasks = ref([])
const loading = ref(true)
const userId = localStorage.getItem('user_id')
const taskDate = ref('')
const showWeightModal = ref(false)
const weightUnit = ref('kg')
const dayMeta = ref({ intensity: '', duration: '', restInterval: '' })

const todayFormatted = computed(() => {
  return new Intl.DateTimeFormat('zh-CN', { 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric', 
    weekday: 'long' 
  }).format(new Date())
})

const completedCount = computed(() => tasks.value.filter(t => t.completed).length)
const totalCount = computed(() => tasks.value.length)
const progress = computed(() => totalCount.value === 0 ? 0 : (completedCount.value / totalCount.value) * 100)

async function fetchPlan() {
  if (!userId) return
  loading.value = true
  try {
    const res = await getTodayPlan(userId)
    const payload = res.data || {}
    taskDate.value = payload.task_date || ''
    tasks.value = Array.isArray(payload.tasks) ? payload.tasks : []
    const firstTask = tasks.value[0] || {}
    dayMeta.value = {
      intensity: firstTask.day_intensity || firstTask.intensity || '',
      duration: firstTask.estimated_duration || (Number.isFinite(firstTask.duration_min) ? `约${firstTask.duration_min}分钟` : ''),
      restInterval: firstTask.rest_interval || ''
    }
  } catch (e) {
    console.error('Failed to fetch today plan', e)
  } finally {
    loading.value = false
  }
}

function formatTaskSetsReps(task) {
  const raw = typeof task?.sets_reps === 'string' ? task.sets_reps.trim() : ''
  return raw || ''
}

async function loadProfileUnit() {
  if (!userId) return
  try {
    const res = await getProfile(userId)
    weightUnit.value = res.data?.weight_unit || 'kg'
  } catch (e) {
    console.error('Failed to load profile unit', e)
  }
}

async function toggleTaskStatus(task) {
  const originalStatus = task.completed
  task.completed = !task.completed // Optimistic update
  
  try {
    await toggleTask({
      user_id: parseInt(userId),
      task_date: taskDate.value || new Date().toISOString().slice(0, 10),
      activity: task.activity,
      completed: task.completed
    })
  } catch (e) {
    task.completed = originalStatus // Rollback
    console.error('Failed to toggle task', e)
  }
}

onMounted(() => {
  fetchPlan()
  loadProfileUnit()
})
</script>
