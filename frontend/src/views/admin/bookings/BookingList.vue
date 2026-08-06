<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input
              v-model="filterParams.keyword"
              placeholder="搜索联系人/用途..."
              allow-clear
              style="width: 220px"
              @press-enter="onSearchSubmit"
              @clear="onSearchSubmit"
            />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear placeholder="审核状态" @change="onSearchSubmit">
              <a-option label="待审核" value="pending" />
              <a-option label="已通过" value="approved" />
              <a-option label="已驳回" value="rejected" />
              <a-option label="已完成" value="completed" />
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
        <template #time="{ record }">
          <span class="time-text">{{ formatTime(record.start_time) }} ~ {{ formatTime(record.end_time).split(' ')[1] }}</span>
        </template>
        <template #purpose="{ record }">
          <span>{{ record.purpose || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <template v-if="record.status === 'pending'">
              <a-button type="text" status="success" size="small" @click="review(record, 'approved')">通过</a-button>
              <a-button type="text" status="danger" size="small" @click="review(record, 'rejected')">驳回</a-button>
            </template>
            <template v-else>
              <a-button v-if="record.status === 'approved'" type="text" status="success" size="small" @click="review(record, 'completed')">完成</a-button>
            </template>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无预约记录" />
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
import { Message } from '@arco-design/web-vue'
import axios from '@/utils/http'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('test-sites/bookings')

const formatTime = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = { pending: '待审核', approved: '已通过', rejected: '已驳回', completed: '已完成' }
const statusTag = (s) => ({ pending: 'orange', approved: 'green', rejected: 'red', completed: 'gray' }[s] || 'gray')

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' },
})

const columns = [
  { title: '场地ID', dataIndex: 'site_id', minWidth: 180 },
  { title: '预约时间', dataIndex: 'start_time', slotName: 'time', width: 170 },
  { title: '联系人', dataIndex: 'contact_name', width: 100 },
  { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
  { title: '用途', dataIndex: 'purpose', slotName: 'purpose', minWidth: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 120, fixed: 'right' },
]

// 审核：通过 / 驳回 / 完成（专用端点）
const review = async (row, status) => {
  try {
    await axios.post(`/api/v1/admin/test-sites/bookings/${row.id}/review`, { status, note: '' })
    Message.success({ approved: '已通过', rejected: '已驳回', completed: '已完成' }[status])
    loadData()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}
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
