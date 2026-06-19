<script lang="ts" setup>
import {ref, computed} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiTable from '../components/TuiTable.vue'
import TuiSearchBar from '../components/TuiSearchBar.vue'
import TuiToast from '../components/TuiToast.vue'
import TuiStatusBar from '../components/TuiStatusBar.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface I {
  id: string;
  name: string;
  tag: string;
  registry: string;
  arch: string;
  created: number;
  size: number;
  hasCtr: boolean
}

const allRows: I[] = [
  {
    id: 'sha256:aaa111',
    name: 'nginx',
    tag: 'latest',
    registry: 'docker.io',
    arch: 'amd64',
    created: 1705312800,
    size: 187_000_000,
    hasCtr: true
  },
  {
    id: 'sha256:bbb222',
    name: 'node',
    tag: '20-alpine',
    registry: 'docker.io',
    arch: 'amd64',
    created: 1704794400,
    size: 126_000_000,
    hasCtr: false
  },
  {
    id: 'sha256:ccc333',
    name: 'redis',
    tag: '7.2',
    registry: 'docker.io',
    arch: 'amd64',
    created: 1704535200,
    size: 41_200_000,
    hasCtr: true
  },
  {
    id: 'sha256:ddd444',
    name: 'postgres',
    tag: '16',
    registry: 'docker.io',
    arch: 'amd64',
    created: 1704016800,
    size: 412_000_000,
    hasCtr: true
  },
  {
    id: 'sha256:eee555',
    name: 'alertmanager',
    tag: 'v0.27',
    registry: 'quay.io',
    arch: 'amd64',
    created: 1703500000,
    size: 82_000_000,
    hasCtr: false
  },
  {
    id: 'sha256:fff666',
    name: 'loki',
    tag: '2.9',
    registry: 'ghcr.io',
    arch: 'arm64',
    created: 1702981600,
    size: 95_000_000,
    hasCtr: false
  },
]
const selected = ref(0);
const filterText = ref('');
const sortBy = ref('name');
const sortAsc = ref(true)
const toast = ref({msg: '', type: 'success' as const})
const filtered = computed(() => {
  let r = allRows
  if (filterText.value) {
    const q = filterText.value.toLowerCase();
    r = r.filter(i => Object.values(i).some(v => String(v).toLowerCase().includes(q)))
  }
  const s = [...r].sort((a: any, b: any) => {
    const cmp = a[sortBy.value] - b[sortBy.value];
    return sortAsc.value ? cmp : -cmp
  })
  if (selected.value >= s.length) selected.value = Math.max(0, s.length - 1);
  return s
})

function toggleSort(c: string) {
  if (sortBy.value === c) sortAsc.value = !sortAsc.value; else {
    sortBy.value = c;
    sortAsc.value = true
  }
}

function st(m: string, t: 'success' | 'error' = 'success') {
  toast.value = {msg: m, type: t}
}

const columns = [
  {key: 'arrow', label: '', flex: 0.3}, {key: 'registry', label: 'REGISTRY', sortable: true, flex: 1.5}, {
    key: 'name',
    label: 'NAME',
    sortable: true,
    flex: 2
  },
  {key: 'tag', label: 'TAG', flex: 1}, {key: 'id', label: 'ID', flex: 1.5}, {
    key: 'arch',
    label: 'ARCH',
    flex: 0.8
  }, {key: 'created', label: 'CREATED', sortable: true, flex: 1.5},
  {key: 'size', label: 'SIZE', sortable: true, flex: 1},
]

function fmtSize(b: number) {
  const u = ['B', 'KB', 'MB', 'GB'];
  let i = 0, s = b;
  while (s >= 1024 && i < 3) {
    s /= 1024;
    i++
  }
  return s.toFixed(i ? 1 : 0) + ' ' + u[i]
}

function fmt(ts: number) {
  return new Date(ts * 1000).toISOString().replace('T', ' ').slice(0, 19)
}

const shortcuts = [
  {key: 'j/k', desc: 'Up/Down'}, {key: 'Tab', desc: 'Panel'}, {key: '→/Enter', desc: 'Containers'}, {
    key: 'd',
    desc: 'Detail'
  },
  {key: 'D', desc: 'Debug'}, {key: 'E', desc: 'Export'}, {key: 'P', desc: 'Pull'}, {key: 'p', desc: 'Prune'}, {
    key: 'o',
    desc: 'Sort'
  },
  {key: 'Ctrl+D', desc: 'Delete'}, {key: '/', desc: 'Filter'}, {key: 'q', desc: 'Quit'},
]
</script>
<template>
  <TuiFrame title="Images">
    <TuiToast v-if="toast.msg" :duration="2000" :message="toast.msg" :type="toast.type" @dismiss="toast.msg=''"/>
    <TuiSearchBar v-model="filterText" placeholder="Filter images..."/>
    <div class="tui-panel">
      <div class="tui-panel-title">Images <span class="tui-dim">│ {{ filtered.length }} items</span></div>
      <TuiTable :columns :rows="filtered" :selected-idx="selected" :sort-asc :sort-by
                @enter="st('Showing containers...')" @select="(i:number)=>selected=i" @sort="toggleSort">
        <template #cell-arrow="{row}"><span :class="row.hasCtr?'':'tui-dim'">{{ row.hasCtr ? '▸' : ' ' }}</span>
        </template>
        <template #cell-id="{value}"><span class="tui-dim">{{ value.length > 12 ? value.slice(0, 12) : value }}</span>
        </template>
        <template #cell-created="{value}"><span class="tui-dim">{{ fmt(value) }}</span></template>
        <template #cell-size="{value}">{{ fmtSize(value) }}</template>
      </TuiTable>
      <div class="tui-footer tui-dim">{{ filtered.length }} images</div>
    </div>
    <TuiStatusBar text="● docker │ 6 images"/>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>
