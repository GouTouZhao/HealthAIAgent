import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useToastStore = defineStore('toast', () => {
  const toasts = ref([])

  function add(toast) {
    const id = Date.now()
    const newToast = {
      id,
      title: toast.title || (toast.type === 'error' ? '错误' : '通知'),
      message: toast.message,
      type: toast.type || 'info', // info, success, error, warning
      duration: toast.duration || 3000
    }
    toasts.value.push(newToast)

    if (newToast.duration > 0) {
      setTimeout(() => {
        remove(id)
      }, newToast.duration)
    }
    return id
  }

  function remove(id) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index > -1) {
      toasts.value.splice(index, 1)
    }
  }

  return { toasts, add, remove }
})
