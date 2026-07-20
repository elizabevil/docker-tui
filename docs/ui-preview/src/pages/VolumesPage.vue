<script lang="ts" setup>
import {ref, computed} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiTable from '../components/TuiTable.vue'
import TuiSearchBar from '../components/TuiSearchBar.vue'
import TuiToast from '../components/TuiToast.vue'
import TuiStatusBar from '../components/TuiStatusBar.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface V {
  name: string;
  driver: string;
  scope: string;
  mountpoint: string;
  createdAt: string
}

const allRows: V[] = [
  {
    name: 'pgdata',
    driver: 'local',
    scope: 'local',
    mountpoint: '/var/lib/docker/volumes/pgdata/_data',
    createdAt: '2025-12-01 10:00:00'
  },
  {
    name: 'redis_data',
    driver: 'local',
    scope: 'local',
    mountpoint: '/var/lib/docker/volumes/redis_data/_data',
    createdAt: '2025-12-05 14:30:00'
  },
  {
    name: 'nginx_html',
    driver: 'local',
    scope: 'local',
    mountpoint: '/var/lib/docker/volumes/nginx_html/_data',
    createdAt: '2025-12-10 09:15:00'
  },
  {
    name: 'loki_data',
    driver: 'local',
    scope: 'local',
    mountpoint: '/var/lib/docker/volumes/loki_data/_data',
    createdAt: '2025-12-15 16:45:00'
  },
  {
    name: 'prometheus',
    driver: 'local',
    scope: 'local',
    mountpoint: '/var/lib/docker/volumes/prometheus/_data',
    createdAt: '2025-12-20 11:00:00'
  },
  {name: 'grafana', driver: 'nfs', scope: 'local', mountpoint: '/mnt/nfs/grafana', createdAt: '2026-01-05 08:00:00'},
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
    r = r.filter(v => Object.values(v).some(v => String(v).toLowerCase().includes(q)))
  }
  const s = [...r].sort((a: any, b: any) => {
    const cmp = String(a[sortBy.value]).localeCompare(String(b[sortBy.value]));
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

const columns = [
  {key: 'name', label: 'NAME', sortable: true, flex: 1.5}, {key: 'driver', label: 'DRIVER', sortable: true, flex: 0.8},
  {key: 'scope', label: 'SCOPE', sortable: true, flex: 0.6}, {key: 'mountpoint', label: 'MOUNTPOINT', flex: 3},
  {key: 'createdAt', label: 'CREATED', sortable: true, flex: 1.5},
]
const shortcuts = [
  {key: 'j/k', desc: 'Up/Down'}, {key: 'Tab', desc: 'Panel'}, {key: 'Space', desc: 'Mark'},
  {key: 'Enter', desc: 'Containers'}, {key: 'd', desc: 'Detail'}, {key: 'Ctrl+D', desc: 'Delete'},
  {key: '/', desc: 'Filter'}, {key: 'q', desc: 'Quit'},
]
</script>
<template>
  <TuiFrame title="Volumes">
    <TuiToast v-if="toast.msg" :duration="2000" :message="toast.msg" @dismiss="toast.msg=''"/>
    <TuiSearchBar v-model="filterText" placeholder="Filter volumes..."/>
    <div class="tui-panel">
      <div class="tui-panel-title">Volumes <span class="tui-dim">│ {{ filtered.length }} items</span></div>
      <TuiTable :columns :rows="filtered" :selected-idx="selected" :sort-asc :sort-by
                @enter="toast={msg:'Containers using volume...',type:'success'}" @select="(i:number)=>selected=i"
                @sort="toggleSort">
        <template #cell-mountpoint="{value}"><span class="tui-dim">{{ value }}</span></template>
      </TuiTable>
      <div class="tui-footer tui-dim">{{ filtered.length }} volumes</div>
    </div>
    <TuiStatusBar text="● docker │ 6 volumes"/>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>
