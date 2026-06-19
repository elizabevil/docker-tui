<script lang="ts" setup>
defineProps<{ title?: string }>()
</script>

<template>
  <div class="tui-frame">
    <div class="tui-frame-titlebar">
      <span class="tui-frame-dot red"></span>
      <span class="tui-frame-dot yellow"></span>
      <span class="tui-frame-dot green"></span>
      <span class="tui-frame-title">dtui — {{ title || 'Terminal' }}</span>
    </div>
    <div class="tui-frame-body">
      <slot/>
    </div>
  </div>
</template>
