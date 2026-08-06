<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索成果名称" allow-clear style="width: 220px" @press-enter="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="阶段" class="form-item">
            <a-select v-model="filterParams.stage" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="lab">实验室</a-option>
              <a-option value="pilot">中试</a-option>
              <a-option value="industrialized">产业化</a-option>
              <a-option value="listed">上市</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增</a-button>
        </a-space>
      </a-form>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="listData"
        :loading="loading"
        row-key="id"
        :pagination="false"
        :row-selection="rowSelection"
        @page-change="loadData"
      >
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || record.achievement_id || '-' }}</span>
        </template>
        <template #stage="{ record }">
          <a-tag :color="stageTag(record.stage)" size="small">{{ stageLabel(record.stage) }}</a-tag>
        </template>
        <template #progress="{ record }">
          <span class="cell-title">{{ record.progress || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无转化数据" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="转化详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="成果名称" :span="2">{{ currentItem.achievement_title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="当前阶段">
            <a-tag :color="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="负责人">{{ currentItem.owner || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="目标完成">{{ formatDate(currentItem.target_date) }}</a-descriptions-item>
          <a-descriptions-item label="进展记录" :span="2">{{ currentItem.progress_notes || '-' }}</a-descriptions-item>
          <a-descriptions-item label="里程碑" :span="2">{{ currentItem.milestones || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑转化' : '新增转化'" :width="560" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="标题"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="成果ID"><a-input v-model="form.achievement_id" placeholder="关联成果 ID" /></a-form-item>
        <a-form-item label="阶段">
          <a-select v-model="form.stage" style="width: 100%">
            <a-option value="lab">实验室</a-option>
            <a-option value="pilot">中试</a-option>
            <a-option value="industrialized">产业化</a-option>
            <a-option value="listed">上市</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="进度说明"><a-input v-model="form.progress" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('transformations')
const formEdit = ref(false)
const formVisible = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', achievement_id: '', stage: 'lab', progress: '' })

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialized: '产业化', listed: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'gray', pilot: 'orangered', industrialized: 'green', listed: 'arcoblue' }[s] || 'gray')
const statusTag = (s) => ({ active: 'green', completed: 'arcoblue', cancelled: 'red', ongoing: 'orangered' }[s] || 'gray')
const statusLabel = (s) => ({ active: '进行中', completed: '已完成', cancelled: '已取消', ongoing: '进行中' }[s] || (s || '-'))

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { stage: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '转化标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '当前阶段', dataIndex: 'stage', slotName: 'stage', width: 100 },
  { title: '负责人ID', dataIndex: 'owner_id', width: 110 },
  { title: '进度说明', dataIndex: 'progress', slotName: 'progress', minWidth: 160 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (row) => { Object.assign(form, row); formEdit.value = true; formVisible.value = true }
const resetForm = () => Object.assign(form, { id: '', title: '', achievement_title: '', stage: 'lab', target_date: '', description: '' })
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入标题'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false; loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除转化记录',
    content: `确定删除转化记录 "${row.achievement_title}" 吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); loadData() }
      catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
