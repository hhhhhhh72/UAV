<template>
  <div class="crud-page">
    <!-- 搜索卡 -->
    <a-card v-if="searchFields.length || $slots.search || creatable" :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item v-for="f in searchFields" :key="f.key" :label="f.label" class="form-item">
            <a-input
              v-if="!f.type || f.type === 'input'"
              v-model="filterParams[f.key]"
              :placeholder="f.placeholder || '请输入'"
              allow-clear
              :style="{ width: (f.width || 200) + 'px' }"
              @press-enter="onSearchSubmit"
              @clear="onSearchSubmit"
            />
            <a-select
              v-else-if="f.type === 'select'"
              v-model="filterParams[f.key]"
              :style="{ width: (f.width || 140) + 'px' }"
              allow-clear
              @change="onSearchSubmit"
            >
              <a-option v-for="o in f.options" :key="o.value" :value="o.value">{{ o.label }}</a-option>
            </a-select>
            <a-date-picker
              v-else-if="f.type === 'date'"
              v-model="filterParams[f.key]"
              value-format="YYYY-MM-DD"
              :style="{ width: (f.width || 140) + 'px' }"
              @change="onSearchSubmit"
            />
            <a-range-picker
              v-else-if="f.type === 'range'"
              v-model="rangeParams[f.key]"
              value-format="YYYY-MM-DD"
              :style="{ width: (f.width || 260) + 'px' }"
              @change="onSearchSubmit"
            />
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button @click="loadData"><template #icon><icon-refresh /></template>刷新</a-button>
          <slot name="search-extra" />
          <a-button v-if="creatable" class="crud-add-btn" type="primary" status="success" @click="$emit('add')">
            <template #icon><icon-plus /></template>{{ addLabel }}
          </a-button>
        </a-space>
      </a-form>
      <!-- 批量操作（选中时显示） -->
      <div v-if="selectedIds.length && (batchDelete || batchActions.length || $slots.batch)" class="batch-bar">
        <span class="batch-info">已选择 <b>{{ selectedIds.length }}</b> 项</span>
        <slot name="batch" :selected="selectedIds" :rows="selectedRows" />
        <a-button
          v-for="act in batchActions"
          :key="act.key"
          type="primary"
          :status="act.status || 'normal'"
          size="small"
          @click="handleBatchAction(act)"
        >{{ act.label }}</a-button>
        <a-button v-if="batchDelete" type="primary" status="danger" size="small" @click="handleBatchDelete">
          <template #icon><icon-delete /></template>批量删除
        </a-button>
      </div>
    </a-card>

    <!-- 表格卡 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="listData"
        :loading="loading"
        :row-key="rowKey"
        :pagination="false"
        :row-selection="selectable ? rowSelection : undefined"
        @selection-change="onSelectionChange"
        @sorter-change="onSorterChange"
      >
        <!-- 仅透传列插槽（columns 中声明了 slotName 的），避免把 search-extra/batch 等传给表格 -->
        <template v-for="col in slotColumns" :key="col.slotName" #[col.slotName]="scope">
          <slot :name="col.slotName" v-bind="scope" />
        </template>
        <template #empty>
          <slot name="empty"><a-empty description="暂无数据" /></slot>
        </template>
      </a-table>

      <div v-if="total > 0" class="pagination-wrap">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </a-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, h } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const props = defineProps({
  resource: { type: String, required: true },      // API 资源名，如 'study-tours'
  columns: { type: Array, required: true },          // a-table columns（含 slotName）
  searchFields: { type: Array, default: () => [] },  // [{key,label,type:'input'|'select'|'date'|'range',options,width,placeholder,startKey,endKey}]
  rowKey: { type: String, default: 'id' },
  selectable: { type: Boolean, default: true },
  creatable: { type: Boolean, default: false },
  batchDelete: { type: Boolean, default: true },
  // 批量动作配置：[{ key, label, status: 'success'|'danger'|'warning', confirm, api: (row) => Promise }]
  batchActions: { type: Array, default: () => [] },
  addLabel: { type: String, default: '新增' },
  defaultParams: { type: Object, default: () => ({}) },
  // 可选：自定义列表请求 (params) => Promise<{ data, total }>，覆盖 resource 的默认 CRUD list（如 JSON 文件后端）
  apiFunction: { type: Function, default: null }
})

const emit = defineEmits(['add', 'sorter-change', 'loaded'])

const api = props.apiFunction ? null : useAdminApi(props.resource)

// 需要透传给 a-table 的列插槽（columns 里声明了 slotName 的）
const slotColumns = computed(() => (props.columns || []).filter(c => c.slotName))

const { listData, loading, total, selectedIds, filterParams, loadData: fetchData, onSearchSubmit: hookSearch, resetParams: hookReset, onSortChange } = useListRequest({
  apiFunction: props.apiFunction || (api ? api.list : null),
  idKey: props.rowKey,
  defaultParams: props.defaultParams
})

// 日期范围搜索字段：值存 rangeParams，提交时合并为 startKey/endKey（默认 start_date/end_date）
const rangeParams = reactive({})

const mergeRangeParams = () => {
  for (const f of props.searchFields || []) {
    if (f.type !== 'range') continue
    const v = rangeParams[f.key]
    const startKey = f.startKey || 'start_date'
    const endKey = f.endKey || 'end_date'
    if (Array.isArray(v) && v.length === 2) {
      filterParams[startKey] = v[0]
      filterParams[endKey] = v[1]
    } else {
      delete filterParams[startKey]
      delete filterParams[endKey]
    }
  }
}

const onSearchSubmit = () => {
  mergeRangeParams()
  hookSearch()
}

const resetParams = () => {
  for (const f of props.searchFields || []) {
    if (f.type !== 'range') continue
    rangeParams[f.key] = null
    delete filterParams[f.startKey || 'start_date']
    delete filterParams[f.endKey || 'end_date']
  }
  hookReset()
}

// 加载数据（包装：通知父组件 loaded，供统计条等消费当前页数据）
const loadData = async () => {
  await fetchData()
  emit('loaded', listData.value, total.value)
}

// a-table 排序事件透传（页面可监听 @sorter-change 适配 useListRequest 的 order 语义）
const onSorterChange = (dataIndex, direction) => {
  emit('sorter-change', dataIndex, direction)
}

// 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value
}))

// Arco 通过 @selection-change 事件通知选择变化（rowSelection 里的 onChange 不会被调用）
const onSelectionChange = (keys) => {
  selectedIds.value = [...keys]
}

// 选中行数据（供批量操作插槽用）
const selectedRows = computed(() =>
  (listData.value || []).filter(r => selectedIds.value.includes(r[props.rowKey]))
)

// 切换每页条数：回到第 1 页并重新加载（Arco 的 change 事件只在页码变化时触发）
const onPageSizeChange = () => {
  filterParams.page = 1
  loadData()
}

// 提取首个失败原因（后端 fail 返回 { error: { message } }，兼容 { message } 与 axios 网络错误）
const firstRejectReason = (results) => {
  const e = results.find(r => r.status === 'rejected')?.reason
  return e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '未知错误'
}

// 批量动作（配置化：批量发布/下架/通过等任意 API 调用）
// 支持 act.prompt = { title, placeholder }：先弹输入框收集一个值（如驳回理由），
// 再逐行执行 act.api(row, promptValue)——供"批量驳回必须留理由"类操作使用。
const handleBatchAction = (act) => {
  const execute = async (promptValue) => {
    // 传选中行的完整数据（后端 Update 是全字段覆盖，只传 status 会清空其他字段）
    const results = await Promise.allSettled(selectedRows.value.map(row => act.api(row, promptValue)))
    const succeeded = results.filter(r => r.status === 'fulfilled').length
    const failed = results.filter(r => r.status === 'rejected').length
    if (failed > 0) {
      Message.error(`成功 ${succeeded} 条，失败 ${failed} 条：${firstRejectReason(results)}`)
    } else {
      Message.success(`${act.label}成功`)
    }
    selectedIds.value = []
    loadData()
  }

  if (act.prompt) {
    const promptVal = ref('')
    Modal.confirm({
      title: act.prompt.title || act.label,
      content: () => h('div', { style: 'padding: 4px 0;' }, [
        h('input', {
          value: promptVal.value,
          placeholder: act.prompt.placeholder || '',
          style: 'width:100%; box-sizing:border-box; padding:6px 10px; border:1px solid var(--color-border); border-radius:4px; outline:none; font-size:14px;',
          onInput: (e) => { promptVal.value = e.target.value }
        })
      ]),
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        const v = (promptVal.value || '').trim()
        if (!v) { Message.warning('请填写内容'); return false }
        return execute(v)
      }
    })
    return
  }

  Modal.confirm({
    title: act.label,
    content: act.confirm || `确定对选中的 ${selectedIds.value.length} 条数据执行「${act.label}」吗？`,
    okText: '确定',
    cancelText: '取消',
    onOk: () => execute()
  })
}

// 批量删除（通用能力）
const handleBatchDelete = () => {
  if (!api) { Message.warning('该数据源不支持批量删除'); return }
  Modal.confirm({
    title: '批量删除',
    content: `确定删除选中的 ${selectedIds.value.length} 条数据吗？删除后不可恢复`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      const results = await Promise.allSettled(selectedIds.value.map(id => api.delete(id)))
      const succeeded = results.filter(r => r.status === 'fulfilled').length
      const failed = results.filter(r => r.status === 'rejected').length
      if (failed > 0) {
        Message.error(`成功 ${succeeded} 条，失败 ${failed} 条：${firstRejectReason(results)}`)
      } else {
        Message.success('已删除 ' + selectedIds.value.length + ' 条')
      }
      selectedIds.value = []
      loadData()
    }
  })
}

defineExpose({ loadData, reload: loadData, onSortChange })

onMounted(loadData)
</script>

<style scoped>
.crud-page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

/* 搜索字段不收缩：a-space 的 flex 子项是 .arco-space-item 包裹层，须在其上禁止收缩，
   否则 form-item 被压缩导致 label 文字溢出压到右侧控件上（如"资源类型"56px 文字挤进 32px 容器） */
.search-form :deep(.arco-space-item) { flex: 0 0 auto; }
/* 覆盖 Arco 默认 .arco-col-5（20.8333% 固定宽度）：label 列按内容宽度，
   避免"资源类型"等长 label 文字溢出压到控件上 */
.search-form :deep(.arco-form-item-label-col) { flex: 0 0 auto; width: auto; }
.search-form :deep(.arco-form-item-label) { white-space: nowrap; }

/* 新增按钮靠右：a-space 会把子元素包一层 .arco-space-item，auto margin 须作用于包裹层 */
.search-form :deep(.arco-space-item:has(.crud-add-btn)) {
  margin-left: auto;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border);
}

.batch-info { font-size: 13px; color: var(--color-text-2); }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
