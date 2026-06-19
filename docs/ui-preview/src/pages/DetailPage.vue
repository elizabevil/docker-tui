<script lang="ts" setup>
import {ref} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface Section {
  title: string;
  collapsed: boolean;
  rows: { label: string; value: string }[]
}

const focus = ref(0)

const sections = ref<Section[]>([
  {
    title: '\u57fa\u672c\u4fe1\u606f', collapsed: false, rows: [
      {label: 'ID', value: 'sha256:aaa111bbb222ccc333ddd444eee555fff666ggg777'},
      {label: 'Tags', value: 'nginx:latest, nginx:stable, nginx:1.25'},
      {label: 'Registry', value: 'docker.io'},
      {label: 'Name', value: 'library/nginx'},
      {label: 'Tag', value: 'latest'},
      {label: 'Digest', value: 'sha256:abc123...'},
      {label: 'Created', value: '2026-01-15 10:00:00'},
      {label: 'Size', value: '187.0 MB'},
    ]
  },
  {
    title: '\u7cfb\u7edf\u4fe1\u606f', collapsed: false, rows: [
      {label: 'Architecture', value: 'amd64'},
      {label: 'OS', value: 'linux'},
      {label: 'OS Version', value: 'Debian GNU/Linux 12 (bookworm)'},
    ]
  },
  {
    title: '\u914d\u7f6e\u4fe1\u606f', collapsed: false, rows: [
      {label: 'Author', value: 'NGINX Docker Maintainers'},
      {label: 'WorkingDir', value: '/'},
      {label: 'User', value: 'nginx'},
      {label: 'Entrypoint', value: '["/docker-entrypoint.sh"]'},
      {label: 'Cmd', value: '["nginx", "-g", "daemon off;"]'},
      {label: 'ExposedPorts', value: '80/tcp, 443/tcp'},
      {label: 'Env', value: 'NGINX_VERSION=1.25.3, PATH=/usr/sbin:...'},
    ]
  },
  {
    title: '\u5b58\u50a8\u4fe1\u606f', collapsed: true, rows: [
      {label: 'Driver', value: 'overlay2'},
      {label: 'Layers', value: '12'},
    ]
  },
  {
    title: '\u6807\u7b7e', collapsed: true, rows: [
      {label: 'maintainer', value: 'NGINX Docker Maintainers'},
      {label: 'org.opencontainers.image.version', value: '1.25.3'},
    ]
  },
  {
    title: '\u5143\u6570\u636e', collapsed: true, rows: [
      {label: 'Source', value: 'docker image inspect'},
    ]
  },
  {
    title: '\u955c\u50cf\u5386\u53f2', collapsed: true, rows: [
      {label: '(collapsed)', value: ''},
    ]
  },
])

function toggleSection(i: number) {
  sections.value[i].collapsed = !sections.value[i].collapsed
}

const shortcuts = [
  {key: '1-7', desc: 'Sections'}, {key: 'Tab', desc: 'Focus'},
  {key: 'Space', desc: 'Fold/Unfold'}, {key: 'j/k', desc: 'Scroll'},
  {key: 'Esc', desc: 'Back'},
]
</script>

<template>
  <TuiFrame title="Image Detail: nginx:latest">
    <div class="tui-detail">
      <div class="tui-detail-title">Image Detail: nginx:latest <span class="tui-dim">[Esc: back]</span></div>
      <div v-for="(sec, i) in sections" :key="i" :class="{ focused: i === focus }" class="tui-detail-section">
        <div class="tui-detail-section-header" @click="toggleSection(i)">
          <span>\u2500 {{ sec.title }} \u2500</span>
          <span class="tui-dim">[{{ sec.collapsed ? '+' : '-' }}] {{ i + 1 }}/{{ sections.length }}</span>
        </div>
        <div v-if="!sec.collapsed" class="tui-detail-section-body">
          <div v-for="(row, j) in sec.rows" :key="j" class="tui-detail-row">
            <span class="tui-detail-label">{{ row.label }}:</span>
            <span class="tui-detail-value">{{ row.value }}</span>
          </div>
        </div>
        <div v-else class="tui-detail-section-body tui-dim" style="padding:4px 8px">
          (collapsed — press Space/Enter or {{ i + 1 }} to expand)
        </div>
      </div>
    </div>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>

<style scoped>
.tui-detail-section.focused {
  outline: 1px solid var(--tui-cyan);
  border-radius: 2px;
}

.tui-detail-section-header {
  display: flex;
  justify-content: space-between;
  cursor: pointer;
  padding: 4px 8px;
}

.tui-detail-section-header:hover {
  background: var(--tui-surface);
}
</style>
