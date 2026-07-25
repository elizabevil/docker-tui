<script lang="ts" setup>
defineProps<{ shortcuts: { key: string; desc: string }[] }>()
</script>

<template>
  <div class="tui-shortcuts">
    <template v-for="(s, i) in shortcuts" :key="i">
      <span class="key">{{ s.key }}</span>
      <span class="desc">{{ s.desc }}</span>
      <span v-if="i < shortcuts.length - 1" class="sep">│</span>
    </template>
  </div>
</template>
