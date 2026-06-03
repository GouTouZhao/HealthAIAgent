<template>
  <div :class="['h-screen w-full flex items-center justify-center bg-background p-4 overflow-hidden relative transition-colors duration-300', isDark ? 'dark' : '']">
    <!-- Background Gradient -->
    <div class="absolute inset-0 bg-gradient-to-tr from-primary/10 via-background to-background pointer-events-none"></div>

    <div class="w-full max-w-md bg-card border border-border rounded-3xl p-10 shadow-2xl z-10 animate-in zoom-in-95 duration-500">
      <div class="text-center mb-10">
        <h1 class="text-3xl font-black text-foreground tracking-tight">创建账号</h1>
        <p class="text-muted-foreground mt-2">加入我们的社区，开启健康生活</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-6">
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

        <div class="space-y-2">
          <label class="text-xs font-bold text-muted-foreground uppercase tracking-widest px-1">确认密码</label>
          <div class="relative group">
            <ShieldCheck class="absolute left-4 top-3.5 w-5 h-5 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <input 
              v-model="confirmPassword" 
              type="password" 
              required
              class="w-full bg-muted border border-border rounded-2xl pl-12 pr-4 py-3.5 focus:border-primary/50 focus:ring-0 transition-all outline-none"
              placeholder="请再次输入密码"
            />
          </div>
        </div>

        <button 
          type="submit" 
          :disabled="loading"
          class="w-full py-4 bg-primary text-white rounded-2xl font-black text-lg shadow-xl shadow-primary/20 hover:bg-primary/90 disabled:opacity-30 disabled:cursor-not-allowed transition-all active:scale-[0.98] mt-4"
        >
          {{ loading ? '注册中...' : '注 册' }}
        </button>
      </form>

      <div class="mt-8 text-center">
        <p class="text-muted-foreground">
          已有账号? 
          <router-link to="/login" class="text-primary font-bold hover:underline ml-1">直接登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, ShieldCheck } from 'lucide-vue-next'
import { register } from '../api'

import { useToastStore } from '../api/toast'

const router = useRouter()
const toast = useToastStore()
const isDark = ref(localStorage.getItem('theme') !== 'light')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)

async function handleRegister() {
  if (password.value !== confirmPassword.value) {
    toast.add({ type: 'warning', message: '两次输入的密码不一致' })
    return
  }
  
  loading.value = true
  try {
    await register({
      username: username.value,
      password: password.value
    })
    toast.add({ type: 'success', message: '注册成功，请登录' })
    router.push('/login')
  } catch (e) {
    toast.add({ 
      type: 'error', 
      title: '注册失败',
      message: e.response?.data?.error || e.message 
    })
  } finally {
    loading.value = false
  }
}
</script>
