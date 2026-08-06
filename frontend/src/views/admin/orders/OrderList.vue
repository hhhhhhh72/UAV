<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="日期范围" class="form-item">
            <a-range-picker
              v-model="dateRange"
              value-format="YYYY-MM-DD"
              style="width: 260px"
              @change="onSearchSubmit"
            />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 130px" allow-clear @change="onSearchSubmit">
              <a-option label="全部状态" value="" />
              <a-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="handleSearch"><template #icon><icon-search /></template>查询</a-button>
        </a-space>
      </a-form>
    </a-card>

    <!-- 交易统计条（基于当前页 + 接口 total） -->
    <a-card :bordered="false" class="stat-card">
      <div class="stats-bar">
        <div class="stat"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">订单总数</span></div>
        <div class="stat money"><span class="stat-num">¥{{ stats.amount }}</span><span class="stat-label">交易额(本页)</span></div>
        <div class="stat done"><span class="stat-num">{{ stats.completed }}</span><span class="stat-label">已完成</span></div>
        <div class="stat rate"><span class="stat-num">{{ stats.rate }}%</span><span class="stat-label">完成率</span></div>
      </div>
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
      >
        <template #amount="{ record }">
          <span>{{ ((record.amount_fen || 0) / 100).toFixed(2) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTagColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
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
    <a-modal v-model:visible="detailVisible" title="订单详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="订单号">{{ currentItem.id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagColor(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="商品 ID">{{ currentItem.product_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="金额(元)">{{ ((currentItem.amount_fen || 0) / 100).toFixed(2) }}</a-descriptions-item>
          <a-descriptions-item label="买家">{{ currentItem.buyer_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="卖家">{{ currentItem.seller_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="下单时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>

        <div class="review-actions">
          <a-divider />
          <span class="review-label">修改状态：</span>
          <a-select v-model="newStatus" style="width: 140px;">
            <a-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
          </a-select>
          <a-button type="primary" @click="onUpdateStatus">更新</a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { getOrderList, updateOrderStatus } from '@/api/admin/order'

const statusOptions = [
  { label: '待付款', value: 'pending' },
  { label: '已付款', value: 'paid' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagColor = (s) => ({ completed: 'green', shipped: 'arcoblue', paid: 'orange', pending: 'gray', cancelled: 'gray' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const dateRange = ref(null)

// 交易统计（分类/金额基于当前页；订单总数取接口 total）
const stats = computed(() => {
  const rows = listData.value || []
  const amount = rows.reduce((s, x) => s + (x.amount_fen || 0), 0) / 100
  const completed = rows.filter((x) => x.status === 'completed').length
  const rate = rows.length ? Math.round((completed / rows.length) * 100) : 0
  return { total: total.value || 0, amount: amount.toLocaleString('zh-CN', { minimumFractionDigits: 2 }), completed, rate }
})

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: getOrderList,
  idKey: 'id',
  defaultParams: { status: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

const columns = [
  { title: '订单号', dataIndex: 'id', width: 180 },
  { title: '商品 ID', dataIndex: 'product_id', minWidth: 120 },
  { title: '买家', dataIndex: 'buyer_id', width: 130 },
  { title: '卖家', dataIndex: 'seller_id', width: 130 },
  { title: '金额(元)', dataIndex: 'amount_fen', slotName: 'amount', width: 110, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 110 },
  { title: '下单时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '操作', slotName: 'actions', width: 120, fixed: 'right' },
]

// 自定义日期范围逻辑
const origOnSearchSubmit = onSearchSubmit
const handleSearch = () => {
  if (dateRange.value && dateRange.value.length === 2) {
    filterParams.startDate = dateRange.value[0]
    filterParams.endDate = dateRange.value[1]
  } else {
    delete filterParams.startDate
    delete filterParams.endDate
  }
  origOnSearchSubmit()
}

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('pending')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || 'pending'
  detailVisible.value = true
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    await updateOrderStatus(currentItem.value.id, newStatus.value)
    currentItem.value.status = newStatus.value
    Message.success('状态已更新')
    loadData()
  } catch (e) { Message.error('更新失败') }
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.stat-card { margin-bottom: 16px; }

.stats-bar { display: flex; gap: 0; }

.stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 4px 16px;
  border-right: 1px solid #EEF1F4;
}

.stat:last-child { border-right: none; }

.stat-num {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-1);
  line-height: 1.2;
}

.stat-label { font-size: 12px; color: #86909C; }

.stat.money .stat-num { color: #E96012; }
.stat.done .stat-num { color: #168A55; }
.stat.rate .stat-num { color: #165DFF; }

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
.review-label { color: #4E5969; }
</style>
