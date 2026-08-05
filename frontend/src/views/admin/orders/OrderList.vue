<template>
  <div class="order-list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
          @change="onSearchSubmit"
        />
        <el-select v-model="filterParams.status" clearable style="width: 130px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
      </div>
    </div>

    <!-- 交易统计条 -->
    <div class="stats-bar">
      <div class="stat"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">订单总数</span></div>
      <div class="stat money"><span class="stat-num">¥{{ stats.amount }}</span><span class="stat-label">交易额(本页)</span></div>
      <div class="stat done"><span class="stat-num">{{ stats.completed }}</span><span class="stat-label">已完成</span></div>
      <div class="stat rate"><span class="stat-num">{{ stats.rate }}%</span><span class="stat-label">完成率</span></div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="订单号" width="180" />
        <el-table-column prop="product_id" label="商品 ID" min-width="120" />
        <el-table-column prop="buyer_id" label="买家" width="130" />
        <el-table-column prop="seller_id" label="卖家" width="130" />
        <el-table-column label="金额(元)" width="110">
          <template #default="{ row }">{{ ((row.amount_fen || 0) / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="下单时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total" layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData" @current-change="loadData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单号">{{ currentItem.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="商品 ID">{{ currentItem.product_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="金额(元)">{{ ((currentItem.amount_fen || 0) / 100).toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="买家">{{ currentItem.buyer_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="卖家">{{ currentItem.seller_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="下单时间">{{ formatDate(currentItem.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <div class="review-actions">
          <el-divider />
          <span style="margin-right: 12px; color: var(--el-text-color-regular);">修改状态：</span>
          <el-select v-model="newStatus" style="width: 140px;">
            <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
          <el-button type="primary" @click="onUpdateStatus">更新</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { showToast, showSuccessToast } from '@/utils/feedback'
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
const statusTagType = (s) => ({ completed: 'success', shipped: 'primary', paid: 'warning', pending: 'info', cancelled: 'info' }[s] || 'info')

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

const { listData, loading, total, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: getOrderList,
  idKey: 'id',
  defaultParams: { status: '' }
})

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
    showSuccessToast('状态已更新')
    loadData()
  } catch (e) { showToast('更新失败') }
}

onMounted(loadData)
</script>

<style scoped>
.order-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }

/* 交易统计条 */
.stats-bar { display: flex; gap: 32px; background: #fff; border-radius: 8px; padding: 14px 20px; margin-bottom: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.stat { display: flex; align-items: baseline; gap: 8px; }
.stat-num { font-size: 22px; font-weight: 700; color: var(--el-text-color-primary); }
.stat.money .stat-num { color: var(--el-color-warning); }
.stat.done .stat-num { color: var(--el-color-success); }
.stat.rate .stat-num { color: var(--el-color-info); }
.stat-label { font-size: 13px; color: var(--el-text-color-secondary); }

.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
