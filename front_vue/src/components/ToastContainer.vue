<template>
  <div class="fixed top-6 right-6 z-[9999] flex flex-col gap-3 pointer-events-none">
    <TransitionGroup 
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="translate-x-full opacity-0"
      enter-to-class="translate-x-0 opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="translate-x-0 opacity-100"
      leave-to-class="translate-x-full opacity-0"
    >
      <div 
        v-for="toast in toastStore.toasts" 
        :key="toast.id"
        class="pointer-events-auto min-w-[300px] max-w-md p-4 rounded-2xl border border-border bg-card shadow-2xl flex gap-4 animate-in slide-in-from-right-2"
      >
        <div :class="['w-10 h-10 rounded-xl flex-shrink-0 flex items-center justify-center', iconBgColor(toast.type)]">
          <component :is="getIcon(toast.type)" :class="['w-6 h-6', iconTextColor(toast.type)]" />
        </div>
        
        <div class="flex-1 pr-2">
          <div class="text-sm font-bold text-foreground">{{ toast.title }}</div>
          <div class="text-xs text-muted-foreground mt-1 leading-relaxed">{{ toast.message }}</div>
        </div>

        <button @click="toastStore.remove(toast.id)" class="text-muted-foreground hover:text-foreground transition-colors h-fit p-1">
          <X class="w-4 h-4" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { X, Info, CheckCircle2, AlertCircle, AlertTriangle } from 'lucide-vue-next'
import { useToastStore } from '../api/toast'

const toastStore = useToastStore()

function getIcon(type) {
  switch (type) {
    case 'success': return CheckCircle2
    case 'error': return AlertCircle
    case 'warning': return AlertTriangle
    default: return Info
  }
}

function iconBgColor(type) {
  switch (type) {
    case 'success': return 'bg-green-500/10'
    case 'error': return 'bg-red-500/10'
    case 'warning': return 'bg-yellow-500/10'
    default: return 'bg-primary/10'
  }
}

function iconTextColor(type) {
  switch (type) {
    case 'success': return 'text-green-500'
    case 'error': return 'text-red-500'
    case 'warning': return 'text-yellow-500'
    default: return 'text-primary'
  }
}
</script>
