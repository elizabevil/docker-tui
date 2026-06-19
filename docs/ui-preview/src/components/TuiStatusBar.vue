<script lang="ts" setup>
defineProps<{ text: string }>()
</script>
<template>
  <div class="tui-statusbar"><span>{{ text }}</span></div>
</template>
<style scoped>
.tui-statusbar {
  background: var(--tui-surface);
  color: var(--tui-white);
  padding: 2px 8px;
  font-size: 12px;
  border-top: 1px solid var(--tui-dark)
}
</style>
