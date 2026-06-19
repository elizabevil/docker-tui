<script lang="ts" setup>
import {ref, computed} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiTable from '../components/TuiTable.vue'
import TuiSearchBar from '../components/TuiSearchBar.vue'
import TuiToast from '../components/TuiToast.vue'
import TuiStatusBar from '../components/TuiStatusBar.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface C {
  id: string;
  name: string;
  image: string;
  state: string;
  status: string;
  created: number;
  ports: string
}

const allRows: C[] = [
  {
    id: 'abc123def456',
    name: 'web-app',
    image: 'nginx:latest',
    state: 'running',
    status: 'Up 3 hours',
    created: 1705312800,
    ports: '0.0.0.0:80→80,0.0.0.0:443→443'
  },
  {
    id: 'bbb222ccc333',
    name: 'api-server',
    image: 'node:20-alpine',
    state: 'running',
    status: 'Up 2 hours',
    created: 1705309200,
    ports: '0.0.0.0:3000→3000'
  },
  {
    id: 'ccc333ddd444',
    name: 'redis-cache',
    image: 'redis:7.2',
    state: 'running',
    status: 'Up 3 hours',
    created: 1705312800,
    ports: '0.0.0.0:6379→6379'
  },
  {
    id: 'ddd444eee555',
    name: 'postgres-db',
    image: 'postgres:16',
    state: 'running',
    status: 'Up 2 hours',
    created: 1705309200,
    ports: '0.0.0.0:5432→5432'
  },
  {
    id: 'eee555fff666',
    name: 'old-web',
    image: 'nginx:1.25',
    state: 'exited',
    status: 'Exited (0) 2 days ago',
    created: 1704016800,
    ports: ''
  },
  {
    id: 'fff666ggg777',
    name: 'test-ctr',
    image: 'busybox:latest',
    state: 'created',
    status: '',
    created: 1705580000,
    ports: ''
  },
]

const selected = ref(0)
const filterText = ref('')
const sortBy = ref('name')
const sortAsc = ref(true)
const toast = ref({msg: '', type: 'success' as const})

const filtered = computed(() => {
  let r = allRows
  if (filterText.value) {
    const q = filterText.value.toLowerCase()
    r = r.filter(c => Object.values(c).some(v => String(v).toLowerCase().includes(q)))
  }
  const sorted = [...r].sort((a: any, b: any) => {
    const av = a[sortBy.value], bv = b[sortBy.value]
    const cmp = typeof av === 'number' ? av - bv : String(av).localeCompare(String(bv))
    return sortAsc.value ? cmp : -cmp
  })
  if (selected.value >= sorted.length) selected.value = Math.max(0, sorted.length - 1)
  return sorted
})

function toggleSort(col: string) {
  if (sortBy.value === col) sortAsc.value = !sortAsc.value
  else {
    sortBy.value = col;
    sortAsc.value = true
  }
}

function showToast(msg: string, t: 'success' | 'error' = 'success') {
  toast.value = {msg, type: t}
}

const columns = [
  {key: 'id', label: 'ID', sortable: true, flex: 1.2},
  {key: 'name', label: 'NAME', sortable: true, flex: 2},
  {key: 'image', label: 'IMAGE', sortable: true, flex: 2},
  {key: 'state', label: 'STATE', sortable: true, flex: 1},
  {key: 'status', label: 'STATUS', flex: 2.5},
  {key: 'ports', label: 'PORTS', flex: 2},
  {key: 'created', label: 'CREATED', sortable: true, flex: 1.8},
]

function stateDot(s: string) {
  return s === 'running' ? '●' : '○'
}

function fmt(ts: number) {
  return new Date(ts * 1000).toISOString().replace('T', ' ').slice(0, 19)
}

function shortPorts(p: string) {
  if (!p) return '—';
  const a = p.split(',');
  return a[0].replace('/tcp', '') + (a.length > 1 ? ',+' : '')
}

const shortcuts = [
  {key: 'j/k', desc: 'Up/Down'}, {key: 'Tab', desc: 'Panel'}, {key: 'Space', desc: 'Mark'},
  {key: 's', desc: 'Start'}, {key: 'S', desc: 'Stop'}, {key: 'R', desc: 'Restart'},
  {key: 'l', desc: 'Logs'}, {key: 'd', desc: 'Detail'}, {key: 'Ctrl+D', desc: 'Delete'},
  {key: '/', desc: 'Filter'}, {key: 'g/G', desc: 'Top/Bottom'}, {key: 'q', desc: 'Quit'},
]

function onKey(key: string) {
  if (key === 'j' || key === 'ArrowDown') selected.value = Math.min(selected.value + 1, filtered.value.length - 1)
  else if (key === 'k' || key === 'ArrowUp') selected.value = Math.max(0, selected.value - 1)
  else if (key === 's') showToast('Starting ' + filtered.value[selected.value]?.name + '...')
  else if (key === 'S') showToast('Stopping...')
  else if (key === 'R') showToast('Restarting...')
  else if (key === 'l') showToast('Opening logs...')
  else if (key === 'd') showToast('Detail view')
  else if (key === 'g') selected.value = 0
  else if (key === 'G') selected.value = filtered.value.length - 1
}

defineExpose({onKey})
</script>
<template>
  <TuiFrame title="Containers">
    <TuiToast v-if="toast.msg" :duration="2000" :message="toast.msg" :type="toast.type" @dismiss="toast.msg=''"/>
    <TuiSearchBar v-model="filterText" placeholder="Filter containers..."/>
    <div class="tui-panel">
      <div class="tui-panel-title">Containers <span class="tui-dim">│ {{ filtered.length }} items</span></div>
      <TuiTable :columns :rows="filtered" :selected-idx="selected" :sort-asc :sort-by
                @enter="showToast('Enter: detail view')" @select="(i:number)=>selected=i" @sort="toggleSort"/>
      <div class="tui-footer tui-dim">{{ filtered.length }} containers</div>
    </div>
    <TuiStatusBar text="● podman │ 6 containers"/>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>
