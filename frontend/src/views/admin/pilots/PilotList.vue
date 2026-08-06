<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索姓名..." allow-clear style="width: 200px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" placeholder="审核状态" allow-clear @change="onSearchSubmit">
              <a-option value="pending">待审核</a-option>
              <a-option value="approved">已认证</a-option>
              <a-option value="rejected">已驳回</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
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
      >
        <template #certCount="{ record }">
          <span>{{ (record.cert_ids || []).length }} 项</span>
        </template>
        <template #bio="{ record }">
          <span>{{ record.bio || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <template v-if="record.status === 'pending'">
              <a-button type="text" status="success" size="small" @click="handleApprove(record)">通过</a-button>
              <a-button type="text" status="danger" size="small" @click="handleReject(record)">驳回</a-button>
            </template>
            <template v-else>
              <a-button v-if="record.status === 'approved'" type="text" status="danger" size="small" @click="handleReject(record)">撤销</a-button>
              <a-button v-if="record.status === 'rejected'" type="text" status="success" size="small" @click="handleApprove(record)">恢复通过</a-button>
            </template>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无飞手申请" />
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
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('certified-pilots')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = { pending: '待审核', approved: '已认证', rejected: '已驳回' }
const statusTag = (s) => ({ pending: 'orangered', approved: 'green', rejected: 'red' }[s] || 'gray')

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' },
})

const columns = [
  { title: '姓名', dataIndex: 'real_name', minWidth: 100 },
  { title: '身份证号', dataIndex: 'id_card', width: 200 },
  { title: '证书', dataIndex: 'cert_ids', slotName: 'certCount', width: 80, align: 'center' },
  { title: '时长(h)', dataIndex: 'flight_hours', width: 80, align: 'center' },
  { title: '擅长领域', dataIndex: 'bio', slotName: 'bio', minWidth: 140 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '申请时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

// 审核：通过 / 驳回（专用端点）
const setStatus = async (row, action, tip) => {
  try {
    await axios.post(`/api/v1/admin/certified-pilots/${row.id}/${action}`)
    Message.success(tip)
    loadData()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}
const handleApprove = (row) => setStatus(row, 'approve', '已通过，飞手进入公开名录')
const handleReject = (row) => {
  Modal.confirm({
    title: '驳回申请',
    content: `确定驳回 ${row.real_name} 的飞手认证申请？`,
    okText: '驳回',
    cancelText: '取消',
    onOk: () => setStatus(row, 'reject', '已驳回')
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
