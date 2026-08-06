<template>
  <div class="page">
    <!-- 搜索过滤区 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 130px" allow-clear @change="onSearchSubmit">
              <a-option label="全部状态" value="" />
              <a-option label="待审核" value="pending" />
              <a-option label="已通过" value="approved" />
              <a-option label="已拒绝" value="rejected" />
            </a-select>
          </a-form-item>
          <a-form-item label="板块" class="form-item">
            <a-select v-model="filterParams.section" style="width: 130px" allow-clear @change="onSearchSubmit">
              <a-option label="全部板块" value="" />
              <a-option label="研学" value="yanxue" />
              <a-option label="无人机销售" value="sale" />
              <a-option label="乐园" value="park" />
            </a-select>
          </a-form-item>
          <a-form-item label="关键词" class="form-item">
            <a-input
              v-model="filterParams.keyword"
              placeholder="搜索评价内容..."
              allow-clear
              style="width: 200px"
              @press-enter="onSearchSubmit"
              @clear="onSearchSubmit"
            />
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
        @sorter-change="handleSorterChange"
      >
        <template #target="{ record }">
          <a-tag :color="sectionTagColor(record.target_type)" size="small">{{ targetTypeLabel(record.target_type) }}</a-tag>
          <span class="target-id">{{ record.target_id || '-' }}</span>
        </template>
        <template #rating="{ record }">
          <span class="stars">{{ '★'.repeat(record.rating || 0) }}{{ '☆'.repeat(5 - (record.rating || 0)) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTagColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <template v-if="record.status === 'pending'">
              <a-button type="text" status="success" size="small" @click="handleStatus(record, 'approved')">通过</a-button>
              <a-button type="text" status="warning" size="small" @click="handleStatus(record, 'rejected')">拒绝</a-button>
            </template>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无评价数据" />
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
import { useListRequest } from '@/hooks/useListRequest'
import { getReviewList, updateReviewStatus, deleteReview } from '@/api/admin/review'

const targetTypeLabel = (t) => ({ demand: '需求', product: '商品', shop: '商家', job: '职位', course: '课程', venue: '场地' }[t] || t || '通用')
const statusLabel = (s) => ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || s)
const sectionTagColor = (t) => ({ demand: 'arcoblue', product: 'green', shop: 'orange', job: 'arcoblue', course: 'arcoblue', venue: 'gray' }[t] || 'gray')
const statusTagColor = (s) => ({ pending: 'orange', approved: 'green', rejected: 'red' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: getReviewList,
  idKey: 'id',
  defaultParams: { status: '', section: '', limit: 20 }
})

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  onSortChange({ prop: dataIndex, order })
}

const columns = [
  { title: '评价人ID', dataIndex: 'reviewer_id', width: 140, tooltip: true },
  { title: '评价对象', dataIndex: 'target_id', slotName: 'target', minWidth: 160 },
  { title: '评分', dataIndex: 'rating', slotName: 'rating', width: 100 },
  { title: '评价内容', dataIndex: 'content', minWidth: 200, tooltip: true },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '评价时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160, sortable: true },
  { title: '操作', slotName: 'actions', width: 180, fixed: 'right' },
]

const handleStatus = async (item, status) => {
  try {
    await updateReviewStatus(item.id, status)
    item.status = status
    Message.success(status === 'approved' ? '已通过' : '已拒绝')
  } catch (e) { Message.error('操作失败') }
}

const handleDelete = (item) => {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteReview(item.id)
        listData.value = listData.value.filter(r => r.id !== item.id)
        Message.success('已删除')
      } catch (e) { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.target-id { margin-left: 6px; color: #86909C; }

.stars { color: #ffd21e; font-size: 14px; letter-spacing: 1px; }
.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
