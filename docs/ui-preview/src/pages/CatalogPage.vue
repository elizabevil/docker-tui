<script lang="ts" setup>
import {useRouter} from 'vue-router'

const router = useRouter()

const items = [
  {path: '/containers', icon: '\u2699', label: 'Container Table'},
  {path: '/images', icon: '\uD83D\uDDBC', label: 'Image Table'},
  {path: '/detail', icon: '\uD83D\uDCCA', label: 'Detail View'},
  {path: '/logs', icon: '\uD83D\uDCCB', label: 'Log View'},
  {path: '/help', icon: '?', label: 'Help View'},
  {path: '/volumes', icon: '\uD83D\uDCE6', label: 'Volume List'},
  {path: '/networks', icon: '\uD83C\uDF10', label: 'Network List'},
]
</script>

<template>
  <div class="tui-catalog">
    <div v-for="item in items" :key="item.path" class="tui-catalog-item" @click="router.push(item.path)">
      <div class="icon">{{ item.icon }}</div>
      <div class="label">{{ item.label }}</div>
    </div>
  </div>
</template>

<style scoped>
.tui-catalog {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px;
}

.tui-catalog-item {
  background: var(--tui-bg);
  border: 1px solid var(--tui-border);
  border-radius: 4px;
  padding: 12px;
  cursor: pointer;
  min-width: 140px;
  text-align: center;
  transition: all 0.15s;
}

.tui-catalog-item:hover {
  border-color: var(--tui-cyan);
}

.tui-catalog-item .label {
  color: var(--tui-gray);
  font-size: 12px;
  margin-top: 4px;
}

.tui-catalog-item .icon {
  font-size: 24px;
}
</style>
