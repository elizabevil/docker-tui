<script lang="ts" setup>
import {computed} from 'vue'

export interface TuiCol {
  key: string;
  label: string;
  sortable?: boolean;
  flex?: number;
  minWidth?: number
}

const props = defineProps<{
  columns: TuiCol[]; rows: Record<string, any>[]; selectedIdx?: number
  sortBy?: string; sortAsc?: boolean
}>()

const emit = defineEmits<{
  select: [idx: number];
  sort: [col: string];
  enter: [row: Record<string, any>, idx: number]
}>()

const columnsFlex = computed(() => {
  const hasFlex = props.columns.some(c => c.flex !== undefined)
  return hasFlex ? props.columns : props.columns.map(c => ({...c, flex: 1}))
})
</script>

<template>
  <div class="tui-table-wrap">
    <div class="tui-tr tui-th">
      <div v-for="col in columnsFlex" :key="col.key" :class="{ sortable: col.sortable }"
           :data-sort="sortBy === col.key ? (sortAsc ? 'asc' : 'desc') : ''"
           :style="{ flex: `${col.flex ?? 1} 0 0%`, minWidth: col.minWidth ? col.minWidth + 'px' : 0 }"
           class="tui-td"
           @click="col.sortable && emit('sort', col.key)">
        <span class="tui-td-inner">{{ col.label }}</span>
      </div>
    </div>
    <div v-for="(row, idx) in rows" :key="idx" :class="{ selected: idx === selectedIdx, alt: idx % 2 === 1 }"
         class="tui-tr"
         @click="emit('select', idx)" @dblclick="emit('enter', row, idx)">
      <div v-for="col in columnsFlex" :key="col.key" :style="{ flex: `${col.flex ?? 1} 0 0%`, minWidth: col.minWidth ? col.minWidth + 'px' : 0 }"
           class="tui-td">
        <span class="tui-td-inner">
          <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">{{ row[col.key] }}</slot>
        </span>
      </div>
    </div>
    <div v-if="rows.length === 0" class="tui-table-empty tui-dim">(empty)</div>
  </div>
</template>

<style scoped>
.tui-table-wrap {
  display: flex;
  flex-direction: column;
  font-size: 13px;
  outline: none
}

.tui-tr {
  display: flex;
  border-bottom: 1px solid var(--tui-surface)
}

.tui-tr.alt {
  background: rgba(128, 128, 128, 0.04)
}

.tui-tr.selected {
  background: rgba(66, 165, 245, 0.12);
  color: var(--tui-blue);
  font-weight: bold
}

.tui-tr:hover:not(.tui-th) {
  background: rgba(66, 165, 245, 0.06)
}

.tui-th {
  color: var(--tui-cyan);
  font-weight: bold;
  border-bottom: 1px solid var(--tui-cyan)
}

.tui-td {
  padding: 2px 4px;
  overflow: hidden;
  display: flex;
  align-items: center
}

.tui-td-inner {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%
}

.tui-th .tui-td {
  padding: 3px 4px
}

.tui-th .tui-td.sortable {
  cursor: pointer
}

/* Sort arrow via ::after — fixed width, doesn't affect layout */
.tui-th .tui-td[data-sort="asc"] .tui-td-inner::after {
  content: ' ▲';
  font-size: 10px;
  margin-left: 2px
}

.tui-th .tui-td[data-sort="desc"] .tui-td-inner::after {
  content: ' ▼';
  font-size: 10px;
  margin-left: 2px
}

.tui-table-empty {
  padding: 16px;
  text-align: center
}
</style>
