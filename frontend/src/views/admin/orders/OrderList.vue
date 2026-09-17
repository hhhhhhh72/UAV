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
        <span>¥{{ ((record.amount_fen || 0) / 100).toFixed(2) }}</span>
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
    <a-modal v-model:visible="detailVisible" title="订单详情" :width="'min(600px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="订单号">{{ currentItem.id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagColor(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="商品 ID">{{ currentItem.product_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="金额(元)">¥{{ ((currentItem.amount_fen || 0) / 100).toFixed(2) }}</a-descriptions-item>
          <a-descriptions-item label="买家">{{ currentItem.buyer_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="卖家">{{ currentItem.seller_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="下单时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>

        <!-- 收货信息：下单时快照到订单。运营靠它核对/协助发货——此前订单表根本没有地址。 -->
        <template v-if="currentItem.receiver_name || currentItem.receiver_address">
          <a-divider>收货信息</a-divider>
          <a-descriptions :column="2" bordered size="medium">
            <a-descriptions-item label="收货人">{{ currentItem.receiver_name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="联系电话">{{ currentItem.receiver_phone || '-' }}</a-descriptions-item>
            <a-descriptions-item label="所在地区">{{ currentItem.receiver_region || '-' }}</a-descriptions-item>
            <a-descriptions-item label="详细地址">{{ currentItem.receiver_address || '-' }}</a-descriptions-item>
          </a-descriptions>
        </template>

        <!-- 发货信息：出库方向的物流留痕（与售后里的"退货单号"方向相反） -->
        <template v-if="currentItem.shipping_tracking || currentItem.shipped_at">
          <a-divider>发货信息</a-divider>
          <a-descriptions :column="2" bordered size="medium">
            <a-descriptions-item label="快递公司">{{ currentItem.shipping_company || '-' }}</a-descriptions-item>
            <a-descriptions-item label="快递单号">{{ currentItem.shipping_tracking || '（自提订单无单号）' }}</a-descriptions-item>
            <a-descriptions-item label="发货时间">{{ currentItem.shipped_at ? formatDate(currentItem.shipped_at) : '-' }}</a-descriptions-item>
          </a-descriptions>
        </template>

        <!-- 售后单（aftersale 记录） -->
        <template v-if="currentItem.aftersale_status">
          <a-divider>售后单</a-divider>
          <a-descriptions :column="2" bordered size="medium">
            <a-descriptions-item label="售后类型">{{ currentItem.aftersale_type === 'return' ? '退货退款' : '仅退款' }}</a-descriptions-item>
            <a-descriptions-item label="审核状态">
              <a-tag :color="aftersaleTagColor(currentItem.aftersale_status)" size="small">{{ aftersaleStatusLabel(currentItem.aftersale_status) }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="退款金额(元)">¥{{ ((currentItem.aftersale_amount_fen || 0) / 100).toFixed(2) }}</a-descriptions-item>
            <a-descriptions-item label="申请时间">{{ formatDate(currentItem.aftersale_time) }}</a-descriptions-item>
            <a-descriptions-item label="原因">{{ currentItem.aftersale_reason || '-' }}</a-descriptions-item>
            <a-descriptions-item label="说明">{{ currentItem.aftersale_desc || '-' }}</a-descriptions-item>
            <!-- 退货退款专属：物流单号与寄回时间（仅 return 单且买家已寄回时才有值） -->
            <a-descriptions-item v-if="currentItem.return_tracking" label="退货单号">{{ currentItem.return_tracking }}</a-descriptions-item>
            <a-descriptions-item v-if="currentItem.returned_at" label="寄回时间">{{ formatDate(currentItem.returned_at) }}</a-descriptions-item>
          </a-descriptions>
        </template>

        <div class="review-actions">
          <a-divider />
          <template v-if="currentItem.aftersale_status === 'pending'">
            <span class="review-label">售后审核：</span>
            <!-- 退货退款的「同意」只代表同意退货，买家寄回并由卖家确认收到后才退款 -->
            <a-button type="primary" status="success" @click="onReviewAftersale('approve')">
              {{ currentItem.aftersale_type === 'return' ? '同意退货' : '同意退款' }}
            </a-button>
            <a-button status="danger" @click="onReviewAftersale('reject')">驳回申请</a-button>
          </template>
          <!-- 退货退款：买家已寄回 → 确认收到货才发起退款 -->
          <template v-else-if="currentItem.aftersale_status === 'returned'">
            <span class="review-label">退货已寄回：</span>
            <a-button type="primary" status="success" :loading="confirmReturning" @click="onConfirmReturn">确认收到退货并发起退款</a-button>
          </template>
          <!-- 退货退款：已同意退货，等待买家寄回（此阶段无管理端操作） -->
          <template v-else-if="currentItem.aftersale_status === 'returning'">
            <span class="review-label review-closed">已同意退货，等待买家寄回商品</span>
          </template>
          <!-- 售后已结案（approved/rejected）：只读展示，不再提供状态修改，防止对已退款完成订单重复操作 -->
          <template v-else-if="currentItem.aftersale_status">
            <span class="review-label review-closed">售后已结案，无需操作</span>
          </template>
          <template v-else>
            <span class="review-label">修改状态：</span>
            <!-- 改单下拉不含 aftersale：后端 UpdateStatusAdmin 明确拒绝直达该状态
                 （必须走"申请售后"接口，否则会留下 aftersale_status 为空的死状态） -->
            <a-select v-model="newStatus" style="width: 140px;">
              <a-option v-for="s in updateStatusOptions" :key="s.value" :label="s.label" :value="s.value" />
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
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { updateOrderStatus, reviewAftersale, confirmReturnReceived } from '@/api/admin/order'
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
// 改单可选状态：排除 aftersale（后端只接受经"申请售后"进入该状态）
const updateStatusOptions = statusOptions.filter(o => o.value !== 'aftersale')
const statusTagColor = (s) => ({ completed: 'green', shipped: 'arcoblue', paid: 'orange', aftersale: 'purple', pending: 'gray', cancelled: 'gray' }[s] || 'gray')

// 售后单审核状态：
//   仅退款 refund ：pending 待审核 → approved 已退款 / rejected 已驳回
//   退货退款 return：pending 待审核 → returning 待买家寄回 → returned 待卖家确认收货 → approved 已退款
// 退货单的退款发生在最后一环，中间的 returning/returned 都还没动钱，文案必须区分开
const aftersaleStatusLabel = (s) => ({
  pending: '待审核',
  returning: '待买家寄回',
  returned: '待确认收货',
  approved: '已退款',
  rejected: '已驳回',
}[s] || s || '-')
const aftersaleTagColor = (s) => ({
  pending: 'orange',
  returning: 'arcoblue',
  returned: 'arcoblue',
  approved: 'green',
  rejected: 'red',
}[s] || 'gray')

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
  { title: '收货人', dataIndex: 'receiver_name', width: 100 },
  { title: '快递单号', dataIndex: 'shipping_tracking', width: 140 },
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

// 退货退款收尾：确认收到买家寄回的商品 → 此刻才发起退款（仅 aftersale_status=returned 可用）
const confirmReturning = ref(false)
const onConfirmReturn = () => {
  if (!currentItem.value || confirmReturning.value) return
  Modal.confirm({
    title: '确认收到退货',
    content: `确认已收到买家寄回的商品？确认后将退款 ¥${((currentItem.value.aftersale_amount_fen || 0) / 100).toFixed(2)} 并结案。`,
    okText: '确认收到',
    cancelText: '取消',
    onOk: async () => {
      confirmReturning.value = true
      try {
        await confirmReturnReceived(currentItem.value.id)
        // 无托管资金的订单（管理端建单/线下成交）后端只结案、不产生退款流水，
        // 因此这里不能断言"退款已完成"（service.refundForAftersale 的 C 分支）。
        Message.success('已确认收到退货，售后已结案')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.message || '操作失败') }
      finally { confirmReturning.value = false }
    }
  })
}

// 售后审核：同意退款（approve）/ 驳回（reject）——仅 aftersale_status=pending 可审
const onReviewAftersale = (action) => {
  if (!currentItem.value) return
  const approve = action === 'approve'
  const isReturn = currentItem.value.aftersale_type === 'return'
  Modal.confirm({
    title: approve ? (isReturn ? '同意退货' : '同意退款') : '驳回售后申请',
    content: approve
      ? (isReturn
        ? `同意退货 ¥${((currentItem.value.aftersale_amount_fen || 0) / 100).toFixed(2)}？买家寄回后由你确认收到，届时才发起退款。`
        : `确认同意退款 ¥${((currentItem.value.aftersale_amount_fen || 0) / 100).toFixed(2)}？结案后订单回到已完成状态。`)
      : '确认驳回该售后申请？驳回后订单回到已完成状态。',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        await reviewAftersale(currentItem.value.id, action)
        // 退货退款的 approve 只是"同意退货"，钱要等卖家确认收到货才退，不能提示"已同意退款"
        Message.success(approve ? (isReturn ? '已同意退货，等待买家寄回' : '已同意退款') : '已驳回')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.message || '操作失败') }
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

.stat-label { font-size: 12px; color: var(--color-text-2); }

.stat.money .stat-num { color: #E96012; }
.stat.done .stat-num { color: #168A55; }
.stat.rate .stat-num { color: #165DFF; }

.time-text { color: var(--color-text-2); font-size: 12px; }
.no-aftersale { color: #C9CDD4; }
.review-closed { color: var(--color-text-3); font-size: 13px; }

.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
.review-label { color: #4E5969; }
</style>
