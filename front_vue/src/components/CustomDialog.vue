<script setup>
import { useDialogStore } from '../api/dialog'
import { X, AlertCircle, HelpCircle, Edit3 } from 'lucide-vue-next'

const dialog = useDialogStore()
</script>

<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="opacity-0 scale-95"
    enter-to-class="opacity-100 scale-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="opacity-100 scale-100"
    leave-to-class="opacity-0 scale-95"
  >
    <div v-if="dialog.show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="bg-white dark:bg-zinc-900 w-full max-w-sm rounded-3xl shadow-2xl overflow-hidden border border-zinc-200 dark:border-zinc-800 animate-in zoom-in-95 duration-200">
        <!-- Header -->
        <div class="p-6 text-center space-y-4">
          <div class="mx-auto w-12 h-12 rounded-2xl flex items-center justify-center" 
            :class="{
              'bg-orange-100 dark:bg-orange-900/30 text-orange-500': dialog.type === 'alert',
              'bg-blue-100 dark:bg-blue-900/30 text-blue-500': dialog.type === 'confirm',
              'bg-green-100 dark:bg-green-900/30 text-green-500': dialog.type === 'prompt'
            }">
            <AlertCircle v-if="dialog.type === 'alert'" class="w-6 h-6" />
            <HelpCircle v-if="dialog.type === 'confirm'" class="w-6 h-6" />
            <Edit3 v-if="dialog.type === 'prompt'" class="w-6 h-6" />
          </div>
          
          <div class="space-y-1">
            <h3 class="text-lg font-bold text-zinc-900 dark:text-zinc-100">{{ dialog.title }}</h3>
            <p class="text-sm text-zinc-500 dark:text-zinc-400 leading-relaxed">{{ dialog.message }}</p>
          </div>

          <div v-if="dialog.type === 'prompt'" class="pt-2">
            <input 
              v-model="dialog.inputValue"
              type="text"
              class="w-full bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl px-4 py-3 text-sm focus:border-orange-500 transition-all outline-none"
              placeholder="请输入..."
              @keyup.enter="dialog.handleConfirm"
              autofocus
            />
          </div>
        </div>

        <!-- Actions -->
        <div class="p-4 bg-zinc-50 dark:bg-zinc-800/50 flex gap-3">
          <button 
            v-if="dialog.type !== 'alert'"
            @click="dialog.handleCancel"
            class="flex-1 py-3 px-4 rounded-2xl font-bold text-sm text-zinc-500 hover:bg-zinc-200 dark:hover:bg-zinc-700 transition-colors"
          >
            {{ dialog.cancelText }}
          </button>
          <button 
            @click="dialog.handleConfirm"
            class="flex-1 py-3 px-4 rounded-2xl font-bold text-sm bg-orange-500 text-white shadow-lg shadow-orange-500/20 hover:bg-orange-600 transition-all active:scale-95"
          >
            {{ dialog.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.3s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
