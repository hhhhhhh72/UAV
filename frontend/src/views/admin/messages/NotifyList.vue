<template>
  <div class="page">
    <!-- 搜索 + 发送 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索标题..." allow-clear style="width: 220px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="类型" class="form-item">
            <a-select v-model="filterParams.msg_type" placeholder="消息类型" style="width: 160px" @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="系统通知">系统通知</a-option>
              <a-option value="活动提醒">活动提醒</a-option>
              <a-option value="审核结果">审核结果</a-option>
              <a-option value="其他">其他</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="warning" style="margin-left: auto" @click="handleSend">发送通知</a-button>
          <a-button type="primary" status="success" @click="handleAdd">新增</a-button>
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
        @sorter-change="handleSortChange"
      >
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #isRead="{ record }">
          <a-tag :color="record.is_read ? 'gray' : 'orange'" size="small">{{ record.is_read ? '已读' : '未读' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无数据" />
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
    <a-modal v-model:visible="detailVisible" title="通知详情" :width="640" :footer="false" :unmount-on-close="true">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID" :span="2">{{ currentItem.id }}</a-descriptions-item>
          <a-descriptions-item label="标题" :span="2">{{ currentItem.title }}</a-descriptions-item>
          <a-descriptions-item label="消息类型">
            <a-tag :color="typeColor[currentItem.msg_type] || 'gray'" size="small">{{ currentItem.msg_type || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="接收者">{{ currentItem.to_user || '-' }}</a-descriptions-item>
          <a-descriptions-item label="发送时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="阅读时间">{{ formatDate(currentItem.read_at) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusColor[currentItem.status] || 'gray'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="内容" :span="2">
            <div style="white-space: pre-wrap; line-height: 1.6;">{{ currentItem.content || '-' }}</div>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（发送/新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑通知' : '发送通知'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="horizontal">
        <a-form-item label="消息标题" required><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="接收者"><a-input v-model="form.receiver_id" placeholder="用户 ID（必填）" /></a-form-item>
        <a-form-item label="消息内容" required><a-input v-model="form.content" type="textarea" :rows="5" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">发送</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('messages')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate()) + ' ' + p(dt.getHours()) + ':' + p(dt.getMinutes())
}

const typeColor = { '系统通知': 'orange', '活动提醒': 'green', '审核结果': 'red', '其他': 'gray' }
const statusLabel = { 'unread': '未读', 'read': '已读' }
const statusColor = { 'unread': 'orange', 'read': 'gray' }

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { msg_type: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

// Arco sorter-change → useListRequest.onSortChange（el-table 的 { prop, order } 形态）
const handleSortChange = (dataIndex, direction) => {
  onSortChange({
    prop: dataIndex,
    order: direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  })
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '接收者', dataIndex: 'receiver_id', width: 140 },
  { title: '发送时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '状态', dataIndex: 'is_read', slotName: 'isRead', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', receiver_id: '', content: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', msg_type: '系统通知', to_user: '', content: '', status: 'unread' })
const handleSend = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, r); formEdit.value = true; formVisible.value = true }

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '发送失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入消息标题'); return }
  if (!form.receiver_id) { Message.warning('请输入接收者用户 ID'); return }
  formLoading.value = true
  try {
    const p = { sender_id: 'system', receiver_id: form.receiver_id, title: form.title, content: form.content }
    await api.create(p)
    Message.success('发送成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除该通知吗？',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); loadData() } catch { Message.error('删除失败') }
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

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
