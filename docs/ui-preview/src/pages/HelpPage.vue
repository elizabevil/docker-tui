<script lang="ts" setup>
import TuiFrame from '../components/TuiFrame.vue'

interface HelpSection {
  title: string;
  items: { keys: string; desc: string }[]
}

const sections: HelpSection[] = [
  {
    title: 'Navigation', items: [
      {keys: '\u2191\u2193 / j k', desc: 'Move cursor'},
      {keys: 'Tab / S-Tab', desc: 'Switch panel'},
      {keys: '\u2192 / Enter', desc: 'Expand / Detail'},
      {keys: '\u2190 / Esc', desc: 'Collapse / Back'},
    ]
  },
  {
    title: 'Containers', items: [
      {keys: 's S R K', desc: 'Start Stop Restart Kill'},
      {keys: 'l m d', desc: 'Logs Stats Detail'},
      {keys: 'Ctrl+D', desc: 'Delete'},
    ]
  },
  {
    title: 'Images', items: [
      {keys: 'P p o', desc: 'Pull Prune Sort'},
      {keys: 'D E', desc: 'Debug Export'},
      {keys: '\u2192 / Enter', desc: 'Show containers'},
    ]
  },
  {
    title: 'Compose', items: [
      {keys: 's S l Ctrl+D', desc: 'Start Stop Logs Down'},
      {keys: '\u2192 / Enter', desc: 'Expand services'},
      {keys: '\u2190', desc: 'Collapse'},
    ]
  },
  {
    title: 'Volumes & Networks', items: [
      {keys: 'd', desc: 'Detail / inspect'},
      {keys: 'Ctrl+D', desc: 'Delete'},
    ]
  },
  {
    title: 'Global', items: [
      {keys: 'Space', desc: 'Mark for bulk ops'},
      {keys: 'H', desc: 'Toggle header bar'},
      {keys: '/ ? q', desc: 'Filter Help Quit'},
    ]
  },
]
</script>

<template>
  <TuiFrame title="Help">
    <div class="tui-help">
      <div class="tui-help-logo">
        <pre class="tui-help-logo-text">\
  \u2588\u2588\u2588\u2588\u2588\u2557\u2588\u2588\u2588\u2588\u2588\u2588\u2557\u2588\u2588\u2557   \u2588\u2588\u2557\u2588\u2588\u2557
  \u2588\u2588\u2554\u2550\u2550\u2588\u2588\u2557\u259a\u259a\u2588\u2588\u259a\u259d\u2588\u2588\u2551   \u2588\u2588\u2551\u2588\u2588\u2551
  \u2588\u2588\u2551  \u2588\u2588\u2551   \u2588\u2588\u2551 \u2588\u2588\u2551   \u2588\u2588\u2551\u2588\u2588\u2551
  \u2588\u2588\u2551  \u2588\u2588\u2551   \u2588\u2588\u2551 \u2588\u2588\u2551   \u2588\u2588\u2551\u2588\u2588\u2551
  \u2588\u2588\u2588\u2588\u2588\u255d   \u2588\u2588\u2551 \u2595\u2588\u2588\u2588\u2588\u2588\u2595\u2588\u2588\u2551
  \u255a\u2550\u2550\u2550\u2550\u2550\u255d    \u255a\u2550\u255d  \u255a\u2550\u2550\u2550\u2550\u255d \u255a\u2550\u255d
  Docker &amp; Podman TUI  v0.2.0\
        </pre>
      </div>
      <div v-for="(sec, i) in sections" :key="i" class="tui-help-section">
        <div class="tui-help-section-title">{{ sec.title }}</div>
        <div v-for="(item, j) in sec.items" :key="j" class="tui-help-row">
          <span class="tui-help-keys">{{ item.keys }}</span>
          <span class="tui-help-desc">{{ item.desc }}</span>
        </div>
      </div>
      <div class="tui-help-footer tui-dim">Press '?' or Esc to close help</div>
    </div>
  </TuiFrame>
</template>

<style scoped>
.tui-help {
  padding: 8px;
}

.tui-help-logo {
  text-align: center;
  margin-bottom: 16px;
}

.tui-help-logo-text {
  color: var(--tui-cyan);
  line-height: 1.2;
  font-size: 13px;
}

.tui-help-section-title {
  color: var(--tui-header-fg);
  font-weight: bold;
  margin: 8px 0 4px;
}

.tui-help-row {
  display: flex;
  gap: 12px;
  padding: 2px 0;
}

.tui-help-keys {
  color: var(--tui-white);
  width: 180px;
}

.tui-help-desc {
  color: var(--tui-gray);
}

.tui-help-footer {
  text-align: center;
  margin-top: 16px;
}
</style>
