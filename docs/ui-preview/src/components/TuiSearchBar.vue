<script lang="ts" setup>
defineProps<{ modelValue: string; placeholder?: string }>()
const emit = defineEmits<{ 'update:modelValue': [v: string] }>()
</script>
<template>
  <div class="tui-searchbar">
    <span class="tui-searchbar-icon">/</span>
    <input :placeholder="placeholder || 'Filter...'" :value="modelValue"
           autofocus
           class="tui-searchbar-input" @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"/>
  </div>
</template>
<style scoped>
.tui-searchbar {
  display: flex;
  align-items: center;
  background: var(--tui-dark);
  padding: 2px 8px;
  border-bottom: 1px solid var(--tui-surface)
}

.tui-searchbar-icon {
  color: var(--tui-cyan);
  margin-right: 6px;
  font-weight: bold
}

.tui-searchbar-input {
  background: transparent;
  border: none;
  color: var(--tui-white);
  outline: none;
  flex: 1;
  font-size: 13px
}

.tui-searchbar-input::placeholder {
  color: var(--tui-gray)
}
</style>
