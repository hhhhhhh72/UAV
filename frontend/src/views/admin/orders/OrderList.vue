<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="orders"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      @loaded="onLoaded"
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
      <template #aftersale="{ record }">
        <a-tag v-if="record.aftersale_status" :color="aftersaleTagColor(record.aftersale_status)" size="small">{{ aftersaleStatusLabel(record.aftersale_status) }}</a-tag>
        <span v-else class="no-aftersale">-</span>
      </template>
      <template #actions="{ record }">
        <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
      </template>
      <template #empty>
        <a-empty description="暂无数据" />
      </template>
    </CrudList>

    <!-- 交易统计条（基于当前页 + 接口 total） -->
    <a-card :bordered="false" class="stat-card">
      <div class="stats-bar">
        <div class="stat"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">订单总数</span></div>
        <div class="stat money"><span class="stat-num">¥{{ stats.amount }}</span><span class="stat-label">交易额(本页)</span></div>
        <div class="stat done"><span class="stat-num">{{ stats.completed }}</span><span class="stat-label">已完成</span></div>
        <div class="stat rate"><span class="stat-num">{{ stats.rate }}%</span><span class="stat-label">完成率</span></div>
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

        <!-- 售后单（aftersale 记录） -->
        <template v-if="currentItem.aftersale_status">
          <a-divider>售后单</a-divider>
          <a-descriptions :column="2" bordered size="medium">
            <a-descriptions-item label="售后类型">{{ currentItem.aftersale_type === 'return' ? '退货退款' : '仅退款' }}</a-descriptions-item>
            <a-descriptions-item label="审核状态">
              <a-tag :color="aftersaleTagColor(currentItem.aftersale_status)" size="small">{{ aftersaleStatusLabel(currentItem.aftersale_status) }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="退款金额(元)">{{ ((currentItem.aftersale_amount_fen || 0) / 100).toFixed(2) }}</a-descriptions-item>
            <a-descriptions-item label="申请时间">{{ formatDate(currentItem.aftersale_time) }}</a-descriptions-item>
            <a-descriptions-item label="原因">{{ currentItem.aftersale_reason || '-' }}</a-descriptions-item>
            <a-descriptions-item label="说明">{{ currentItem.aftersale_desc || '-' }}</a-descriptions-item>
          </a-descriptions>
        </template>

        <div class="review-actions">
          <a-divider />
          <template v-if="currentItem.aftersale_status === 'pending'">
            <span class="review-label">售后审核：</span>
            <a-button type="primary" status="success" @click="onReviewAftersale('approve')">同意退款</a-button>
            <a-button status="danger" @click="onReviewAftersale('reject')">驳回申请</a-button>
          </template>
          <template v-else>
            <span class="review-label">修改状态：</span>
            <a-select v-model="newStatus" style="width: 140px;">
              <a-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
            </a-select>
            <a-button type="primary" @click="onUpdateStatus">更新</a-button>
          </template>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { updateOrderStatus, reviewAftersale } from '@/api/admin/order'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()

const statusOptions = [
  { label: '待付款', value: 'pending' },
  { label: '已付款', value: 'paid' },
  { label: '已发货', value: 'shipped' },
  { label: '退款/售后', value: 'aftersale' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagColor = (s) => ({ completed: 'green', shipped: 'arcoblue', paid: 'orange', aftersale: 'purple', pending: 'gray', cancelled: 'gray' }[s] || 'gray')

// 售后单审核状态：pending=待审核 / approved=已同意退款 / rejected=已驳回
const aftersaleStatusLabel = (s) => ({ pending: '待审核', approved: '已同意退款', rejected: '已驳回' }[s] || s || '-')
const aftersaleTagColor = (s) => ({ pending: 'orange', approved: 'green', rejected: 'red' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 订单无合适的批量业务动作（状态机按行流转，金融记录不做批量变更）
const batchActions = []

const searchFields = [
  { key: 'status', label: '状态', type: 'select', width: 130, options: [
    { value: '', label: '全部状态' },
    ...statusOptions
  ]},
  // 日期范围：提交时合并为 start_date/end_date（后端 listAdminOrders 按 created_at 过滤）
  { key: 'dateRange', label: '日期范围', type: 'range', width: 260 }
]

const columns = [
  { title: '订单号', dataIndex: 'id', width: 180 },
  { title: '商品 ID', dataIndex: 'product_id', minWidth: 120 },
  { title: '买家', dataIndex: 'buyer_id', width: 130 },
  { title: '卖家', dataIndex: 'seller_id', width: 130 },
  { title: '金额(元)', dataIndex: 'amount_fen', slotName: 'amount', width: 110, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 110 },
  { title: '售后', dataIndex: 'aftersale_status', slotName: 'aftersale', width: 100 },
  { title: '下单时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '操作', slotName: 'actions', width: 120, fixed: 'right' },
]

// 交易统计（分类/金额基于当前页；订单总数取接口 total）
const stats = ref({ total: 0, amount: '0.00', completed: 0, rate: 0 })
const onLoaded = (rows, totalCount) => {
  const amount = (rows || []).reduce((s, x) => s + (x.amount_fen || 0), 0) / 100
  const completed = (rows || []).filter((x) => x.status === 'completed').length
  const rate = (rows || []).length ? Math.round((completed / (rows || []).length) * 100) : 0
  stats.value = {
    total: totalCount || 0,
    amount: amount.toLocaleString('zh-CN', { minimumFractionDigits: 2 }),
    completed,
    rate
  }
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
    crudRef.value?.reload()
  } catch (e) { Message.error('更新失败') }
}

// 售后审核：同意退款（approve）/ 驳回（reject）——仅 aftersale+pending 可审
const onReviewAftersale = (action) => {
  if (!currentItem.value) return
  const approve = action === 'approve'
  Modal.confirm({
    title: approve ? '同意退款' : '驳回售后申请',
    content: approve
      ? `确认同意退款 ¥${((currentItem.value.aftersale_amount_fen || 0) / 100).toFixed(2)}？结案后订单回到已完成状态。`
      : '确认驳回该售后申请？驳回后订单回到已完成状态。',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        await reviewAftersale(currentItem.value.id, action)
        Message.success(approve ? '已同意退款' : '已驳回')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

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
.no-aftersale { color: #C9CDD4; }

.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
.review-label { color: #4E5969; }
</style>
