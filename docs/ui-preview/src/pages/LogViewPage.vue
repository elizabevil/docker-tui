<script lang="ts" setup>
import {ref} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface LogLine {
  num: number;
  ts: string;
  stream: string;
  text: string
}

const logLines = ref<LogLine[]>([
  {num: 1, ts: '2026-01-15T10:00:00.123Z', stream: 'stdout', text: 'GET /index.html 200 OK (0.023s)'},
  {num: 2, ts: '2026-01-15T10:00:01.456Z', stream: 'stdout', text: 'POST /api/data 201 Created (0.045s)'},
  {num: 3, ts: '2026-01-15T10:00:02.789Z', stream: 'stderr', text: 'connection refused to backend:5432'},
  {num: 4, ts: '2026-01-15T10:00:03.012Z', stream: 'stdout', text: 'GET /healthz 200 OK (0.001s)'},
  {num: 5, ts: '2026-01-15T10:00:04.345Z', stream: 'stdout', text: 'PUT /api/users 201 Created (0.032s)'},
  {
    num: 6,
    ts: '2026-01-15T10:00:05.678Z',
    stream: 'stderr',
    text: 'slow query detected: SELECT * FROM logs WHERE ts > NOW() - 1h'
  },
  {num: 7, ts: '2026-01-15T10:00:06.901Z', stream: 'stdout', text: 'GET /static/app.js 200 OK (0.011s)'},
  {num: 8, ts: '2026-01-15T10:00:07.234Z', stream: 'stdout', text: 'DELETE /api/sessions 204 No Content'},
  {num: 9, ts: '2026-01-15T10:00:08.567Z', stream: 'stdout', text: 'POST /api/orders 201 Created (0.089s)'},
  {num: 10, ts: '2026-01-15T10:00:09.890Z', stream: 'stderr', text: 'disk usage warning: 85%'},
])

const scrollPos = ref(0)
const lineHeight = 22
const maxVisible = 15

const visibleLines = logLines.value.slice(scrollPos.value, scrollPos.value + maxVisible)

const shortcuts = [
  {key: 'Esc', desc: 'Back'}, {key: 'j/k', desc: 'Scroll'},
  {key: 'PgUp/Dn', desc: 'Page'}, {key: 'g/G', desc: 'Top/Bottom'},
]
</script>

<template>
  <TuiFrame title="Logs: nginx-container">
    <div class="tui-panel">
      <div class="tui-panel-title">Logs: abc123def456 <span class="tui-dim">\u2502 {{ logLines.length }} lines (auto-refresh 2s)</span>
      </div>
      <div class="tui-log-container">
        <div v-for="line in visibleLines" :key="line.num" class="tui-log-line">
          <span class="tui-log-num">{{ String(line.num).padStart(4) }}</span>
          <span class="tui-log-ts">{{ line.ts }}</span>
          <span :class="['tui-log-stream', line.stream === 'stderr' ? 'stderr' : 'stdout']">{{ line.stream }}</span>
          <span :class="['tui-log-text', line.stream === 'stderr' ? 'stderr' : '']">{{ line.text }}</span>
        </div>
      </div>
      <div class="tui-footer">
        <span>{{ scrollPos + 1 }}-{{ Math.min(scrollPos + maxVisible, logLines.length) }}/{{ logLines.length }} \u2502 j/k scroll \u2502 Esc back</span>
      </div>
    </div>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>

<style scoped>
.tui-log-container {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 12px;
  line-height: 22px;
}

.tui-log-num {
  color: var(--tui-gray);
  width: 32px;
  display: inline-block;
  text-align: right;
  margin-right: 8px;
}

.tui-log-ts {
  color: var(--tui-cyan);
  margin-right: 8px;
}

.tui-log-stream {
  width: 48px;
  display: inline-block;
  margin-right: 8px;
}

.tui-log-stream.stderr {
  color: var(--tui-red);
}

.tui-log-text.stderr {
  color: var(--tui-red);
}
</style>
