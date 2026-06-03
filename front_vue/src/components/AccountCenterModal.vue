<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-2 md:p-4 bg-background/80 backdrop-blur-md">
    <div class="bg-card border border-border w-full max-w-xl rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[95vh]">
      <div class="px-6 md:px-8 py-4 md:py-6 border-b border-border flex items-center justify-between">
        <div>
          <h2 class="text-xl md:text-2xl font-bold text-foreground">个人中心</h2>
          <p class="text-xs md:text-sm text-muted-foreground mt-1">修改名称、头像、密码与账号操作</p>
        </div>
        <button @click="$emit('close')" class="p-2 hover:bg-muted rounded-xl transition-colors">
          <X class="w-6 h-6" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-6 md:px-8 py-6 md:py-8 space-y-8">
        <div class="flex items-center gap-5">
          <div class="w-16 h-16 md:w-20 md:h-20 rounded-2xl overflow-hidden border border-border bg-muted flex items-center justify-center">
            <img v-if="avatarPreview" :src="avatarPreview" class="w-full h-full object-cover" />
            <User v-else class="w-8 h-8 md:w-10 md:h-10 text-muted-foreground" />
          </div>
          <div class="flex-1 space-y-2">
            <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest">账号特权</label>
            <div class="flex items-center gap-2">
              <button
                @click="$emit('open-membership')"
                class="px-4 py-2 rounded-xl bg-orange-500 text-white hover:bg-orange-600 transition-all text-sm font-bold flex items-center gap-2"
              >
                <Crown class="w-4 h-4" />
                会员中心
              </button>
              <span
                class="text-xs px-2.5 py-1 rounded-lg border"
                :class="isProActive ? 'text-orange-600 border-orange-200 bg-orange-50' : 'text-muted-foreground border-border bg-muted'"
              >
                {{ isProActive ? `PRO剩余${membershipRemainingDays}天` : '当前 FREE' }}
              </span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-5">
          <div class="w-16 h-16 md:w-20 md:h-20 flex items-center justify-center">
            <!-- Space for avatar if needed, or keep alignment -->
          </div>
          <div class="flex-1 space-y-2">
            <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest">头像（自动压缩为150x150）</label>
            <div>
              <button
                @click="avatarInput?.click()"
                class="px-4 py-2 rounded-xl bg-muted border border-border hover:bg-muted/80 transition-all text-sm font-medium"
              >
                上传头像
              </button>
            </div>
            <input ref="avatarInput" type="file" accept="image/*" class="hidden" @change="onAvatarChange" />
          </div>
        </div>

        <div class="space-y-2">
          <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest">名称</label>
          <input
            v-model.trim="username"
            type="text"
            class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 focus:ring-0 transition-all"
            placeholder="输入新的名称"
          />
        </div>

        <div class="pt-2 border-t border-border space-y-4">
          <h3 class="font-semibold">重设密码</h3>
          <input
            v-model="oldPassword"
            type="password"
            class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 focus:ring-0 transition-all"
            placeholder="旧密码"
          />
          <input
            v-model="newPassword"
            type="password"
            class="w-full bg-muted border border-border rounded-xl px-4 py-3 focus:border-primary/50 focus:ring-0 transition-all"
            placeholder="新密码（至少6位）"
          />
          <button
            @click="handleResetPassword"
            :disabled="passwordSaving"
            class="w-full px-6 py-3 rounded-xl border border-border hover:bg-muted transition-all disabled:opacity-50"
          >
            {{ passwordSaving ? '重设中...' : '确认重设密码' }}
          </button>
        </div>
      </div>

      <div class="px-6 md:px-8 py-4 md:py-6 border-t border-border flex flex-wrap md:flex-nowrap gap-3">
        <button
          @click="$emit('logout')"
          class="flex-1 md:flex-none px-5 py-3 rounded-2xl border border-red-500/40 text-red-400 hover:bg-red-500/10 transition-all font-semibold text-sm"
        >
          退出登录
        </button>
        <button
          @click="$emit('close')"
          class="flex-1 px-6 py-3 rounded-2xl border border-border hover:bg-muted transition-all font-semibold text-sm"
        >
          关闭
        </button>
        <button
          @click="handleSave"
          :disabled="saving"
          class="w-full md:flex-1 px-6 py-3 rounded-2xl bg-primary text-white font-semibold shadow-lg shadow-primary/20 hover:bg-primary/90 disabled:opacity-50 text-sm"
        >
          {{ saving ? '保存中...' : '保存资料' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { X, User, Crown } from 'lucide-vue-next'
import { getAccount, updateAccount, resetPassword, getMembershipInfo } from '../api'
import { useToastStore } from '../api/toast'

const emit = defineEmits(['close', 'updated', 'logout', 'open-membership'])
const userId = Number(localStorage.getItem('user_id') || 0)
const toast = useToastStore()

const avatarInput = ref(null)
const avatarPreview = ref('')
const username = ref('')
const avatarBase64 = ref('')
const oldPassword = ref('')
const newPassword = ref('')
const saving = ref(false)
const passwordSaving = ref(false)
const membershipType = ref('FREE')
const membershipRemainingDays = ref(0)
const isProActive = ref(false)

async function loadAccount() {
  if (!userId) return
  try {
    const res = await getAccount(userId)
    username.value = res.data?.username || ''
    avatarPreview.value = res.data?.avatar_base64 || ''

    const membershipRes = await getMembershipInfo(userId)
    const membershipPayload = membershipRes?.data || membershipRes || {}
    membershipType.value = String(membershipPayload?.membership_type || 'FREE')
    const expireAt = membershipPayload?.membership_expire_at ? new Date(membershipPayload.membership_expire_at) : null
    if (expireAt instanceof Date && !Number.isNaN(expireAt.getTime())) {
      const diffMs = expireAt.getTime() - Date.now()
      membershipRemainingDays.value = diffMs > 0 ? Math.ceil(diffMs / (24 * 60 * 60 * 1000)) : 0
    } else {
      membershipRemainingDays.value = 0
    }
    isProActive.value = membershipType.value !== 'FREE' && membershipRemainingDays.value > 0
  } catch (e) {
    toast.add({ type: 'error', title: '读取失败', message: e.response?.data?.error || e.message })
  }
}

function onAvatarChange(e) {
  const file = e.target.files?.[0]
  if (!file || !file.type.startsWith('image/')) return

  const reader = new FileReader()
  reader.onload = () => {
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = 150
      canvas.height = 150
      const ctx = canvas.getContext('2d')
      if (!ctx) return
      ctx.clearRect(0, 0, 150, 150)
      ctx.drawImage(img, 0, 0, 150, 150)
      const dataUrl = canvas.toDataURL('image/png')
      avatarPreview.value = dataUrl
      avatarBase64.value = dataUrl
    }
    img.src = String(reader.result || '')
  }
  reader.readAsDataURL(file)
}

async function handleSave() {
  if (!userId) return
  const payload = { user_id: userId }
  if (username.value) payload.username = username.value
  if (avatarBase64.value) payload.avatar_base64 = avatarBase64.value

  saving.value = true
  try {
    await updateAccount(payload)
    toast.add({ type: 'success', message: '个人中心保存成功' })
    avatarBase64.value = ''
    emit('updated')
  } catch (e) {
    toast.add({ type: 'error', title: '保存失败', message: e.response?.data?.error || e.message })
  } finally {
    saving.value = false
  }
}

async function handleResetPassword() {
  if (!oldPassword.value || !newPassword.value) {
    toast.add({ type: 'warning', message: '请输入旧密码和新密码' })
    return
  }
  passwordSaving.value = true
  try {
    await resetPassword({ user_id: userId, old_password: oldPassword.value, new_password: newPassword.value })
    toast.add({ type: 'success', message: '密码重设成功' })
    oldPassword.value = ''
    newPassword.value = ''
  } catch (e) {
    toast.add({ type: 'error', title: '重设失败', message: e.response?.data?.error || e.message })
  } finally {
    passwordSaving.value = false
  }
}

onMounted(() => {
  loadAccount()
})
</script>
