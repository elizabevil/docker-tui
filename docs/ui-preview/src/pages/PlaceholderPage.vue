<script lang="ts" setup>
import {useRoute} from 'vue-router'

const route = useRoute()
</script>

<template>
  <div class="tui-placeholder">
    <div class="tui-placeholder-icon">&#9632;</div>
    <div class="tui-placeholder-title">{{ route.meta.title }}</div>
    <div class="tui-placeholder-desc">UI preview coming soon</div>
  </div>
</template>

<style scoped>
.tui-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: var(--tui-gray);
}

.tui-placeholder-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.tui-placeholder-title {
  font-size: 18px;
  color: var(--tui-white);
  margin-bottom: 8px;
}

.tui-placeholder-desc {
  font-size: 13px;
}
</style>
