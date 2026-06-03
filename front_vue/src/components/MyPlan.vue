<template>
  <div class="h-full flex flex-col max-w-6xl mx-auto px-4 py-8">
    <div class="mb-8">
      <h1 class="text-2xl md:text-3xl font-bold text-foreground">我的训练计划</h1>
      <p class="text-muted-foreground mt-1 text-sm">查看当前计划，并管理最多2条历史计划</p>
    </div>

    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="w-10 h-10 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else-if="!plan" class="flex-1 flex flex-col items-center justify-center text-center opacity-40 gap-4">
      <div class="w-20 h-20 rounded-3xl bg-muted flex items-center justify-center">
        <LayoutDashboard class="w-10 h-10" />
      </div>
      <div>
        <h3 class="text-xl font-semibold">暂无计划</h3>
        <p class="mt-2">在 AI 对话中选择“制定计划”来生成一个吧！</p>
      </div>
    </div>

    <div v-else class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-6">
      <div class="min-h-0 overflow-y-auto pr-2 order-2 lg:order-1">
        <!-- Mobile History Button -->
        <div class="lg:hidden mb-4">
          <button 
            @click="showMobileHistory = true"
            class="w-full flex items-center justify-center gap-2 py-3 rounded-2xl border-2 border-primary/20 text-primary bg-primary/5 font-bold text-sm shadow-sm active:scale-95 transition-all"
          >
            <History class="w-5 h-5" />
            历史计划管理 ({{ historyPlans.length }})
          </button>
        </div>

        <div v-if="selectedHistoryId !== null" class="mb-3 inline-flex items-center rounded-full border border-primary/50 bg-primary/10 px-3 py-1 text-xs text-primary">
          正在查看：历史计划 #{{ selectedHistoryId }}
        </div>
        <PlanTable :data="displayPlan || plan" />
      </div>

      <!-- Desktop Sidebar -->
      <aside class="hidden lg:flex flex-col rounded-2xl border border-border bg-card p-4 min-h-0 h-auto order-2">
        <div class="text-sm font-semibold text-foreground mb-3">计划管理</div>
        
        <button
          @click="viewCurrentPlan"
          :class="[
            'mb-3 px-3 py-2 rounded-xl border text-sm text-left transition-colors',
            selectedHistoryId === null ? 'border-primary bg-primary/15 text-primary' : 'border-border text-muted-foreground hover:text-foreground'
          ]"
        >
          查看当前计划
        </button>

        <div v-if="historyLoading" class="text-sm text-muted-foreground py-4 text-center">加载中...</div>
        <div v-else-if="!historyPlans.length" class="text-sm text-muted-foreground py-4 text-center">暂无历史计划</div>

        <div v-else class="space-y-2 overflow-y-auto pr-1">
          <div
            v-for="item in historyPlans"
            :key="item.id"
            :class="[
              'rounded-xl border p-3 bg-muted/30',
              selectedHistoryId === item.id ? 'border-primary/60' : 'border-border'
            ]"
          >
            <button
              @click="handleViewHistory(item)"
              class="text-left w-full"
            >
              <div class="text-xs text-muted-foreground">历史计划 #{{ item.id }}</div>
              <div class="text-xs text-muted-foreground mt-1">{{ formatTime(item.created_at) }}</div>
            </button>

            <div class="mt-2 flex gap-2">
              <button
                @click="replaceFromHistory(item)"
                :disabled="replacingId === item.id"
                class="flex-1 px-2 py-1.5 rounded-md text-xs border border-primary/50 text-primary hover:bg-primary/10 disabled:opacity-50"
              >
                {{ replacingId === item.id ? '替换中...' : '替换为当前' }}
              </button>
              <button
                @click="removeHistory(item)"
                :disabled="deletingId === item.id"
                class="px-2 py-1.5 rounded-md text-xs border border-red-500/50 text-red-300 hover:bg-red-500/10 disabled:opacity-50"
              >
                {{ deletingId === item.id ? '删除中...' : '删除' }}
              </button>
            </div>
          </div>
        </div>
      </aside>

      <!-- Mobile History Modal -->
      <Teleport to="body">
        <Transition
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="opacity-0 translate-y-full"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 translate-y-full"
        >
          <div v-if="showMobileHistory" class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center p-4">
            <div @click="showMobileHistory = false" class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>
            
            <div class="relative w-full max-w-lg bg-card border border-border rounded-t-3xl sm:rounded-3xl shadow-2xl flex flex-col max-h-[80vh] overflow-hidden animate-in slide-in-from-bottom duration-300">
              <div class="px-6 py-4 border-b border-border flex items-center justify-between bg-card/80 backdrop-blur-sm sticky top-0 z-10">
                <div class="flex items-center gap-2">
                  <History class="w-5 h-5 text-primary" />
                  <span class="font-bold">历史计划管理</span>
                </div>
                <button @click="showMobileHistory = false" class="p-2 hover:bg-muted rounded-xl transition-colors">
                  <X class="w-6 h-6" />
                </button>
              </div>

              <div class="p-6 overflow-y-auto space-y-4">
                <button
                  @click="handleViewCurrent"
                  :class="[
                    'w-full px-4 py-3 rounded-2xl border text-sm text-left transition-all',
                    selectedHistoryId === null ? 'border-primary bg-primary/10 text-primary font-bold' : 'border-border text-muted-foreground bg-muted/30'
                  ]"
                >
                  <div class="flex items-center justify-between">
                    <span>查看当前计划</span>
                    <div v-if="selectedHistoryId === null" class="w-2 h-2 rounded-full bg-primary animate-pulse"></div>
                  </div>
                </button>

                <div class="text-xs font-bold text-muted-foreground uppercase tracking-widest px-1">历史计划记录</div>

                <div v-if="historyLoading" class="flex flex-col items-center py-10 gap-3 opacity-50">
                  <div class="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
                  <div class="text-sm">正在加载历史...</div>
                </div>
                <div v-else-if="!historyPlans.length" class="text-center py-10 bg-muted/20 rounded-2xl border border-dashed border-border text-muted-foreground italic">
                  暂无历史记录
                </div>
                <div v-else class="space-y-3">
                  <div
                    v-for="item in historyPlans"
                    :key="item.id"
                    :class="[
                      'rounded-2xl border p-4 transition-all',
                      selectedHistoryId === item.id ? 'border-primary bg-primary/5 ring-1 ring-primary/20' : 'border-border bg-muted/20'
                    ]"
                  >
                    <button
                      @click="handleViewHistory(item)"
                      class="text-left w-full"
                    >
                      <div class="flex items-center justify-between mb-1">
                        <span class="text-sm font-bold" :class="selectedHistoryId === item.id ? 'text-primary' : 'text-foreground'">
                          历史计划 #{{ item.id }}
                        </span>
                        <span class="text-[10px] text-muted-foreground">{{ formatTime(item.created_at) }}</span>
                      </div>
                    </button>

                    <div class="mt-4 flex gap-3">
                      <button
                        @click="replaceFromHistory(item)"
                        :disabled="replacingId === item.id"
                        class="flex-1 py-2.5 rounded-xl text-xs font-bold bg-primary/10 text-primary border border-primary/20 hover:bg-primary/20 transition-all disabled:opacity-50"
                      >
                        {{ replacingId === item.id ? '替换中...' : '应用此计划' }}
                      </button>
                      <button
                        @click="removeHistory(item)"
                        :disabled="deletingId === item.id"
                        class="px-4 py-2.5 rounded-xl text-xs font-bold bg-red-500/10 text-red-500 border border-red-500/20 hover:bg-red-500/20 transition-all disabled:opacity-50"
                      >
                        <Trash2 v-if="deletingId !== item.id" class="w-4 h-4" />
                        <span v-else>...</span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { LayoutDashboard, History, X, Trash2 } from 'lucide-vue-next'
import PlanTable from '../components/render/PlanTable.vue'
import { deletePlanHistory, getPlanHistory, getProfile, replacePlanFromHistory } from '../api'
import { useToastStore } from '../api/toast'
import { useDialogStore } from '../api/dialog'

const plan = ref(null)
const displayPlan = ref(null)
const loading = ref(true)
const userId = localStorage.getItem('user_id')
const historyPlans = ref([])
const historyLoading = ref(false)
const selectedHistoryId = ref(null)
const deletingId = ref(null)
const replacingId = ref(null)
const showMobileHistory = ref(false)
const toast = useToastStore()
const dialog = useDialogStore()

function normalizePlanShape(raw) {
  if (!raw || typeof raw !== 'object') return null
  if (raw.new_plan && typeof raw.new_plan === 'object') {
    return normalizePlanShape(raw.new_plan)
  }
  if (raw.plan && typeof raw.plan === 'object') {
    return normalizePlanShape(raw.plan)
  }
  if (Array.isArray(raw.months)) {
    return raw
  }
  return null
}

function parsePlan(raw) {
  if (!raw) return null
  if (typeof raw === 'string') {
    try {
      return normalizePlanShape(JSON.parse(raw))
    } catch {
      return null
    }
  }
  if (typeof raw === 'object') {
    return normalizePlanShape(raw)
  }
  return null
}

function formatTime(value) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return date.toLocaleString()
}

function viewCurrentPlan() {
  selectedHistoryId.value = null
  displayPlan.value = plan.value
}

function handleViewCurrent() {
  viewCurrentPlan()
  showMobileHistory.value = false
}

function viewHistoryPlan(item) {
  if (!item?.plan) {
    toast.add({ type: 'error', title: '渲染失败', message: '该历史计划数据无效' })
    return
  }
  selectedHistoryId.value = item.id
  displayPlan.value = item.plan
}

function handleViewHistory(item) {
  viewHistoryPlan(item)
  showMobileHistory.value = false
}

async function fetchPlan() {
  if (!userId) return
  loading.value = true
  try {
    const res = await getProfile(userId)
    if (res.data) {
      localStorage.setItem('profile_cache', JSON.stringify(res.data))
    }
    const rawPlan = res.data?.latest_plan_json || res.data?.plan_data
    plan.value = parsePlan(rawPlan)
    displayPlan.value = plan.value
  } catch (e) {
    console.error('Failed to fetch plan', e)
    toast.add({ type: 'error', title: '读取失败', message: e.response?.data?.error || e.message })
  } finally {
    loading.value = false
  }
}

async function fetchHistoryPlans() {
  if (!userId) return
  historyLoading.value = true
  try {
    const res = await getPlanHistory(userId)
    const items = Array.isArray(res.data?.items) ? res.data.items : []
    historyPlans.value = items
      .map(item => ({
        id: item.id,
        created_at: item.created_at,
        plan: parsePlan(item.plan)
      }))
      .filter(item => item.plan)
  } catch (e) {
    console.error('Failed to fetch plan history', e)
    toast.add({ type: 'error', title: '读取失败', message: e.response?.data?.error || e.message })
  } finally {
    historyLoading.value = false
  }
}

async function removeHistory(item) {
  if (!item?.id) return
  const ok = await dialog.confirm('删除历史计划', '确认删除该历史计划吗？此操作无法撤销。')
  if (!ok) return
  deletingId.value = item.id
  try {
    await deletePlanHistory(userId, item.id)
    toast.add({ type: 'success', message: '历史计划已删除' })
    if (selectedHistoryId.value === item.id) {
      viewCurrentPlan()
    }
    await fetchHistoryPlans()
  } catch (e) {
    toast.add({ type: 'error', title: '删除失败', message: e.response?.data?.error || e.message })
  } finally {
    deletingId.value = null
  }
}

async function replaceFromHistory(item) {
  if (!item?.id) return
  const ok = await dialog.confirm('替换计划', '确认使用该历史计划替换当前计划吗？当前计划将被移动到历史记录。')
  if (!ok) return
  replacingId.value = item.id
  try {
    await replacePlanFromHistory({ user_id: Number(userId), record_id: item.id })
    toast.add({ type: 'success', message: '已替换为历史计划' })
    await Promise.all([fetchPlan(), fetchHistoryPlans()])
    viewCurrentPlan()
  } catch (e) {
    toast.add({ type: 'error', title: '替换失败', message: e.response?.data?.error || e.message })
  } finally {
    replacingId.value = null
  }
}

onMounted(() => {
  fetchPlan()
  fetchHistoryPlans()
})
</script>
