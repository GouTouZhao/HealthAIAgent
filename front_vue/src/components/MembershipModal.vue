<script setup>
import { ref, onMounted, computed } from 'vue'
import { 
  X, Crown, Check, AlertCircle, TrendingUp, 
  Wallet, Gift, Zap, ShieldCheck, Clock
} from 'lucide-vue-next'
import { getMembershipInfo, purchaseMembership, withdrawBalance } from '../api'
import { useDialogStore } from '../api/dialog'

const props = defineProps({
  show: Boolean,
  userId: [Number, String]
})

const emit = defineEmits(['close', 'refresh'])

const dialog = useDialogStore()
const loading = ref(false)
const info = ref({
  membership_type: 'FREE',
  membership_expire_at: null,
  balance: 0,
  free_trial_used: false,
  full_weeks_count: 0
})

const currentPlanId = computed(() => {
  const membershipType = String(info.value?.membership_type || 'FREE')
  const expireAtText = info.value?.membership_expire_at
  const expireAt = expireAtText ? new Date(expireAtText) : null
  const trialActive = (membershipType === 'TRY') || (
    membershipType === 'PRO_ANNUAL' &&
    info.value?.free_trial_used &&
    expireAt instanceof Date &&
    !Number.isNaN(expireAt.getTime()) &&
    expireAt.getTime() > Date.now()
  )

  if (trialActive) return 'TRY'
  if (membershipType === 'PRO_HALF') return 'PRO_HALF'
  if (membershipType === 'PRO_ANNUAL') return 'PRO_ANNUAL'
  return 'FREE'
})

const membershipLabel = computed(() => {
  if (currentPlanId.value === 'TRY') return '试用版（PRO权益）'
  if (currentPlanId.value === 'PRO_ANNUAL') return '年度PRO'
  if (currentPlanId.value === 'PRO_HALF') return '半年PRO'
  return '基础版'
})

const normalizedUserId = computed(() => Number(props.userId || 0))

const selectedPlan = ref('PRO_ANNUAL')

const isPaidPlan = (planId) => planId === 'PRO_HALF' || planId === 'PRO_ANNUAL'

const selectPlan = (planId) => {
  selectedPlan.value = planId
}

const getPlanButtonLabel = (plan) => {
  if (plan.id === currentPlanId.value) return '当前方案'
  if (plan.id === 'FREE' && currentPlanId.value === 'TRY') return 'PRO套餐生效'
  if (isPaidPlan(plan.id)) return '立即开通'
  if (plan.id === 'TRY' && info.value.free_trial_used) return '已领取'
  return '立即开通'
}

const isPlanButtonDisabled = (plan) => {
  if (loading.value) return true
  if (plan.id === 'FREE') return true
  if (plan.id === currentPlanId.value) return true
  if (isPaidPlan(plan.id)) return true
  if (plan.id === 'TRY' && info.value.free_trial_used) return true
  return false
}

const isPermanentHighlightedPlan = (planId) => planId === currentPlanId.value

const isActionDisabledStyle = (plan) => plan.id === 'FREE'

const plans = [
  {
    id: 'FREE',
    name: '基础版 (FREE)',
    price: 0,
    desc: '核心体验',
    features: ['食物识别', '专家咨询', '基础功能'],
    disabled: ['制定计划', '修改计划', '今日任务', '记忆功能'],
    bg: 'bg-gray-100 dark:bg-zinc-800'
  },
  {
    id: 'TRY',
    name: '试用版 (TRY)',
    price: 0,
    desc: '新用户专享',
    features: ['一个月PRO体验', '解锁所有功能', '限购一次'],
    bg: 'bg-orange-50 dark:bg-orange-950/20'
  },
  {
    id: 'PRO_HALF',
    name: '半年卡 (PRO)',
    price: 48,
    desc: '性价比之选',
    features: ['解锁所有功能', '商城95折优惠', '全勤返现'],
    bg: 'bg-orange-100 dark:bg-orange-900/30'
  },
  {
    id: 'PRO_ANNUAL',
    name: '年卡 (PRO)',
    price: 88,
    desc: '极致体验',
    features: ['解锁所有功能', '商城95折优惠', '高额全勤返现'],
    bg: 'bg-orange-500'
  }
]

const currentPlan = computed(() => plans.find(p => p.id === selectedPlan.value) || plans[3])

const fetchInfo = async () => {
  if (!normalizedUserId.value) return
  loading.value = true
  try {
    const res = await getMembershipInfo(normalizedUserId.value)
    const payload = res?.data || res || {}
    info.value = {
      membership_type: 'FREE',
      membership_expire_at: null,
      balance: 0,
      free_trial_used: false,
      full_weeks_count: 0,
      ...payload
    }
    selectedPlan.value = currentPlanId.value
  } catch (err) {
    console.error('Fetch membership info failed:', err)
  } finally {
    loading.value = false
  }
}

const handlePurchase = async (planId) => {
  if (!normalizedUserId.value) {
    await dialog.alert('无法开通', '未检测到用户登录信息，请重新登录后再试。')
    return
  }
  if (planId === 'FREE' || isPaidPlan(planId)) return

  loading.value = true
  try {
    const res = await purchaseMembership({
      user_id: normalizedUserId.value,
      plan: planId
    })
    await dialog.alert('购买成功', '您的会员权益已即时生效，感谢支持！')
    await fetchInfo()
    emit('refresh')
  } catch (err) {
    dialog.alert('购买失败', err.response?.data?.error || '处理订单时出现问题，请稍后再试。')
  } finally {
    loading.value = false
  }
}

const handleWithdraw = async () => {
  if (!normalizedUserId.value) {
    await dialog.alert('无法提现', '未检测到用户登录信息，请重新登录后再试。')
    return
  }
  if (!info.value.balance) return
  const amount = await dialog.prompt('余额提现', `当前可提现金额: ¥${info.value.balance.toFixed(2)}。请输入提现金额:`, info.value.balance.toString())
  if (!amount || isNaN(amount) || parseFloat(amount) <= 0) return
  
  try {
    await withdrawBalance({
      user_id: normalizedUserId.value,
      amount: parseFloat(amount)
    })
    await dialog.alert('申请已提交', '您的提现申请已在处理中，预计1-3个工作日到账。')
    await fetchInfo()
  } catch (err) {
    dialog.alert('提现失败', err.response?.data?.error || '提现过程中出现问题，请联系客服。')
  }
}

const cashbackStatus = computed(() => {
  const weeks = info.value.full_weeks_count || 0
  const effectivePlanId = currentPlanId.value
  const isAnnual = effectivePlanId === 'PRO_ANNUAL'
  const isHalf = effectivePlanId === 'PRO_HALF'
  
  if (isAnnual) {
    if (weeks >= 48) return { label: '50% 返还已解锁', progress: 100, next: '全额达成' }
    if (weeks >= 30) return { label: '30% 已解锁', next: '还需 18 周解锁 50%', progress: (weeks/48)*100 }
    if (weeks >= 10) return { label: '10% 已解锁', next: '还需 20 周解锁 30%', progress: (weeks/30)*100 }
    return { label: '全勤挑战中', next: `还需 ${10 - weeks} 周解锁 10%`, progress: (weeks/10)*100 }
  } else if (isHalf) {
    // Half weeks: 5, 15, 24
    if (weeks >= 24) return { label: '50% 返还已解锁', progress: 100, next: '全额达成' }
    if (weeks >= 15) return { label: '30% 已解锁', next: '还需 9 周解锁 50%', progress: (weeks/24)*100 }
    if (weeks >= 5) return { label: '10% 已解锁', next: '还需 10 周解锁 30%', progress: (weeks/15)*100 }
    return { label: '全勤挑战中', next: `还需 ${5 - weeks} 周解锁 10%`, progress: (weeks/5)*100 }
  }
  return null
})

onMounted(fetchInfo)
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-white/70 backdrop-blur-sm">
    <div class="bg-white w-full max-w-4xl rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh] border border-zinc-200">
      <!-- Header -->
      <div class="p-6 flex justify-between items-center border-b border-zinc-200">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-orange-500 rounded-xl">
            <Crown class="w-6 h-6 text-white" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-zinc-900">会员中心</h2>
            <p class="text-sm text-zinc-500">升级会员，解锁极致健康体验</p>
          </div>
        </div>
        <button @click="emit('close')" class="p-2 hover:bg-zinc-100 rounded-full transition-colors">
          <X class="w-6 h-6 text-zinc-500" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-8">
        <!-- Current Status -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="p-4 rounded-2xl bg-zinc-50 border border-zinc-200 flex items-center gap-4">
            <div class="p-3 bg-zinc-200 rounded-xl">
              <Zap class="w-5 h-5 text-orange-500" />
            </div>
            <div>
              <div class="text-xs text-zinc-500 uppercase font-semibold">当前身份</div>
              <div class="font-bold text-zinc-900">{{ membershipLabel }}</div>
            </div>
          </div>
          <div class="p-4 rounded-2xl bg-zinc-50 border border-zinc-200 flex items-center gap-4">
            <div class="p-3 bg-zinc-200 rounded-xl">
              <Wallet class="w-5 h-5 text-green-500" />
            </div>
            <div class="flex-1">
              <div class="text-xs text-zinc-500 uppercase font-semibold">账户余额</div>
              <div class="flex items-center justify-between">
                <span class="font-bold text-zinc-900">¥{{ info.balance?.toFixed(2) }}</span>
                <button @click="handleWithdraw" class="text-xs font-bold text-orange-500 hover:underline" :disabled="!info.balance">提现</button>
              </div>
            </div>
          </div>
          <div class="p-4 rounded-2xl bg-zinc-50 border border-zinc-200 flex items-center gap-4">
            <div class="p-3 bg-zinc-200 rounded-xl">
              <TrendingUp class="w-5 h-5 text-blue-500" />
            </div>
            <div>
              <div class="text-xs text-zinc-500 uppercase font-semibold">全勤周数</div>
              <div class="font-bold text-zinc-900">{{ info.full_weeks_count }} 周</div>
            </div>
          </div>
        </div>

        <!-- Cashback Banner (If Member) -->
        <div v-if="cashbackStatus" class="p-6 rounded-3xl bg-gradient-to-r from-orange-500 to-rose-500 text-white shadow-xl relative overflow-hidden">
          <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div class="space-y-2">
              <div class="flex items-center gap-2">
                <Gift class="w-5 h-5" />
                <span class="font-bold uppercase tracking-wider text-sm">全勤返现计划</span>
              </div>
              <h3 class="text-2xl font-black">{{ cashbackStatus.label }}</h3>
              <p class="text-orange-100 text-sm font-medium">{{ cashbackStatus.next }}</p>
            </div>
            <div class="flex-1 max-w-xs">
              <div class="flex justify-between text-xs font-bold mb-2">
                <span>当前进度</span>
                <span>{{ Math.round(cashbackStatus.progress) }}%</span>
              </div>
              <div class="h-3 bg-white/20 rounded-full overflow-hidden">
                <div class="h-full bg-white transition-all duration-1000" :style="{ width: `${cashbackStatus.progress}%` }"></div>
              </div>
            </div>
          </div>
          <!-- Decorative Background -->
          <Crown class="absolute -right-10 -bottom-10 w-48 h-48 text-white/10 rotate-12" />
        </div>

        <!-- Plans Cards -->
        <div>
          <h3 class="text-lg font-bold mb-6 text-zinc-900 flex items-center gap-2">
            选择您的成长方案
          </h3>
          <!-- Horizontal scroll on mobile, grid on desktop -->
          <div class="flex overflow-x-auto pb-4 gap-4 md:grid md:grid-cols-4 md:overflow-x-visible no-scrollbar">
            <div v-for="plan in plans" :key="plan.id" 
              @click="selectPlan(plan.id)"
              class="relative flex-shrink-0 w-[260px] md:w-auto flex flex-col p-6 rounded-3xl border-2 transition-all cursor-pointer group"
              :class="[
                selectedPlan === plan.id 
                  ? 'border-orange-500 shadow-xl bg-orange-50/80' 
                  : 'border-zinc-200 hover:border-orange-300 bg-zinc-50/50'
              ]">

              <div class="mb-4">
                <h4 class="font-bold text-zinc-900">{{ plan.name }}</h4>
                <p class="text-xs text-zinc-500">{{ plan.desc }}</p>
              </div>

              <div class="mb-6">
                <div class="flex items-baseline gap-1">
                  <span class="text-2xl font-black text-zinc-900">¥{{ plan.price }}</span>
                  <span v-if="plan.price > 0" class="text-zinc-500 text-xs">/ 期</span>
                </div>
              </div>

              <button 
                @click.stop="selectPlan(plan.id); handlePurchase(plan.id)"
                class="w-full py-2.5 rounded-xl font-bold text-sm transition-all mb-6"
                :disabled="isPlanButtonDisabled(plan)"
                :class="[
                  (selectedPlan === plan.id || isPermanentHighlightedPlan(plan.id))
                    ? 'bg-orange-500 text-white shadow-lg shadow-orange-500/30'
                    : 'bg-zinc-100 text-zinc-500 group-hover:bg-orange-500 group-hover:text-white group-hover:shadow-lg group-hover:shadow-orange-500/30',
                  isActionDisabledStyle(plan) ? 'opacity-50 cursor-not-allowed' : ''
                ]">
                {{ getPlanButtonLabel(plan) }}
              </button>

              <div class="space-y-3 mt-auto">
                <div v-for="feat in plan.features" :key="feat" class="flex items-center gap-2 text-xs font-medium text-zinc-700">
                  <Check class="w-3.5 h-3.5 text-green-500 shrink-0" />
                  <span>{{ feat }}</span>
                </div>
                <div v-for="dis in plan.disabled" :key="dis" class="flex items-center gap-2 text-xs font-medium text-zinc-400">
                  <X class="w-3.5 h-3.5 text-zinc-300 shrink-0" />
                  <span class="line-through">{{ dis }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Plan Description -->
        <div class="p-6 rounded-3xl bg-zinc-50 border border-zinc-200">
          <h4 class="font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <AlertCircle class="w-4 h-4 text-orange-500" />
            方案详情
          </h4>
          <div class="space-y-4 text-sm text-zinc-600 leading-relaxed">
            <p v-if="selectedPlan === 'FREE'">
              <strong>基础版：</strong> 适合轻度使用。可用 AI 食物识别与健康咨询；训练计划、计划调整、今日任务与记忆能力保持关闭状态。
            </p>
            <p v-if="selectedPlan === 'TRY'">
              <strong>试用版：</strong> 新用户可免费领取 1 个月完整 PRO 权益（仅可领取一次）。点击“立即开通”后立即生效。
            </p>
            <p v-if="selectedPlan === 'PRO_HALF'">
              <strong>半年 PRO：</strong> 当前版本暂未开放在线购买。此卡将提供 180 天完整 PRO 权益、商城折扣及全勤返现活动，后续可直接开通。
            </p>
            <p v-if="selectedPlan === 'PRO_ANNUAL'">
              <strong>年度 PRO：</strong> 此卡将提供 365 天完整 PRO 权益、商城折扣及更高档位全勤返现活动。
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Hide scrollbar for Chrome, Safari and Opera */
.no-scrollbar::-webkit-scrollbar {
  display: none;
}

/* Hide scrollbar for IE, Edge and Firefox */
.no-scrollbar {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}

/* Custom Scrollbar for better UX (Desktop main modal) */
::-webkit-scrollbar {
  width: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 10px;
}
</style>
