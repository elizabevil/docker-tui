<script lang="ts" setup>
defineProps<{ title: string }>()
</script>

<template>
  <div class="tui-panel">
    <div class="tui-panel-title">{{ title }}</div>
    <slot/>
    <div v-if="$slots.footer" class="tui-footer">
      <slot name="footer"/>
    </div>
  </div>
</template>
