<template>
  <div class="flex flex-col h-full max-w-5xl mx-auto px-4 relative">
    <!-- Chat Header -->
    <div class="hidden lg:flex py-4 border-b border-border items-center justify-between bg-background/80 backdrop-blur-sm sticky top-0 z-10">
      <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-muted border border-border">
        <span class="text-sm font-semibold">AI 对话</span>
      </div>
      <div class="text-sm text-muted-foreground flex items-center gap-2">
        <div class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
        AI 已就绪
      </div>
    </div>

    <!-- Message List -->
    <div class="flex-1 overflow-y-auto py-8 space-y-8 scroll-smooth" ref="messageContainer" @scroll="handleMessageScroll">
      <div v-if="historyLoadingMore" class="sticky top-2 z-10 flex justify-center">
        <div class="px-3 py-1 rounded-full border border-border bg-card text-xs text-muted-foreground flex items-center gap-2">
          <span class="inline-block w-3 h-3 border-2 border-muted-foreground/40 border-t-muted-foreground rounded-full animate-spin"></span>
          正在加载更多对话...
        </div>
      </div>
      <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full opacity-30 text-center gap-4">
        <div class="w-16 h-16 rounded-2xl bg-muted flex items-center justify-center">
          <MessageSquare class="w-8 h-8" />
        </div>
        <div>
          <h2 class="text-2xl font-bold">你好，我是你的 AI 助手</h2>
          <p class="mt-2">请选择上方模式开始对话，或者直接向我提问</p>
        </div>
      </div>

      <div v-for="(msg, index) in messages" :key="index" :class="['flex gap-4', msg.role === 'user' ? 'flex-row-reverse' : '']">
        <!-- Avatar -->
        <div :class="[
          'w-10 h-10 rounded-xl flex-shrink-0 flex items-center justify-center',
          msg.role === 'user' ? 'bg-primary/20 text-primary' : 'bg-muted text-foreground'
        ]">
          <img v-if="msg.role === 'user' && effectiveUserAvatar" :src="effectiveUserAvatar" class="w-full h-full object-cover" />
          <User v-else-if="msg.role === 'user'" class="w-6 h-6" />
          <Bot v-else class="w-6 h-6" />
        </div>

        <!-- Content -->
        <div :class="[
          'max-w-[90%] md:max-w-[85%] rounded-2xl p-4',
          msg.role === 'user' ? 'bg-muted/50 border border-border' : 'bg-card border border-border'
        ]">
          <!-- User Image if any -->
          <img v-if="msg.image" :src="msg.image" class="max-w-[200px] md:max-w-xs rounded-lg mb-3 border border-border" />
          
          <div class="text-sm leading-relaxed whitespace-pre-wrap text-foreground">{{ msg.content }}</div>

          <!-- JSON Table Rendering -->
          <div v-if="msg.json" class="mt-4 overflow-x-auto border border-border rounded-xl">
            <component 
              :is="getRenderComponent(msg.jsonType)" 
              :data="msg.json"
            />
          </div>

          <div class="mt-2 text-[10px] text-muted-foreground/80">{{ formatTime(msg.timestamp) }}</div>
        </div>
      </div>

      <div v-if="loading" class="flex gap-4">
        <div class="w-10 h-10 rounded-xl bg-muted flex items-center justify-center">
          <Bot class="w-6 h-6" />
        </div>
        <div class="bg-card border border-border rounded-2xl p-4">
          <div class="text-xs text-muted-foreground mb-2">{{ loadingStepText }}</div>
          <div class="flex gap-1">
            <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce"></div>
            <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce [animation-delay:0.2s]"></div>
            <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce [animation-delay:0.4s]"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div
      class="py-6 bg-background"
      @dragenter.prevent="handleDragEnter"
      @dragover.prevent="handleDragOver"
      @dragleave.prevent="handleDragLeave"
      @drop.prevent="handleDrop"
    >
      <div v-if="pendingImage" class="mb-3 flex justify-start animate-in fade-in slide-in-from-top-2">
        <div class="inline-flex items-start gap-3 p-3 rounded-2xl border border-border bg-card shadow-lg">
          <img :src="pendingImage" class="h-24 w-24 md:h-36 md:w-36 object-cover rounded-xl" />
          <button @click="pendingImage = null" class="p-1 hover:bg-red-500/20 rounded-full text-red-500 transition-colors" title="移除图片">
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <div class="mb-3 flex flex-wrap gap-2">
        <button
          v-for="mode in modes"
          :key="mode.value"
          @click="handleModeClick(mode.value)"
          :class="[
            'px-4 py-2 rounded-xl text-xs md:text-sm font-semibold border transition-all relative overflow-hidden',
            currentMode === mode.value ? 'bg-primary border-primary text-white shadow-lg shadow-primary/20' : 'bg-muted border-border text-muted-foreground hover:border-primary/40',
            (props.membershipType === 'FREE' && (mode.value === '2' || mode.value === '3')) ? 'opacity-70 grayscale-[0.5]' : ''
          ]"
        >
          {{ mode.label }}
          <Crown v-if="props.membershipType === 'FREE' && (mode.value === '2' || mode.value === '3')" class="w-3 h-3 absolute top-0.5 right-0.5 text-orange-500" />
        </button>
      </div>

      <div
        v-if="dragActive"
        class="mb-3 h-32 rounded-2xl border-2 border-dashed border-primary/70 bg-primary/10 flex items-center justify-center text-sm font-medium text-primary"
      >
        请将图片拖入此处
      </div>

      <div class="relative group">
        <!-- Input Box -->
        <div class="flex items-end gap-2 p-2 rounded-2xl border border-border bg-card focus-within:border-primary/50 transition-all duration-300 shadow-lg">
          <button 
            @click="triggerImageUpload"
            class="p-3 text-muted-foreground hover:text-foreground transition-colors rounded-xl hover:bg-muted"
            title="上传图片"
          >
            <ImagePlus class="w-6 h-6" />
          </button>
          
          <textarea
            ref="inputRef"
            v-model="input"
            rows="1"
            placeholder="输入您的问题..."
            class="flex-1 bg-transparent border-none focus:ring-0 text-sm py-3 px-2 resize-none max-h-40 overflow-y-auto"
            @keydown.enter.prevent="handleEnter"
            @input="autoResize"
            @paste="handlePaste"
          ></textarea>

          <button 
            @click="sendMessage"
            :disabled="!input.trim() && !pendingImage || loading"
            class="p-2.5 md:p-3 bg-primary text-white rounded-xl disabled:opacity-30 disabled:cursor-not-allowed hover:bg-primary/90 transition-all active:scale-95 shadow-lg shadow-primary/20"
          >
            <Send class="w-5 h-5 md:w-6 md:h-6" />
          </button>
        </div>
        <input type="file" ref="fileInput" class="hidden" accept="image/*" @change="handleImageChange" />
      </div>
      <p class="text-[10px] text-muted-foreground text-center mt-3">默认模式：其他</p>
    </div>

    <div v-if="showReplacePlanConfirm" class="absolute inset-0 z-20 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="w-full max-w-md rounded-2xl border border-border bg-card p-5">
        <h3 class="text-base font-semibold">检测到当前已有训练计划</h3>
        <p class="text-sm text-muted-foreground mt-2">继续发送“制定计划”将替换当前计划。你可以取消，输入内容会保留在输入框。</p>

        <div class="mt-4 flex justify-end gap-2">
          <button
            @click="cancelReplacePlanConfirm"
            class="px-3 py-2 rounded-lg border border-border text-sm text-muted-foreground hover:text-foreground"
          >
            取消
          </button>
          <button
            @click="confirmReplacePlanAndSend"
            class="px-3 py-2 rounded-lg bg-primary text-white text-sm font-semibold"
          >
            替换并发送（{{ replacePlanCountdown }}s）
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, watch, onBeforeUnmount, computed } from 'vue'
import { 
  Send, 
  ImagePlus, 
  User, 
  Bot, 
  MessageSquare, 
  X,
  Crown
} from 'lucide-vue-next'
import FoodResultTable from './render/FoodResultTable.vue'
import PlanTable from './render/PlanTable.vue'
import { getChatMessages } from '../api'
import { useDialogStore } from '../api/dialog'

const props = defineProps({
  chatId: {
    type: [String, Number],
    default: null
  },
  chatHistory: {
    type: Array,
    default: () => []
  },
  userAvatar: {
    type: String,
    default: ''
  },
  membershipType: {
    type: String,
    default: 'FREE'
  }
})

const emit = defineEmits(['ensure-chat-created', 'rename-chat-from-first-message', 'touch-chat', 'open-membership'])

const dialog = useDialogStore()
const currentMode = ref('4')
const input = ref('')
const loading = ref(false)
const messages = ref([])
const pendingImage = ref(null)
const inputRef = ref(null)
const fileInput = ref(null)
const messageContainer = ref(null)
const dragActive = ref(false)
const dragCounter = ref(0)
const loadingStepText = ref('等待开始...')
const historyLoadingMore = ref(false)
const historyHasMore = ref(true)
const historyBeforeId = ref(null)
const INITIAL_PAIR_COUNT = 5
const LOAD_MORE_PAIR_COUNT = 3
const skipNextChatInitLoad = ref(false)
const showReplacePlanConfirm = ref(false)
const replacePlanCountdown = ref(10)
const replacePlanTimer = ref(null)

const modes = [
  { value: '1', label: '食物分析' },
  { value: '2', label: '制定计划' },
  { value: '3', label: '修改计划' },
  { value: '4', label: '其他' }
]

const userId = Number(localStorage.getItem('user_id') || 0)
const effectiveUserAvatar = computed(() => {
  if (props.userAvatar) {
    return props.userAvatar
  }
  try {
    const profileCache = JSON.parse(localStorage.getItem('profile_cache') || '{}')
    return profileCache.avatar_base64 || ''
  } catch {
    return ''
  }
})

function autoResize() {
  const el = inputRef.value
  el.style.height = 'auto'
  el.style.height = el.scrollHeight + 'px'
}

function triggerImageUpload() {
  fileInput.value.click()
}

async function handleImageChange(e) {
  const file = e.target.files[0]
  if (file) {
    await readFile(file)
  }
  e.target.value = ''
}

async function readFile(file) {
  if (!file.type.startsWith('image/')) return
  const dataUrl = await readFileAsDataUrl(file)
  pendingImage.value = await compressImageDataUrl(dataUrl)
}

function readFileAsDataUrl(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      resolve(e.target.result)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function loadImage(dataUrl) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = reject
    img.src = dataUrl
  })
}

async function compressImageDataUrl(dataUrl) {
  const img = await loadImage(dataUrl)
  const width = img.naturalWidth || img.width
  const height = img.naturalHeight || img.height
  if (!width || !height) {
    return dataUrl
  }

  const shortSide = Math.min(width, height)
  const longSide = Math.max(width, height)

  let scale = 1
  if (shortSide > 224) {
    scale = 224 / shortSide
  } else if (longSide > 500) {
    scale = 500 / longSide
  }

  if (scale >= 1) {
    return dataUrl
  }

  const targetWidth = Math.max(1, Math.round(width * scale))
  const targetHeight = Math.max(1, Math.round(height * scale))
  const canvas = document.createElement('canvas')
  canvas.width = targetWidth
  canvas.height = targetHeight
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    return dataUrl
  }
  ctx.drawImage(img, 0, 0, targetWidth, targetHeight)
  return canvas.toDataURL('image/jpeg', 0.88)
}

async function handlePaste(e) {
  const item = Array.from(e.clipboardData.items).find(x => x.type.startsWith('image/'))
  if (item) {
    const file = item.getAsFile()
    await readFile(file)
  }
}

function buildChangeSummaryText(resultData) {
  const changes = Array.isArray(resultData?.changes)
    ? resultData.changes.map(item => String(item || '').trim()).filter(Boolean)
    : []
  if (!changes.length) {
    return '计划修改完成。'
  }
  return `计划修改完成：\n- ${changes.join('\n- ')}`
}

async function handleDrop(e) {
  dragActive.value = false
  dragCounter.value = 0
  const file = Array.from(e.dataTransfer.files).find(item => item.type.startsWith('image/'))
  if (file) {
    await readFile(file)
  }
}

function handleDragEnter(e) {
  if (!hasImageFile(e)) return
  dragCounter.value += 1
  dragActive.value = true
}

function handleDragOver(e) {
  if (!hasImageFile(e)) return
  dragActive.value = true
}

function handleDragLeave() {
  dragCounter.value = Math.max(0, dragCounter.value - 1)
  if (dragCounter.value === 0) {
    dragActive.value = false
  }
}

function hasImageFile(e) {
  const items = Array.from(e.dataTransfer?.items || [])
  if (items.some(item => item.kind === 'file' && item.type.startsWith('image/'))) {
    return true
  }
  return Array.from(e.dataTransfer?.files || []).some(file => file.type.startsWith('image/'))
}

function handleEnter(e) {
  if (e.shiftKey) return
  sendMessage()
}

function hasExistingCurrentPlan() {
  try {
    const profileCache = JSON.parse(localStorage.getItem('profile_cache') || '{}')
    const rawPlan = profileCache?.latest_plan_json || profileCache?.plan_data
    if (!rawPlan) {
      return false
    }
    if (typeof rawPlan === 'string') {
      return rawPlan.trim() !== '' && rawPlan.trim() !== '{}' && rawPlan.trim() !== 'null'
    }
    if (typeof rawPlan === 'object') {
      return Object.keys(rawPlan).length > 0
    }
    return false
  } catch {
    return false
  }
}

function clearReplacePlanTimer() {
  if (replacePlanTimer.value) {
    clearInterval(replacePlanTimer.value)
    replacePlanTimer.value = null
  }
}

function openReplacePlanConfirm() {
  showReplacePlanConfirm.value = true
  replacePlanCountdown.value = 10
  clearReplacePlanTimer()
  replacePlanTimer.value = setInterval(() => {
    replacePlanCountdown.value -= 1
    if (replacePlanCountdown.value <= 0) {
      confirmReplacePlanAndSend()
    }
  }, 1000)
}

function cancelReplacePlanConfirm() {
  showReplacePlanConfirm.value = false
  clearReplacePlanTimer()
}

function confirmReplacePlanAndSend() {
  showReplacePlanConfirm.value = false
  clearReplacePlanTimer()
  sendMessage(true)
}

function updateProfilePlanCache(resultData) {
  let nextPlan = null
  if (resultData?.new_plan && typeof resultData.new_plan === 'object') {
    nextPlan = resultData.new_plan
  } else if (resultData?.months && typeof resultData === 'object') {
    nextPlan = resultData
  }
  if (!nextPlan) {
    return
  }
  try {
    const profileCache = JSON.parse(localStorage.getItem('profile_cache') || '{}')
    profileCache.latest_plan_json = JSON.stringify(nextPlan)
    localStorage.setItem('profile_cache', JSON.stringify(profileCache))
  } catch {
    // ignore cache update errors
  }
}

async function handleModeClick(modeValue) {
  if (props.membershipType === 'FREE' && (modeValue === '2' || modeValue === '3')) {
    const ok = await dialog.confirm('PRO 会员专属', '制定与修改计划是 PRO 会员功能，是否前往开通会员？(新用户可免费领一个月)')
    if (ok) {
      emit('open-membership')
    }
    return
  }
  currentMode.value = modeValue
}

async function sendMessage(ignoreReplaceConfirm = false) {
  if (loading.value || (!input.value.trim() && !pendingImage.value)) return
  if (!ignoreReplaceConfirm && currentMode.value === '2' && hasExistingCurrentPlan()) {
    openReplacePlanConfirm()
    return
  }

  let activeChatId = props.chatId
  if (!activeChatId) {
    skipNextChatInitLoad.value = true
    emit('ensure-chat-created')
    await nextTick()
    activeChatId = props.chatId
  }
  if (!activeChatId) {
    skipNextChatInitLoad.value = false
    return
  }

  const isFirstMessageInChat = messages.value.length === 0
  
  const userMsg = {
    role: 'user',
    content: input.value,
    image: pendingImage.value,
    timestamp: new Date().toISOString()
  }
  
  messages.value.push(userMsg)
  const currentInput = input.value
  const currentImg = pendingImage.value
  const mode = currentMode.value

  if (isFirstMessageInChat && String(currentInput || '').trim()) {
    emit('rename-chat-from-first-message', {
      chatId: activeChatId,
      text: currentInput
    })
  }

  emit('touch-chat', activeChatId)
  
  input.value = ''
  pendingImage.value = null
  loading.value = true
  loadingStepText.value = '当前步骤：请求已发送'
  
  await nextTick()
  scrollToBottom()

  try {
    const resultData = await askAgentWithStream({
      user_id: userId,
      chat_id: String(activeChatId),
      input: currentInput,
      image: currentImg,
      mode: mode
    })
    
    let botMsg = { role: 'assistant' }
    
    // Check if the response is JSON (based on mode or content)
    if (mode === '1' && resultData.calories_per_100g !== undefined) {
      botMsg.json = resultData
      botMsg.jsonType = 'food'
      botMsg.content = `识别结果: ${resultData.food_name}`
    } else if ((mode === '2' || mode === '3') && resultData.new_plan) {
      botMsg.json = resultData.new_plan
      botMsg.jsonType = 'plan'
      botMsg.content = mode === '2' ? '计划制定完成:' : buildChangeSummaryText(resultData)
    } else if ((mode === '2' || mode === '3') && resultData.months) {
      botMsg.json = resultData
      botMsg.jsonType = 'plan'
      botMsg.content = '计划制定完成:'
    } else if (mode === '4' && typeof resultData?.answer === 'string' && resultData.answer.trim()) {
      botMsg.content = resultData.answer.trim()
    } else {
      botMsg.content = typeof resultData === 'string' ? resultData : JSON.stringify(resultData, null, 2)
    }

    if (mode === '2' || mode === '3') {
      updateProfilePlanCache(resultData)
    }
    
    botMsg.timestamp = new Date().toISOString()
    messages.value.push(botMsg)
    emit('touch-chat', activeChatId)
  } catch (e) {
    messages.value.push({
      role: 'assistant',
      content: '抱歉，处理您的请求时出现了错误：' + (e.response?.data?.error || e.message),
      timestamp: new Date().toISOString()
    })
    emit('touch-chat', activeChatId)
  } finally {
    loading.value = false
    loadingStepText.value = '等待开始...'
    nextTick(scrollToBottom)
  }
}

async function askAgentWithStream(payload) {
  const token = localStorage.getItem('token')
  const headers = { 'Content-Type': 'application/json' }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const response = await fetch('/api/agent/ask-stream', {
    method: 'POST',
    headers,
    body: JSON.stringify(payload)
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.error || `请求失败(${response.status})`)
  }

  if (!response.body) {
    throw new Error('未收到流式响应')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let finalData = null

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const rawLine of lines) {
      const line = rawLine.trim()
      if (!line) continue

      let event
      try {
        event = JSON.parse(line)
      } catch {
        continue
      }

      if (event.type === 'step' && event.step) {
        loadingStepText.value = `当前步骤：${event.step}`
      } else if (event.type === 'error') {
        throw new Error(event.error || '流水线执行失败')
      } else if (event.type === 'result') {
        finalData = event.data
      }
    }
  }

  if (!finalData) {
    throw new Error('未收到最终结果')
  }
  return finalData
}

function scrollToBottom() {
  if (messageContainer.value) {
    messageContainer.value.scrollTop = messageContainer.value.scrollHeight
  }
}

function formatTime(timestamp) {
  if (!timestamp) {
    return ''
  }
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function normalizeMessage(item) {
  let parsedJSON = null
  if (item?.json_data) {
    try {
      parsedJSON = JSON.parse(item.json_data)
    } catch {
      parsedJSON = null
    }
  }
  return {
    id: item?.id,
    role: item?.role,
    content: item?.content || '',
    image: item?.image_base64 || null,
    timestamp: item?.created_at || new Date().toISOString(),
    json: parsedJSON,
    jsonType: item?.json_type || null
  }
}

async function loadMessages(chatId, pairCount = INITIAL_PAIR_COUNT, beforeId = null) {
  if (!chatId || !userId) {
    return { list: [], hasMore: false, nextBeforeId: null }
  }
  const response = await getChatMessages(userId, String(chatId), pairCount, beforeId)
  const data = response?.data || {}
  return {
    list: Array.isArray(data.messages) ? data.messages.map(normalizeMessage) : [],
    hasMore: Boolean(data.has_more),
    nextBeforeId: data.next_before_id || null
  }
}

async function loadInitialMessages(chatId) {
  if (!chatId) {
    messages.value = []
    historyHasMore.value = false
    historyBeforeId.value = null
    return
  }
  try {
    const { list, hasMore, nextBeforeId } = await loadMessages(chatId, INITIAL_PAIR_COUNT)
    messages.value = list
    historyHasMore.value = hasMore
    historyBeforeId.value = nextBeforeId
  } catch {
    messages.value = []
    historyHasMore.value = false
    historyBeforeId.value = null
  }
}

async function loadMoreMessages() {
  if (!props.chatId || !historyHasMore.value || historyLoadingMore.value || !historyBeforeId.value) {
    return
  }

  const container = messageContainer.value
  const prevHeight = container?.scrollHeight || 0
  historyLoadingMore.value = true
  try {
    const { list, hasMore, nextBeforeId } = await loadMessages(props.chatId, LOAD_MORE_PAIR_COUNT, historyBeforeId.value)
    if (list.length > 0) {
      messages.value = [...list, ...messages.value]
      await nextTick()
      if (container) {
        const diff = container.scrollHeight - prevHeight
        container.scrollTop += diff
      }
    }
    historyHasMore.value = hasMore
    historyBeforeId.value = nextBeforeId
  } finally {
    historyLoadingMore.value = false
  }
}

function handleMessageScroll(e) {
  if (e.target?.scrollTop <= 20) {
    loadMoreMessages()
  }
}

watch(
  () => props.chatId,
  async (newChatId) => {
    if (skipNextChatInitLoad.value) {
      skipNextChatInitLoad.value = false
      return
    }
    await loadInitialMessages(newChatId)
    await nextTick()
    scrollToBottom()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  loadingStepText.value = '等待开始...'
  clearReplacePlanTimer()
})

function getRenderComponent(type) {
  if (type === 'food') return FoodResultTable
  if (type === 'plan') return PlanTable
  return null
}
</script>
