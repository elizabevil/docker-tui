<script lang="ts" setup>
import {watch} from 'vue'

const props = defineProps<{ message: string; type?: 'success' | 'error'; duration?: number }>()
const emit = defineEmits<{ dismiss: [] }>()

if (props.duration && props.duration > 0) {
  setTimeout(() => emit('dismiss'), props.duration)
}
</script>
<template>
  <transition name="toast-fade">
    <div v-if="message" :class="['tui-toast', type || 'success']">
      {{ type === 'error' ? '✕' : '✓' }} {{ message }}
    </div>
  </transition>
</template>
<style scoped>
.tui-toast {
  padding: 2px 8px;
  font-size: 12px;
  border-bottom: 1px solid var(--tui-surface)
}

.tui-toast.success {
  background: rgba(46, 204, 113, 0.15);
  color: var(--tui-green)
}

.tui-toast.error {
  background: rgba(239, 83, 80, 0.15);
  color: var(--tui-red)
}

.toast-fade-enter-active, .toast-fade-leave-active {
  transition: opacity .3s
}

.toast-fade-enter-from, .toast-fade-leave-to {
  opacity: 0
}
</style>
