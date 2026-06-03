<template>
  <div class="h-full flex flex-col max-w-4xl mx-auto px-4 py-8">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-foreground">身体数据</h1>
        <p class="text-muted-foreground mt-1 text-sm">追踪您的身体各项指标变化</p>
      </div>
      <button 
        @click="$emit('open-profile')"
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-muted border border-border hover:bg-muted/80 transition-all text-sm font-medium"
      >
        <Edit3 class="w-4 h-4" />
        编辑数据
      </button>
    </div>

    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="w-10 h-10 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
      <!-- BMI Card -->
      <div class="col-span-full p-6 md:p-8 rounded-3xl bg-gradient-to-br from-primary/20 to-transparent border border-primary/20 flex items-center justify-between">
        <div class="space-y-2 md:space-y-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 md:w-12 md:h-12 rounded-2xl bg-primary/20 flex items-center justify-center text-primary">
              <Activity class="w-5 h-5 md:w-6 md:h-6" />
            </div>
            <h2 class="text-lg md:text-xl font-bold">当前 BMI 指数</h2>
          </div>
          <div class="flex items-baseline gap-2">
            <span class="text-4xl md:text-6xl font-black text-foreground">{{ bmi || '--' }}</span>
            <span :class="['text-sm md:text-lg font-bold px-3 py-1 rounded-lg', bmiBgColor, bmiTextColor]">
              {{ bmiStatus }}
            </span>
          </div>
          <p class="text-xs md:text-sm text-muted-foreground max-w-md hidden sm:block">
            BMI 是衡量身体胖瘦程度的一个常用指标。您的数值处于「{{ bmiStatus }}」范围。
          </p>
        </div>
        
        <!-- Simple BMI Gauge placeholder -->
        <div class="hidden lg:block w-32 h-32 relative">
          <svg viewBox="0 0 32 32" class="w-full h-full transform -rotate-90">
            <circle cx="16" cy="16" r="14" fill="none" stroke="currentColor" stroke-width="3" class="text-muted" />
            <circle 
              cx="16" cy="16" r="14" fill="none" stroke="currentColor" stroke-width="3" 
              stroke-dasharray="88" 
              :stroke-dashoffset="88 - (Math.min(parseFloat(bmi || 0), 40) / 40 * 88)"
              :class="bmiTextColor" 
            />
          </svg>
          <div class="absolute inset-0 flex items-center justify-center font-bold text-xs uppercase tracking-tighter">
            Gauge
          </div>
        </div>
      </div>

      <!-- Stats Cards -->
      <div class="col-span-full grid grid-cols-3 gap-3 md:grid-cols-3 md:gap-6 lg:grid-cols-3">
        <div
          v-for="stat in stats"
          :key="stat.label"
          @click="handleStatCardClick(stat.label)"
          :class="[
            'p-4 md:p-6 rounded-2xl bg-muted/30 border border-border flex flex-col items-center justify-center text-center gap-2 aspect-square md:aspect-auto md:items-start md:text-left',
            stat.label === '体重' ? 'cursor-pointer hover:border-primary/40 transition-colors' : ''
          ]"
        >
          <div class="flex items-center justify-between w-full md:flex-row flex-col gap-2">
            <div class="w-8 h-8 md:w-10 md:h-10 rounded-xl bg-muted flex items-center justify-center text-muted-foreground">
              <component :is="stat.icon" class="w-4 h-4 md:w-5 md:h-5" />
            </div>
            <span class="text-[10px] md:text-xs font-bold text-muted-foreground uppercase tracking-widest">{{ stat.label }}</span>
          </div>
          <div class="flex items-baseline gap-1 mt-1">
            <span class="text-xl md:text-2xl lg:text-3xl font-bold text-foreground">{{ stat.value || '--' }}</span>
            <span class="text-[10px] md:text-sm text-muted-foreground">{{ stat.unit }}</span>
          </div>
        </div>
      </div>

      <!-- Goal & Intensity Row -->
      <div class="col-span-full grid grid-cols-1 md:grid-cols-3 gap-3 md:gap-6">
        <!-- Goal Card -->
        <div class="md:col-span-2 p-4 md:p-6 rounded-2xl bg-card border border-border flex items-center gap-4 md:gap-6">
          <div class="w-12 h-12 md:w-16 md:h-16 rounded-2xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
            <Target class="w-6 h-6 md:w-8 md:h-8" />
          </div>
          <div class="min-w-0">
            <div class="text-[10px] md:text-xs font-bold text-muted-foreground uppercase tracking-widest mb-1">核心诉求</div>
            <div class="text-base md:text-lg lg:text-xl font-bold text-foreground truncate">{{ profileGoalText }}</div>
            <p class="text-[10px] md:text-sm text-muted-foreground mt-0.5 truncate md:whitespace-normal">
              {{ profile?.sports_description || profile?.current_exercise_desc || '您的 AI 计划将围绕这一目标为您量身定制。' }}
            </p>
          </div>
        </div>

        <!-- Intensity Card -->
        <div class="p-4 md:p-6 rounded-2xl bg-card border border-border flex flex-col justify-between">
          <div>
            <div class="text-[10px] md:text-xs font-bold text-muted-foreground uppercase tracking-widest mb-1">训练强度</div>
            <div class="text-base md:text-lg lg:text-xl font-bold text-foreground capitalize">{{ intensityName }}</div>
          </div>
          <div class="mt-2 md:mt-4 flex gap-1">
            <div v-for="i in 3" :key="i" :class="['h-1.5 flex-1 rounded-full', i <= intensityLevel ? 'bg-primary' : 'bg-muted']"></div>
          </div>
        </div>
      </div>
    </div>

    <WeightUpdateModal
      v-if="showWeightModal && Number(userId) > 0"
      :user-id="Number(userId)"
      :default-unit="weightUnit"
      @close="showWeightModal = false"
      @updated="fetchProfile"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Activity, 
  Ruler, 
  Weight, 
  User, 
  Target, 
  Edit3, 
  Zap 
} from 'lucide-vue-next'
import { getProfile } from '../api'
import WeightUpdateModal from './WeightUpdateModal.vue'

defineEmits(['open-profile'])

const profile = ref(null)
const loading = ref(true)
const userId = localStorage.getItem('user_id')
const showWeightModal = ref(false)

const weightUnit = computed(() => profile.value?.weight_unit || 'kg')

const stats = computed(() => {
  const unit = profile.value?.weight_unit || 'kg'
  const weightDisplay = profile.value?.weight
  const weightRawKg = profile.value?.weight_kg
  const weightValue =
    weightDisplay != null
      ? weightDisplay
      : unit === 'jin'
        ? (Number(weightRawKg || 0) * 2).toFixed(1)
        : weightRawKg

  return [
    { label: '身高', value: profile.value?.height ?? profile.value?.height_cm, unit: 'cm', icon: Ruler },
    { label: '体重', value: weightValue, unit: unit === 'jin' ? '斤' : 'kg', icon: Weight },
    { label: '年龄', value: profile.value?.age, unit: '岁', icon: User },
  ]
})

const bmi = computed(() => {
  const backendBmi = Number(profile.value?.bmi)
  if (Number.isFinite(backendBmi) && backendBmi > 0) {
    return backendBmi.toFixed(1)
  }
  const hVal = profile.value?.height ?? profile.value?.height_cm
  const weightDisplay = profile.value?.weight ?? profile.value?.weight_kg
  const wVal = weightUnit.value === 'jin' ? Number(weightDisplay || 0) / 2 : Number(weightDisplay || 0)
  if (!hVal || !wVal) return null
  const h = hVal / 100
  return (wVal / (h * h)).toFixed(1)
})

const bmiStatus = computed(() => {
  const backendStatus = String(profile.value?.bmi_status || '').trim()
  if (backendStatus) {
    return backendStatus
  }
  const val = parseFloat(bmi.value)
  if (!val) return '未知'
  if (val < 18.5) return '偏瘦'
  if (val < 24) return '正常'
  if (val < 28) return '超重'
  return '肥胖'
})

const bmiTextColor = computed(() => {
  const s = bmiStatus.value
  if (s === '正常') return 'text-green-500'
  if (s === '偏瘦') return 'text-blue-500'
  if (s === '超重') return 'text-yellow-500'
  return 'text-red-500'
})

const bmiBgColor = computed(() => {
  const s = bmiStatus.value
  if (s === '正常') return 'bg-green-500/10'
  if (s === '偏瘦') return 'bg-blue-500/10'
  if (s === '超重') return 'bg-yellow-500/10'
  return 'bg-red-500/10'
})

const intensityName = computed(() => {
  const levels = { beginner: '新手', normal: '正常', hardcore: '硬汉' }
  const key = profile.value?.target_intensity || profile.value?.expected_intensity
  return levels[key] || '未设置'
})

const intensityLevel = computed(() => {
  const levels = { beginner: 1, normal: 2, hardcore: 3 }
  const key = profile.value?.target_intensity || profile.value?.expected_intensity
  return levels[key] || 0
})

const profileGoalText = computed(() => {
  if (Array.isArray(profile.value?.core_goals) && profile.value.core_goals.length) {
    return profile.value.core_goals.join('、')
  }
  if (profile.value?.core_goal) {
    return profile.value.core_goal
  }
  return '未设置'
})

async function fetchProfile() {
  if (!userId) return
  loading.value = true
  try {
    const res = await getProfile(userId)
    profile.value = res.data
  } catch (e) {
    console.error('Failed to fetch profile', e)
  } finally {
    loading.value = false
  }
}

function handleStatCardClick(label) {
  if (label === '体重') {
    showWeightModal.value = true
  }
}

onMounted(() => {
  fetchProfile()
})
</script>
