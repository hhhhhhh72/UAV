<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索难题标题" allow-clear style="width: 220px" @press-enter="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="published">征集中</a-option>
              <a-option value="closed">已关闭</a-option>
              <a-option value="resolved">已解决</a-option>
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
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #field="{ record }">
          <a-tag size="small">{{ record.field || '-' }}</a-tag>
        </template>
        <template #reward="{ record }">
          <span class="cell-amount">{{ formatMoney(record.reward_fen) }}</span>
        </template>
        <template #deadline="{ record }">
          <span class="time-text">{{ formatDate(record.deadline) }}</span>
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
          <a-empty description="暂无难题数据" />
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
    <a-modal v-model:visible="detailVisible" title="难题详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="难题标题" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发布企业">{{ currentItem.company || '-' }}</a-descriptions-item>
          <a-descriptions-item label="领域">{{ currentItem.field || '-' }}</a-descriptions-item>
          <a-descriptions-item label="悬赏金额">{{ formatMoney(currentItem.reward_fen) }}</a-descriptions-item>
          <a-descriptions-item label="截止日期">{{ formatDate(currentItem.deadline) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="难题描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="技术要求" :span="2">{{ currentItem.requirements || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑难题' : '新增难题'" :width="560" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="难题名称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="预算(分)"><a-input-number v-model="form.budget_fen" :min="0" style="width: 100%" /></a-form-item>
        <a-form-item label="截止日期"><a-date-picker v-model="form.deadline" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="published">征集中</a-option>
            <a-option value="closed">已关闭</a-option>
            <a-option value="resolved">已解决</a-option>
          </a-select>
        </a-form-item>
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

const api = useAdminApi('rd-challenges')

const statusLabel = (s) => ({ open: '征集中', closed: '已关闭', resolved: '已解决' }[s] || s || '-')
const statusTag = (s) => ({ open: 'orangered', closed: 'gray', resolved: 'green' }[s] || 'gray')

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { status: '' }
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
  { title: '难题标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '发布企业ID', dataIndex: 'poster_id', minWidth: 140 },
  { title: '领域', dataIndex: 'field', slotName: 'field', width: 120 },
  { title: '悬赏金额', dataIndex: 'reward_fen', slotName: 'reward', width: 130, align: 'right' },
  { title: '截止日期', dataIndex: 'deadline', slotName: 'deadline', width: 130 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', field: '', budget_fen: 0, deadline: '', status: 'published', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', company: '', field: '', reward_fen: 0, deadline: '', status: 'open', description: '', requirements: '' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, { id: r.id, title: r.title || '', field: r.field || '', budget_fen: r.budget_fen || 0, deadline: r.deadline || '', status: r.status || 'published', description: r.description || '' }); formEdit.value = true; formVisible.value = true }
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入难题标题'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false; loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除难题',
    content: `确定删除难题"${r.title}"吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); loadData() }
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

.cell-amount { color: #E96012; font-weight: 500; }

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
