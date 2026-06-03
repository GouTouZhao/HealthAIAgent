<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-2 md:p-4 bg-background/80 backdrop-blur-md animate-in fade-in duration-300">
    <div class="bg-card border border-border w-full max-w-3xl rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[95vh] animate-in zoom-in-95 duration-300">
      <div class="px-4 md:px-8 py-4 md:py-6 border-b border-border flex items-center justify-between">
        <div>
          <h2 class="text-xl md:text-2xl font-bold text-foreground">完善身体信息</h2>
          <p class="text-xs text-muted-foreground mt-1">出生日期、身高、体重、性别、强度、核心诉求为必填</p>
        </div>
        <button @click="$emit('close')" class="p-2 hover:bg-muted rounded-xl transition-colors">
          <X class="w-6 h-6" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-4 md:px-8 py-6 md:py-8 space-y-8">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">身高 (cm) *</label>
            <input v-model.number="form.height" type="number" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50" placeholder="例如: 175" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">体重 *</label>
            <div class="grid grid-cols-[1fr_96px] gap-2">
              <input
                v-model.number="form.weight"
                type="number"
                :readonly="profileExists"
                @click="openWeightUpdate"
                class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50"
                :class="profileExists ? 'cursor-pointer hover:border-primary/40' : ''"
                :placeholder="form.weight_unit === 'jin' ? '例如: 140' : '例如: 70'"
              />
              <div class="w-full bg-muted border border-border rounded-xl p-1 grid grid-cols-2 gap-1">
                <button
                  type="button"
                  @click="handleWeightUnitChange('kg')"
                  :class="[
                    'rounded-lg text-xs font-semibold transition-colors',
                    form.weight_unit === 'kg' ? 'bg-primary text-white' : 'text-muted-foreground hover:text-white'
                  ]"
                >
                  KG
                </button>
                <button
                  type="button"
                  @click="handleWeightUnitChange('jin')"
                  :class="[
                    'rounded-lg text-xs font-semibold transition-colors',
                    form.weight_unit === 'jin' ? 'bg-primary text-white' : 'text-muted-foreground hover:text-white'
                  ]"
                >
                  斤
                </button>
              </div>
            </div>
            <p v-if="profileExists" class="text-xs text-muted-foreground">体重请点击输入框进入“更新体重”界面，通过算法更新。</p>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">出生日期 *</label>
            <input v-model="form.birth_date" type="date" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">性别 *</label>
            <select v-model="form.gender" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 appearance-none">
              <option value="male">男</option>
              <option value="female">女</option>
              <option value="other">其他</option>
            </select>
          </div>
        </div>

        <div v-if="bmi" class="p-4 rounded-2xl bg-primary/5 border border-primary/20 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-xl bg-primary/20 flex items-center justify-center text-primary">
              <Activity class="w-6 h-6" />
            </div>
            <div>
              <div class="text-xs font-bold text-primary uppercase tracking-widest">BMI</div>
              <div class="text-2xl font-black text-foreground">{{ bmi }}</div>
            </div>
          </div>
          <div :class="['font-bold', bmiColor]">{{ bmiStatus }}</div>
        </div>

        <div class="space-y-6 pt-4 border-t border-border">
          <div class="space-y-3">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">核心诉求（可多选）*</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="goal in goals"
                :key="goal"
                @click="toggleGoal(goal)"
                :class="[
                  'px-4 py-2 rounded-xl text-sm font-medium border transition-all duration-200',
                  form.core_goals.includes(goal) ? 'bg-primary border-primary text-white shadow-lg shadow-primary/20' : 'bg-muted border-border text-muted-foreground hover:border-primary/50'
                ]"
              >
                {{ goal }}
              </button>
            </div>
          </div>

          <div v-if="form.core_goals.includes('其他')" class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">详细诉求（选择“其他”时必填）*</label>
            <textarea v-model="form.detailed_goal" rows="3" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 resize-none" placeholder="请详细描述您的目标"></textarea>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">伤病史</label>
            <textarea v-model="form.injury_history" rows="2" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 resize-none" placeholder="如：腰椎间盘突出、膝盖损伤等"></textarea>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">喜欢的运动</label>
            <input v-model="form.favorite_sports" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50" placeholder="例如: 跑步, 游泳, 力量训练" />
          </div>

          <div class="flex items-center justify-between p-4 bg-muted/30 rounded-2xl border border-border">
            <div>
              <div class="font-bold">将喜欢的运动加入训练计划</div>
              <div class="text-xs text-muted-foreground">开启后才会显示运动描述</div>
            </div>
            <button
              @click="form.auto_add_to_plan = !form.auto_add_to_plan"
              :class="['w-12 h-6 rounded-full transition-all duration-300 relative', form.auto_add_to_plan ? 'bg-primary' : 'bg-border']"
            >
              <div :class="['absolute top-1 w-4 h-4 bg-white rounded-full transition-all duration-300', form.auto_add_to_plan ? 'left-7' : 'left-1']"></div>
            </button>
          </div>

          <div v-if="form.auto_add_to_plan" class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">运动描述</label>
            <textarea v-model="form.sports_description" rows="3" class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 resize-none" placeholder="描述您的运动习惯，例如：每周3次，每次1小时，中等强度"></textarea>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">期望训练强度 *</label>
            <div class="grid grid-cols-3 gap-3">
              <button
                v-for="level in intensityLevels"
                :key="level.val"
                @click="form.target_intensity = level.val"
                :class="[
                  'px-3 py-3 rounded-xl border flex flex-col items-center gap-1 transition-all',
                  form.target_intensity === level.val ? 'bg-primary/10 border-primary text-primary shadow-lg shadow-primary/10' : 'bg-muted border-border text-muted-foreground hover:border-primary/50'
                ]"
              >
                <component :is="level.icon" class="w-5 h-5" />
                <span class="text-[10px] font-bold uppercase tracking-wider">{{ level.name }}</span>
              </button>
            </div>
          </div>

          <div class="space-y-3 pt-2 border-t border-border">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="flex items-center gap-2">
                  <div class="font-bold">个人记忆</div>
                  <Crown v-if="membershipType === 'FREE'" class="w-3 h-3 text-orange-500" />
                </div>
                <div class="text-xs text-muted-foreground">可查看并删除系统已记录的长期偏好与限制</div>
              </div>
              <button
                type="button"
                @click="toggleMemoryPanel"
                class="px-4 py-2 rounded-xl border border-border text-sm font-semibold hover:bg-muted transition-all"
              >
                {{ memoryPanelVisible ? '收起记忆' : '查看记忆' }}
              </button>
            </div>

            <div v-if="memoryPanelVisible" class="rounded-2xl border border-border bg-muted/20 p-4 space-y-3">
              <div v-if="memoryLoading" class="text-sm text-muted-foreground">加载中...</div>
              <template v-else>
                <div v-if="!memoryItems.length" class="text-sm text-muted-foreground">暂无记忆</div>
                <div v-else class="space-y-2 max-h-56 overflow-y-auto pr-1">
                  <div
                    v-for="item in memoryItems"
                    :key="item.id"
                    class="flex items-start justify-between gap-3 rounded-xl border border-border bg-background/60 p-3"
                  >
                    <div class="space-y-1">
                      <div class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider bg-primary/10 text-primary">
                        {{ memoryTagLabel(item.tag) }}
                      </div>
                      <div class="text-sm leading-relaxed">{{ item.memory }}</div>
                    </div>
                    <button
                      type="button"
                      @click="handleDeleteMemory(item.id)"
                      :disabled="memoryDeletingId === item.id"
                      class="px-3 py-1.5 text-xs rounded-lg border border-red-500/30 text-red-400 hover:bg-red-500/10 disabled:opacity-40"
                    >
                      {{ memoryDeletingId === item.id ? '删除中...' : '删除' }}
                    </button>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>

      <div class="px-4 md:px-8 py-4 md:py-6 border-t border-border flex gap-3 md:gap-4">
        <button @click="$emit('close')" class="flex-1 px-4 py-3 md:px-6 md:py-4 rounded-2xl border border-border font-bold hover:bg-muted transition-all text-sm md:text-base">取消</button>
        <button
          @click="handleSubmit"
          :disabled="!isFormValid || saving"
          class="flex-[2] px-4 py-3 md:px-6 md:py-4 bg-primary text-white rounded-2xl font-bold shadow-lg shadow-primary/20 hover:bg-primary/90 disabled:opacity-30 disabled:cursor-not-allowed transition-all active:scale-95 text-sm md:text-base"
        >
          {{ saving ? '保存中...' : '保存更改' }}
        </button>
      </div>
    </div>

  </div>

  <WeightUpdateModal
    v-if="showWeightModal && userId > 0"
    :user-id="userId"
    :default-unit="form.weight_unit"
    @close="showWeightModal = false"
    @updated="handleWeightUpdated"
  />
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { X, Activity, Flame, Dumbbell, ShieldAlert, Crown } from 'lucide-vue-next'
import { getProfile, saveProfile, getMemoryList, deleteMemory } from '../api'
import { useToastStore } from '../api/toast'
import { useDialogStore } from '../api/dialog'
import WeightUpdateModal from './WeightUpdateModal.vue'

const props = defineProps({
  membershipType: {
    type: String,
    default: 'FREE'
  }
})

const emit = defineEmits(['close', 'updated', 'open-membership'])
const dialog = useDialogStore()
const userId = Number(localStorage.getItem('user_id') || 0)
const saving = ref(false)
const toast = useToastStore()
const profileExists = ref(false)
const showWeightModal = ref(false)
const memoryPanelVisible = ref(false)
const memoryLoading = ref(false)
const memoryDeletingId = ref(0)
const memoryItems = ref([])

const form = ref({
  user_id: userId,
  height: null,
  weight: null,
  weight_unit: 'kg',
  birth_date: '',
  gender: 'male',
  core_goals: [],
  detailed_goal: '',
  injury_history: '',
  favorite_sports: '',
  auto_add_to_plan: false,
  sports_description: '',
  target_intensity: ''
})

const goals = [
  '减脂',
  '增肌',
  '塑形',
  '体能提升',
  '马拉松备战',
  '产后恢复',
  '康复训练',
  '提高柔韧性',
  '改善代谢',
  '其他'
]

const intensityLevels = [
  { val: 'beginner', name: '新手', icon: Flame },
  { val: 'normal', name: '正常', icon: Dumbbell },
  { val: 'hardcore', name: '硬汉', icon: ShieldAlert }
]

const bmi = computed(() => {
  if (!form.value.height || !form.value.weight) return null
  const weightKg = form.value.weight_unit === 'jin' ? form.value.weight / 2 : form.value.weight
  const h = form.value.height / 100
  return (weightKg / (h * h)).toFixed(1)
})

const bmiStatus = computed(() => {
  const val = parseFloat(bmi.value)
  if (val < 18.5) return '偏瘦'
  if (val < 24) return '正常'
  if (val < 28) return '超重'
  return '肥胖'
})

const bmiColor = computed(() => {
  const status = bmiStatus.value
  if (status === '正常') return 'text-green-500'
  if (status === '偏瘦') return 'text-blue-500'
  if (status === '超重') return 'text-yellow-500'
  return 'text-red-500'
})

const isFormValid = computed(() => {
  if (!form.value.birth_date || !form.value.height || !form.value.weight || !form.value.gender || !form.value.target_intensity) {
    return false
  }
  if (!form.value.core_goals.length) {
    return false
  }
  if (form.value.core_goals.includes('其他') && !form.value.detailed_goal.trim()) {
    return false
  }
  return true
})

function toggleGoal(goal) {
  const idx = form.value.core_goals.indexOf(goal)
  if (idx >= 0) {
    form.value.core_goals.splice(idx, 1)
  } else {
    form.value.core_goals.push(goal)
  }
}

function handleWeightUnitChange(nextUnit) {
  const prevUnit = form.value.weight_unit
  if (!nextUnit || nextUnit === prevUnit) {
    return
  }
  if (form.value.weight) {
    const weight = Number(form.value.weight)
    if (prevUnit === 'kg' && nextUnit === 'jin') {
      form.value.weight = Number((weight * 2).toFixed(1))
    } else if (prevUnit === 'jin' && nextUnit === 'kg') {
      form.value.weight = Number((weight / 2).toFixed(1))
    }
  }
  form.value.weight_unit = nextUnit
}

function openWeightUpdate() {
  if (!profileExists.value) {
    return
  }
  showWeightModal.value = true
}

function memoryTagLabel(tag) {
  const normalized = String(tag || '').toLowerCase()
  if (normalized === 'constraint') return '限制'
  if (normalized === 'injury') return '伤病'
  if (normalized === 'goal') return '目标'
  if (normalized === 'preference') return '偏好'
  if (normalized === 'schedule') return '时间'
  if (normalized === 'device') return '器械'
  return '其他'
}

async function loadMemoryItems() {
  if (!userId) return
  memoryLoading.value = true
  try {
    const res = await getMemoryList(userId)
    const items = Array.isArray(res?.data?.items) ? res.data.items : []
    memoryItems.value = items
  } catch (e) {
    toast.add({ type: 'error', title: '加载失败', message: e.response?.data?.error || e.message })
  } finally {
    memoryLoading.value = false
  }
}

async function toggleMemoryPanel() {
  if (props.membershipType === 'FREE') {
    const ok = await dialog.confirm('PRO 会员专属', '个性化记忆是 PRO 会员功能，是否前往开通会员？(新用户可免费领一个月)')
    if (ok) {
      emit('open-membership')
    }
    return
  }
  memoryPanelVisible.value = !memoryPanelVisible.value
  if (memoryPanelVisible.value) {
    await loadMemoryItems()
  }
}

async function handleDeleteMemory(memoryId) {
  if (!memoryId || !userId) return
  const ok = await dialog.confirm('删除确认', '确认删除这条记忆吗？系统将不再针对此偏好进行优化。')
  if (!ok) return
  memoryDeletingId.value = Number(memoryId)
  try {
    await deleteMemory(userId, memoryId)
    toast.add({ type: 'success', message: '记忆已删除' })
    await loadMemoryItems()
  } catch (e) {
    toast.add({ type: 'error', title: '删除失败', message: e.response?.data?.error || e.message })
  } finally {
    memoryDeletingId.value = 0
  }
}

async function loadData() {
  if (!userId) return
  try {
    const res = await getProfile(userId)
    if (res.data) {
      form.value.user_id = userId
      profileExists.value = !!res.data.profile_exists
      if (res.data.profile_exists) {
        form.value.height = res.data.height ?? res.data.height_cm ?? null
        form.value.weight = res.data.weight ?? res.data.weight_kg ?? null
        form.value.weight_unit = res.data.weight_unit || 'kg'
        form.value.birth_date = res.data.birth_date || ''
        form.value.gender = res.data.gender || 'male'
        form.value.core_goals = Array.isArray(res.data.core_goals) ? res.data.core_goals : []
        form.value.detailed_goal = res.data.detailed_goal || ''
        form.value.injury_history = res.data.injury_history || ''
        form.value.favorite_sports = res.data.favorite_sports || ''
        form.value.auto_add_to_plan = !!res.data.auto_add_to_plan
        form.value.sports_description = res.data.sports_description || ''
        form.value.target_intensity = res.data.target_intensity || ''
      }
    }
  } catch (e) {
    console.error('Failed to load profile', e)
  }
}

async function handleSubmit() {
  if (!isFormValid.value) {
    toast.add({ type: 'warning', message: '请先完成所有必填项' })
    return
  }
  saving.value = true
  try {
    await saveProfile({
      user_id: form.value.user_id,
      height: form.value.height,
      weight: form.value.weight,
      weight_unit: form.value.weight_unit,
      birth_date: form.value.birth_date,
      gender: form.value.gender,
      core_goals: form.value.core_goals,
      detailed_goal: form.value.detailed_goal,
      injury_history: form.value.injury_history,
      favorite_sports: form.value.favorite_sports,
      auto_add_to_plan: form.value.auto_add_to_plan,
      sports_description: form.value.auto_add_to_plan ? form.value.sports_description : '',
      target_intensity: form.value.target_intensity
    })
    toast.add({ type: 'success', message: '个人信息保存成功' })
    emit('updated')
    emit('close')
  } catch (e) {
    toast.add({ type: 'error', title: '保存失败', message: e.response?.data?.error || e.message })
  } finally {
    saving.value = false
  }
}

async function handleWeightUpdated() {
  showWeightModal.value = false
  await loadData()
  emit('updated')
}

onMounted(() => {
  loadData()
})
</script>
