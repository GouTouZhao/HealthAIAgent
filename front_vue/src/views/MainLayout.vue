<template>
  <div :class="['flex h-screen bg-background text-foreground font-sans overflow-hidden transition-colors duration-300', isDark ? 'dark' : '']">
    <!-- Mobile Header -->
    <header class="lg:hidden fixed top-0 left-0 right-0 h-14 border-b border-border bg-background flex items-center justify-between px-4 z-50">
      <button @click="isSidebarOpen = true" class="p-2 -ml-2 hover:bg-muted rounded-lg transition-colors">
        <Menu class="w-6 h-6" />
      </button>
      <div class="font-bold text-primary">{{ currentViewName }}</div>
      <div class="w-10 h-10"></div> <!-- Spacer -->
    </header>

    <!-- Sidebar Overlay -->
    <div 
      v-if="isSidebarOpen" 
      @click="isSidebarOpen = false"
      class="lg:hidden fixed inset-0 bg-black/50 z-[60] backdrop-blur-sm"
    ></div>

    <!-- Sidebar -->
    <aside 
      :class="[
        'w-72 flex-shrink-0 flex flex-col border-r border-border bg-[var(--sidebar-bg)] transition-all duration-300 z-[70]',
        'lg:relative lg:translate-x-0 fixed inset-y-0 left-0',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      ]"
    >
      <!-- New Chat Button -->
      <div class="p-4 flex items-center justify-between gap-2">
        <button 
          @click="startNewChat"
          class="flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-xl border border-border hover:bg-muted transition-all duration-200 group"
        >
          <Plus class="w-5 h-5 group-hover:text-primary transition-colors" />
          <span class="font-medium">新建对话</span>
        </button>
        <button 
          @click="toggleTheme"
          class="p-3 rounded-xl border border-border hover:bg-muted transition-all duration-200 text-muted-foreground hover:text-primary"
          title="切换主题"
        >
          <Sun v-if="isDark" class="w-5 h-5" />
          <Moon v-else class="w-5 h-5" />
        </button>
      </div>

      <!-- Main Menu -->
      <nav class="flex flex-col gap-1 px-4 py-2">
        <button 
          v-for="item in menuItems" 
          :key="item.id"
          @click="handleMenuClick(item.id)"
          :class="[
            'flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200',
            currentView === item.id ? 'bg-primary text-primary-foreground' : 'text-foreground hover:bg-muted/50'
          ]"
        >
          <component :is="item.icon" class="w-5 h-5" />
          <span class="font-medium">{{ item.name }}</span>
        </button>
      </nav>

      <!-- Chat History -->
      <div class="flex-1 overflow-y-auto px-4 py-4 scroll-smooth">
        <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3 px-3">
          最近对话
        </div>
        <div class="flex flex-col gap-1">
          <div
            v-for="chat in chatHistory" 
            :key="chat.id"
            @click="selectChat(chat)"
            @keydown.enter.prevent="selectChat(chat)"
            tabindex="0"
            role="button"
            :class="[
              'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 text-left truncate group',
              selectedChatId === chat.id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
            ]"
          >
            <MessageSquare class="w-4 h-4 flex-shrink-0" />
            <span class="truncate flex-1">{{ chat.title }}</span>
            <button
              @click.stop="deleteChat(chat.id)"
              class="opacity-0 group-hover:opacity-100 transition-opacity text-xs px-2 py-0.5 rounded-md hover:bg-red-500/20 hover:text-red-400"
            >
              删除
            </button>
          </div>
        </div>
      </div>

      <!-- User Section -->
      <div class="p-4 border-t border-border relative">
        <button 
          @click="showAccountCenter = true"
          class="w-full flex items-center gap-3 p-3 rounded-xl hover:bg-muted transition-all duration-200 group"
        >
          <div class="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-primary group-hover:bg-primary/30 transition-colors overflow-hidden">
            <img v-if="avatarBase64" :src="avatarBase64" class="w-full h-full object-cover" />
            <User v-else class="w-6 h-6" />
          </div>
          <div class="flex-1 text-left">
            <div class="font-medium truncate">{{ userName || '用户' }}</div>
            <div class="text-xs text-muted-foreground">个人中心</div>
          </div>
          <Settings class="w-5 h-5 text-muted-foreground group-hover:text-foreground transition-all" />
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col relative overflow-hidden bg-background lg:pt-0 pt-14">
      <div class="flex-1 overflow-y-auto">
        <component 
          :is="activeComp" 
          :chat-id="selectedChatId"
          :chat-history="chatHistory"
          :user-avatar="avatarBase64"
          :membership-type="membershipType"
          @open-profile="showProfile = true"
          @open-membership="showMembership = true"
          @ensure-chat-created="ensureChatCreated"
          @rename-chat-from-first-message="renameChatFromFirstMessage"
          @touch-chat="touchChat"
        />
      </div>
    </main>

    <!-- Profile Modal -->
    <ProfileModal 
      v-if="showProfile" 
      :membership-type="membershipType"
      @close="showProfile = false" 
      @open-membership="showMembership = true"
      @updated="handleProfileUpdated"
    />

    <AccountCenterModal
      v-if="showAccountCenter"
      @close="showAccountCenter = false"
      @open-membership="showMembership = true"
      @updated="loadProfile"
      @logout="handleLogout"
    />

    <MembershipModal
      v-if="showMembership"
      :show="showMembership"
      :user-id="userId"
      @close="showMembership = false"
      @refresh="loadMembership"
    />

    <CustomDialog />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Plus, 
  LayoutDashboard, 
  MessageSquare, 
  Calendar, 
  LineChart, 
  User, 
  Settings,
  Menu,
  Sun,
  Moon,
  X,
  ShoppingCart,
  Crown
} from 'lucide-vue-next'
import AIChat from '../components/AIChat.vue'
import TodayPlan from '../components/TodayPlan.vue'
import MyPlan from '../components/MyPlan.vue'
import BodyData from '../components/BodyData.vue'
import ProfileModal from '../components/ProfileModal.vue'
import AccountCenterModal from '../components/AccountCenterModal.vue'
import MembershipModal from '../components/MembershipModal.vue'
import CustomDialog from '../components/CustomDialog.vue'
import MallView from '../components/MallView.vue'
import { getProfile, deleteChatMessages, getMembershipInfo } from '../api'
import { useToastStore } from '../api/toast'
import { useDialogStore } from '../api/dialog'
import { useRouter } from 'vue-router'

const currentView = ref('today') // Default view from plan
const selectedChatId = ref(null)
const showProfile = ref(false)
const showAccountCenter = ref(false)
const showMembership = ref(false)
const isSidebarOpen = ref(false)
const isDark = ref(localStorage.getItem('theme') !== 'light')
const userName = ref('')
const avatarBase64 = ref('')
const userId = localStorage.getItem('user_id')
const toast = useToastStore()
const dialog = useDialogStore()
const router = useRouter()

const menuItems = [
  { id: 'today', name: '今日计划', icon: Calendar },
  { id: 'chat', name: 'AI 对话', icon: MessageSquare },
  { id: 'my-plan', name: '查看我的计划', icon: LayoutDashboard },
  { id: 'body-data', name: '查看身体数据', icon: LineChart },
  { id: 'mall', name: '健身商城', icon: ShoppingCart },
]

const chatHistory = ref([])
const membershipType = ref('FREE')
const CHAT_HISTORY_STORAGE_KEY = 'chat_history'
const CHAT_MESSAGES_STORAGE_KEY = 'chat_messages_by_id'

// Handle reactive view
const components = {
  today: TodayPlan,
  chat: AIChat,
  'my-plan': MyPlan,
  'body-data': BodyData,
  mall: MallView
}

const activeComp = computed(() => components[currentView.value])

const currentViewName = computed(() => {
  const item = menuItems.find(i => i.id === currentView.value)
  return item ? item.name : 'AI 健康助手'
})

function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function handleLogout() {
  localStorage.removeItem('token')
  localStorage.removeItem('user_id')
  toast.add({ type: 'info', message: '已退出登录' })
  router.push('/login')
}

function isProfileComplete(profile) {
  if (!profile || !profile.profile_exists) return false
  if (!profile.birth_date || !profile.height || !profile.weight || !profile.gender || !profile.target_intensity) {
    return false
  }
  if (!Array.isArray(profile.core_goals) || !profile.core_goals.length) {
    return false
  }
  if (profile.core_goals.includes('其他') && !String(profile.detailed_goal || '').trim()) {
    return false
  }
  return true
}

function ensureChatAllowed() {
  const profileRaw = localStorage.getItem('profile_cache')
  const profile = profileRaw ? JSON.parse(profileRaw) : null
  if (!isProfileComplete(profile)) {
    toast.add({
      type: 'warning',
      title: '请先完善个人信息',
      message: '缺少必填项，已为你打开资料弹窗。'
    })
    showProfile.value = true
    return false
  }
  return true
}

async function handleMenuClick(viewId) {
  if (viewId === 'chat' && !ensureChatAllowed()) {
    return
  }
  
  if (membershipType.value === 'FREE' && (viewId === 'today' || viewId === 'my-plan')) {
    const ok = await dialog.confirm('PRO 会员专属', '今日任务与运动计划是 PRO 会员功能，是否前往开通会员？(新用户可免费领一个月)')
    if (ok) {
      showMembership.value = true
    }
    return
  }

  currentView.value = viewId
  isSidebarOpen.value = false // Close sidebar on mobile
}

function startNewChat() {
  if (!ensureChatAllowed()) return
  currentView.value = 'chat'
  selectedChatId.value = createNewChat()
  isSidebarOpen.value = false // Close sidebar on mobile
}

function selectChat(chat) {
  if (!ensureChatAllowed()) return
  currentView.value = 'chat'
  selectedChatId.value = chat.id
  isSidebarOpen.value = false // Close sidebar on mobile
}

async function deleteChat(chatId) {
  const ok = await dialog.confirm('删除确认', '确认删除该对话记录吗？删除后无法恢复。')
  if (!ok) return
  if (userId && chatId) {
    try {
      await deleteChatMessages(Number(userId), String(chatId))
    } catch (e) {
      console.warn('delete chat messages from backend failed:', e)
    }
  }
  chatHistory.value = chatHistory.value.filter(chat => chat.id !== chatId)
  clearChatMessages(chatId)
  if (selectedChatId.value === chatId) {
    selectedChatId.value = chatHistory.value[0]?.id || null
  }
  persistChatHistory()
}

function persistChatHistory() {
  localStorage.setItem(CHAT_HISTORY_STORAGE_KEY, JSON.stringify(chatHistory.value))
}

function loadChatHistory() {
  const raw = localStorage.getItem(CHAT_HISTORY_STORAGE_KEY)
  if (!raw) {
    chatHistory.value = []
    return
  }
  try {
    const parsed = JSON.parse(raw)
    chatHistory.value = Array.isArray(parsed) ? parsed : []
  } catch {
    chatHistory.value = []
  }
}

function createNewChat() {
  const chatId = Date.now()
  chatHistory.value.unshift({
    id: chatId,
    title: `新对话 ${new Date().toLocaleString()}`,
    auto_named: true,
    updated_at: new Date().toISOString()
  })
  persistChatHistory()
  return chatId
}

function ensureChatCreated() {
  if (selectedChatId.value) {
    return selectedChatId.value
  }
  if (!ensureChatAllowed()) {
    return null
  }
  currentView.value = 'chat'
  selectedChatId.value = createNewChat()
  return selectedChatId.value
}

function renameChatFromFirstMessage(payload) {
  const chatId = payload?.chatId
  const text = String(payload?.text || '').trim()
  if (!chatId || !text) {
    return
  }
  const chat = chatHistory.value.find(item => item.id === chatId)
  if (!chat || chat.auto_named === false) {
    return
  }
  chat.title = text.slice(0, 24)
  chat.auto_named = false
  chat.updated_at = new Date().toISOString()
  chatHistory.value = [chat, ...chatHistory.value.filter(item => item.id !== chatId)]
  persistChatHistory()
}

function touchChat(chatId) {
  if (!chatId) {
    return
  }
  const chat = chatHistory.value.find(item => item.id === chatId)
  if (!chat) {
    return
  }
  chat.updated_at = new Date().toISOString()
  chatHistory.value = [chat, ...chatHistory.value.filter(item => item.id !== chatId)]
  persistChatHistory()
}

function clearChatMessages(chatId) {
  const raw = localStorage.getItem(CHAT_MESSAGES_STORAGE_KEY)
  if (!raw) {
    return
  }
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object') {
      return
    }
    delete parsed[String(chatId)]
    localStorage.setItem(CHAT_MESSAGES_STORAGE_KEY, JSON.stringify(parsed))
  } catch {
    localStorage.removeItem(CHAT_MESSAGES_STORAGE_KEY)
  }
}

async function loadProfile() {
  if (!userId) return
  try {
    const res = await getProfile(userId)
    if (res.data) {
      localStorage.setItem('profile_cache', JSON.stringify(res.data))
      userName.value = res.data.username || '用户'
      avatarBase64.value = res.data.avatar_base64 || ''
      if (!isProfileComplete(res.data)) {
        showProfile.value = true
      }
    }
    loadMembership()
  } catch (e) {
    console.error('Failed to load profile', e)
  }
}

async function loadMembership() {
  if (!userId) return
  try {
    const res = await getMembershipInfo(userId)
    const payload = res?.data || res || {}
    membershipType.value = payload.membership_type || 'FREE'
    localStorage.setItem('membership_type', membershipType.value)
  } catch (e) {
    console.error('Failed to load membership', e)
  }
}

function handleProfileUpdated() {
  loadProfile()
}

onMounted(() => {
  loadChatHistory()
  loadProfile()
})
</script>
