<script lang="ts" setup>
import {ref} from 'vue'
import {useRouter, useRoute} from 'vue-router'

const router = useRouter()
const route = useRoute()
const themes = ['default', 'nord', 'dracula']
const currentTheme = ref('default')

function setTheme(name: string) {
  currentTheme.value = name
  document.documentElement.setAttribute('data-theme', name)
}

const pages = [
  {path: '/containers', label: 'Containers'},
  {path: '/images', label: 'Images'},
  {path: '/volumes', label: 'Volumes'},
  {path: '/networks', label: 'Networks'},
  {path: '/detail', label: 'Detail'},
  {path: '/logs', label: 'Logs'},
  {path: '/help', label: 'Help'},
  {path: '/catalog', label: 'Catalog'},
]

function onGlobalKey(e: KeyboardEvent) {
  const t = e.target as HTMLElement
  if (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA') return
  if (e.key === 'Tab') {
    e.preventDefault()
    const idx = pages.findIndex(p => p.path === route.path)
    const next = (idx + (e.shiftKey ? -1 : 1) + pages.length) % pages.length
    router.push(pages[next].path)
  }
  if (e.key === '?') router.push('/help')
  if (e.key === 'q' && confirm('Quit preview?')) window.close()
}
</script>
<template>
  <div class="tui-app" tabindex="0" @keydown="onGlobalKey">
    <nav class="tui-sidebar">
      <div class="tui-sidebar-title">dtui preview</div>
      <div class="tui-sidebar-nav">
        <div class="tui-sidebar-section">Pages</div>
        <router-link v-for="p in pages" :key="p.path" :class="{ active: route.path === p.path }"
                     :to="p.path">{{ p.label }}
        </router-link>
        <div class="tui-sidebar-section">Settings</div>
        <div class="tui-sidebar-row">
          <select v-model="currentTheme" class="tui-select" @change="setTheme(currentTheme)">
            <option v-for="t in themes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
      </div>
      <div class="tui-sidebar-footer tui-dim">Tab: panel │ ?: help │ q: quit</div>
    </nav>
    <div class="tui-main">
      <div class="tui-preview-area" @keydown="onGlobalKey">
        <router-view/>
      </div>
    </div>
  </div>
</template>
<style scoped>
.tui-app {
  display: flex;
  height: 100vh;
  background: var(--tui-bg);
  color: var(--tui-white);
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  outline: none
}

.tui-sidebar {
  width: 200px;
  background: var(--tui-dark);
  border-right: 1px solid var(--tui-surface);
  padding: 12px;
  display: flex;
  flex-direction: column;
  flex-shrink: 0
}

.tui-sidebar-title {
  color: var(--tui-cyan);
  font-weight: bold;
  font-size: 14px;
  margin-bottom: 16px
}

.tui-sidebar-nav {
  flex: 1
}

.tui-sidebar-section {
  color: var(--tui-gray);
  font-size: 11px;
  text-transform: uppercase;
  margin: 12px 0 4px;
  letter-spacing: 1px
}

.tui-sidebar-nav a {
  display: block;
  color: var(--tui-white);
  text-decoration: none;
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 13px
}

.tui-sidebar-nav a.active, .tui-sidebar-nav a:hover {
  background: var(--tui-surface);
  color: var(--tui-cyan)
}

.tui-sidebar-row {
  padding: 4px 8px
}

.tui-sidebar-footer {
  font-size: 11px;
  padding: 8px 0;
  text-align: center
}

.tui-select {
  background: var(--tui-surface);
  color: var(--tui-white);
  border: 1px solid var(--tui-gray);
  padding: 2px 4px;
  width: 100%;
  font-size: 12px
}

.tui-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden
}

.tui-preview-area {
  flex: 1;
  padding: 8px;
  overflow: auto
}
</style>
