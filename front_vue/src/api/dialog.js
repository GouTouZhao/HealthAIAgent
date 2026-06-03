import { defineStore } from 'pinia'

export const useDialogStore = defineStore('dialog', {
  state: () => ({
    show: false,
    title: '',
    message: '',
    type: 'alert', // 'alert', 'confirm', 'prompt'
    confirmText: '确定',
    cancelText: '取消',
    inputValue: '',
    resolve: null,
  }),
  actions: {
    alert(title, message) {
      this.title = title
      this.message = message
      this.type = 'alert'
      this.show = true
      return new Promise((resolve) => {
        this.resolve = resolve
      })
    },
    confirm(title, message) {
      this.title = title
      this.message = message
      this.type = 'confirm'
      this.show = true
      return new Promise((resolve) => {
        this.resolve = resolve
      })
    },
    prompt(title, message, defaultValue = '') {
      this.title = title
      this.message = message
      this.type = 'prompt'
      this.inputValue = defaultValue
      this.show = true
      return new Promise((resolve) => {
        this.resolve = resolve
      })
    },
    handleConfirm() {
      const result = this.type === 'prompt' ? this.inputValue : true
      this.show = false
      if (this.resolve) this.resolve(result)
    },
    handleCancel() {
      this.show = false
      if (this.resolve) this.resolve(false)
    }
  }
})
