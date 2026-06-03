<template>
  <div :class="['h-screen w-full flex items-center justify-center bg-background p-4 overflow-hidden relative transition-colors duration-300', isDark ? 'dark' : '']">
    <!-- Background Gradient -->
    <div class="absolute inset-0 bg-gradient-to-br from-primary/10 via-background to-background pointer-events-none"></div>
    <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/20 rounded-full blur-[128px] animate-pulse pointer-events-none"></div>

    <div class="w-full max-w-md bg-card border border-border rounded-3xl p-10 shadow-2xl z-10 animate-in zoom-in-95 duration-500">
      <div class="text-center mb-10">
        <div class="w-16 h-16 bg-primary/20 text-primary rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-lg shadow-primary/20">
          <Activity class="w-10 h-10" />
        </div>
        <h1 class="text-3xl font-black text-foreground tracking-tight">欢迎回来</h1>
        <p class="text-muted-foreground mt-2">登录以继续您的 AI 健身旅程</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-6">
        <div class="space-y-2">
          <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest px-1">用户名</label>
          <div class="relative group">
            <User class="absolute left-4 top-3.5 w-5 h-5 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <input 
              v-model="username" 
              type="text" 
              required
              class="w-full bg-muted border border-border rounded-2xl pl-12 pr-4 py-3.5 focus:border-primary/50 focus:ring-0 transition-all outline-none"
              placeholder="请输入用户名"
            />
          </div>
        </div>

        <div class="space-y-2">
          <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest px-1">密码</label>
          <div class="relative group">
            <Lock class="absolute left-4 top-3.5 w-5 h-5 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <input 
              v-model="password" 
              type="password" 
              required
              class="w-full bg-muted border border-border rounded-2xl pl-12 pr-4 py-3.5 focus:border-primary/50 focus:ring-0 transition-all outline-none"
              placeholder="请输入密码"
            />
          </div>
        </div>

        <button 
          type="submit" 
          :disabled="loading"
          class="w-full py-4 bg-primary text-white rounded-2xl font-black text-lg shadow-xl shadow-primary/20 hover:bg-primary/90 disabled:opacity-30 disabled:cursor-not-allowed transition-all active:scale-[0.98] mt-4"
        >
          {{ loading ? '登录中...' : '登 录' }}
        </button>
      </form>

      <div class="mt-8 text-center">
        <p class="text-muted-foreground">
          还没有账号? 
          <router-link to="/register" class="text-primary font-bold hover:underline ml-1">立即注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Activity } from 'lucide-vue-next'
import { login } from '../api'

import { useToastStore } from '../api/toast'

const router = useRouter()
const toast = useToastStore()
const isDark = ref(localStorage.getItem('theme') !== 'light')
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!username.value || !password.value) {
    toast.add({ type: 'warning', message: '请输入用户名和密码' })
    return
  }
  
  loading.value = true
  try {
    const res = await login({
      username: username.value,
      password: password.value
    })
    
    if (res.data && res.data.token) {
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('user_id', res.data.user_id.toString())
      toast.add({ type: 'success', message: '登录成功' })
      router.push('/')
    } else {
      toast.add({ type: 'error', message: '登录失败：无效的响应' })
    }
  } catch (e) {
    toast.add({ 
      type: 'error', 
      title: '登录失败',
      message: e.response?.data?.error || e.message 
    })
  } finally {
    loading.value = false
  }
}
</script>
