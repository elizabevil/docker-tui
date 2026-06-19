<script lang="ts" setup>
import {ref, computed} from 'vue'
import TuiFrame from '../components/TuiFrame.vue'
import TuiTable from '../components/TuiTable.vue'
import TuiSearchBar from '../components/TuiSearchBar.vue'
import TuiToast from '../components/TuiToast.vue'
import TuiStatusBar from '../components/TuiStatusBar.vue'
import TuiShortcuts from '../components/TuiShortcuts.vue'

interface N {
  name: string;
  driver: string;
  scope: string;
  subnet: string;
  containers: number;
  id: string
}

const allRows: N[] = [
  {name: 'bridge', driver: 'bridge', scope: 'local', subnet: '172.17.0.0/16', containers: 3, id: 'abc123def456'},
  {name: 'host', driver: 'host', scope: 'local', subnet: '', containers: 0, id: 'host'},
  {name: 'none', driver: 'null', scope: 'local', subnet: '', containers: 0, id: 'none'},
  {name: 'app_net', driver: 'bridge', scope: 'local', subnet: '10.10.0.0/24', containers: 5, id: 'aaa111bbb222'},
  {name: 'db_net', driver: 'bridge', scope: 'local', subnet: '10.20.0.0/24', containers: 2, id: 'ccc333ddd444'},
  {name: 'monitor', driver: 'overlay', scope: 'swarm', subnet: '10.30.0.0/24', containers: 0, id: 'eee555fff666'},
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
    r = r.filter(n => Object.values(n).some(v => String(v).toLowerCase().includes(q)))
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
  {key: 'scope', label: 'SCOPE', flex: 0.6}, {key: 'subnet', label: 'SUBNET', flex: 2},
  {key: 'containers', label: 'CTR', sortable: true, flex: 0.5}, {key: 'id', label: 'ID', flex: 1.5},
]
const shortcuts = [
  {key: 'j/k', desc: 'Up/Down'}, {key: 'Tab', desc: 'Panel'}, {key: 'Space', desc: 'Mark'},
  {key: 'Enter', desc: 'Containers'}, {key: 'd', desc: 'Detail'}, {key: 'Ctrl+D', desc: 'Delete'},
  {key: '/', desc: 'Filter'}, {key: 'q', desc: 'Quit'},
]
</script>
<template>
  <TuiFrame title="Networks">
    <TuiToast v-if="toast.msg" :duration="2000" :message="toast.msg" @dismiss="toast.msg=''"/>
    <TuiSearchBar v-model="filterText" placeholder="Filter networks..."/>
    <div class="tui-panel">
      <div class="tui-panel-title">Networks <span class="tui-dim">│ {{ filtered.length }} items</span></div>
      <TuiTable :columns :rows="filtered" :selected-idx="selected" :sort-asc :sort-by
                @enter="toast={msg:'Containers in network...',type:'success'}" @select="(i:number)=>selected=i"
                @sort="toggleSort">
        <template #cell-subnet="{value}"><span class="tui-dim">{{ value || '—' }}</span></template>
        <template #cell-id="{value}"><span class="tui-dim">{{ value.length > 12 ? value.slice(0, 12) : value }}</span>
        </template>
        <template #cell-containers="{value}">{{ value > 0 ? value : '—' }}</template>
      </TuiTable>
      <div class="tui-footer tui-dim">{{ filtered.length }} networks</div>
    </div>
    <TuiStatusBar text="● docker │ 6 networks"/>
    <TuiShortcuts :shortcuts="shortcuts"/>
  </TuiFrame>
</template>
