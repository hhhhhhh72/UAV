<template>
  <view class="orders-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">我的订单</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 状态筛选 -->
    <view class="orders-filters">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="s in statusOptions"
            :key="s.value"
            class="filter-chip"
            :class="{ active: currentStatus === s.value }"
            @tap="currentStatus = s.value"
          >{{ s.label }}</view>
        </view>
      </scroll-view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">订单记录</text>
      <text class="list-count">共 {{ filteredOrders.length }} 条</text>
    </view>

    <!-- 空状态 -->
    <view v-if="filteredOrders.length === 0" class="state-panel">
      <view class="state-mark">⌁</view>
      <text class="state-title">{{ loadError ? '加载失败' : '暂无订单' }}</text>
      <text class="state-desc">{{ loadError ? '网络异常，请稍后重试' : '确认接单或接单作业后，订单会展示在这里' }}</text>
      <view v-if="loadError" class="state-btn" @tap="fetchOrders">重新加载</view>
    </view>

    <!-- 订单列表 -->
    <view v-else class="order-list">
      <view v-for="order in filteredOrders" :key="order.id" class="order-card">
        <view class="order-head">
          <text class="order-no">{{ order.order_no || shortId(order.id) }}</text>
          <text class="tag" :class="statusTagClass(order.status)">{{ statusLabel(order.status) }}</text>
        </view>
        <view class="order-row"><text class="row-label">需求方</text><text class="row-value">{{ order.publisher_name || '企业用户' }}</text></view>
        <view class="order-row"><text class="row-label">服务方</text><text class="row-value">{{ order.worker_name || '飞手' }}</text></view>
        <view class="order-row"><text class="row-label">订单金额</text><text class="row-value strong">{{ formatAmount(order.amount_fen) }}</text></view>
        <view v-if="order.rework_note" class="order-row"><text class="row-label">整改要求</text><text class="row-value">{{ order.rework_note }}</text></view>
        <view v-if="order.cancel_reason" class="order-row"><text class="row-label">取消原因</text><text class="row-value">{{ order.cancel_reason }}</text></view>
        <view class="order-row"><text class="row-label">创建时间</text><text class="row-value">{{ formatDate(order.created_at) }}</text></view>

        <view class="order-actions">
          <template v-if="order.status === 'pending' && isWorker(order)">
            <view class="order-btn primary" @tap="startWork(order)">开始作业</view>
            <view class="order-btn" @tap="cancelOrder(order)">取消订单</view>
          </template>
          <template v-else-if="order.status === 'ongoing' && isWorker(order)">
            <view class="order-btn primary" @tap="completeWork(order)">确认完成</view>
            <view class="order-btn" @tap="cancelOrder(order)">取消订单</view>
          </template>
          <template v-else-if="order.status === 'awaiting_accept' && isPublisher(order)">
            <view class="order-btn primary" @tap="acceptWork(order)">验收通过</view>
            <view class="order-btn" @tap="openRework(order)">提出整改</view>
          </template>
          <template v-else-if="!order.status || order.status === 'pending' || order.status === 'ongoing' || order.status === 'awaiting_accept'">
            <view class="order-btn" @tap="cancelOrder(order)">取消订单</view>
          </template>
        </view>
      </view>
    </view>

    <!-- 提出整改 -->
    <u-popup :show="reworkShow" position="bottom" round @close="reworkShow = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">提出整改要求</text>
          <view class="sheet-close" @tap="reworkShow = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <text class="sheet-desc">填写整改要求后，订单将退回服务方继续作业。</text>
          <input
            class="note-input"
            v-model="reworkNote"
            placeholder="例如：成果照片缺少缺陷标注，请补充后重新提交"
            placeholder-class="note-ph"
          />
          <view class="sheet-btn primary" :class="{ disabled: !reworkNote }" @tap="submitRework">提交整改要求</view>
        </view>
      </view>
    </u-popup>

    <!-- 取消订单 -->
    <u-popup :show="cancelShow" position="bottom" round @close="cancelShow = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">取消订单</text>
          <view class="sheet-close" @tap="cancelShow = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <text class="sheet-desc">取消后订单即终止，请填写取消原因。</text>
          <input
            class="note-input"
            v-model="cancelReason"
            placeholder="例如：双方协商一致取消"
            placeholder-class="note-ph"
          />
          <view class="sheet-btn primary" :class="{ disabled: !cancelReason }" @tap="submitCancel">确认取消</view>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { safeNavigateTo } from '../../../utils/nav'
import { request, getStoredUser, getErrorMessage } from '../../../utils/request'

const orders = ref([])
const currentStatus = ref('')
const loadError = ref(false)
const reworkShow = ref(false)
const reworkTarget = ref(null)
const reworkNote = ref('')
const cancelShow = ref(false)
const cancelTarget = ref(null)
const cancelReason = ref('')

const STATUS_LABEL = {
  pending: '待开始',
  ongoing: '进行中',
  awaiting_accept: '待验收',
  completed: '已完成',
  cancelled: '已取消',
}
const statusLabel = (s) => STATUS_LABEL[s] || s || ''
const statusTagClass = (s) =>
  s === 'completed' ? 'green' : s === 'cancelled' ? 'gray' : s === 'awaiting_accept' ? 'orange' : 'blue'

const statusOptions = [
  { label: '全部', value: '' },
  { label: '待开始', value: 'pending' },
  { label: '进行中', value: 'ongoing' },
  { label: '待验收', value: 'awaiting_accept' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
]

const filteredOrders = computed(() =>
  currentStatus.value === '' ? orders.value : orders.value.filter((o) => o.status === currentStatus.value)
)

const me = getStoredUser() || {}
const isWorker = (o) => o.worker_id && me.id && o.worker_id === me.id
const isPublisher = (o) => o.publisher_id && me.id && o.publisher_id === me.id

const shortId = (id) => (id || '').length > 10 ? id.slice(-8) : (id || '-')
const formatAmount = (fen) => (fen == null || fen === 0 ? '面议' : (fen / 100).toFixed(2).replace(/\.00$/, '') + ' 元')
const formatDate = (iso) => {
  if (!iso) return ''
  const d = new Date(iso)
  const m = d.getMonth() + 1
  const day = d.getDate()
  return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
}

const fetchOrders = async () => {
  loadError.value = false
  try {
    const res = await request({ url: '/api/v1/work-orders/mine' })
    const data = Array.isArray(res) ? res : (res && res.data) || []
    orders.value = data
  } catch {
    loadError.value = true
    orders.value = []
  }
}

onLoad(fetchOrders)
onPullDownRefresh(() => {
  fetchOrders().finally(() => uni.stopPullDownRefresh())
})

async function startWork(order) {
  try {
    await request({ url: '/api/v1/work-orders/' + encodeURIComponent(order.id) + '/start', method: 'POST' })
    order.status = 'ongoing'
    uni.showToast({ title: '已开始作业', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

async function completeWork(order) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '确认完成', content: '确认作业已完成并提交验收？', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  try {
    await request({ url: '/api/v1/work-orders/' + encodeURIComponent(order.id) + '/complete', method: 'POST', data: { result_photos: [] } })
    order.status = 'awaiting_accept'
    uni.showToast({ title: '已提交验收', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

async function acceptWork(order) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '验收通过', content: '确认验收通过？通过后订单完成。', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  try {
    await request({ url: '/api/v1/work-orders/' + encodeURIComponent(order.id) + '/accept', method: 'POST' })
    order.status = 'completed'
    uni.showToast({ title: '验收通过', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

function openRework(order) {
  reworkTarget.value = order
  reworkNote.value = ''
  reworkShow.value = true
}

async function submitRework() {
  const order = reworkTarget.value
  if (!order || !reworkNote.value) return
  try {
    await request({
      url: '/api/v1/work-orders/' + encodeURIComponent(order.id) + '/rework',
      method: 'POST',
      data: { note: reworkNote.value },
    })
    reworkShow.value = false
    order.status = 'ongoing'
    order.rework_note = reworkNote.value
    uni.showToast({ title: '已提出整改要求', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

function cancelOrder(order) {
  cancelTarget.value = order
  cancelReason.value = ''
  cancelShow.value = true
}

async function submitCancel() {
  const order = cancelTarget.value
  if (!order || !cancelReason.value) return
  try {
    await request({
      url: '/api/v1/work-orders/' + encodeURIComponent(order.id) + '/cancel',
      method: 'POST',
      data: { reason: cancelReason.value },
    })
    cancelShow.value = false
    order.status = 'cancelled'
    order.cancel_reason = cancelReason.value
    uni.showToast({ title: '订单已取消', icon: 'none' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}

const goBack = () => uni.navigateBack()
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-spacer { width: 72rpx; }

/* 筛选 */
.orders-filters {
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  padding: 20rpx 24rpx;
}
.filter-scroll { white-space: nowrap; }
.filter-inner { display: inline-flex; gap: 12rpx; }
.filter-chip {
  display: inline-flex;
  align-items: center;
  height: 56rpx;
  padding: 0 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #344054;
  font-size: 24rpx;
  box-sizing: border-box;
}
.filter-chip.active {
  color: #0A66C2;
  border-color: #B9D6EF;
  background: #EAF3FB;
  font-weight: 650;
}

/* 列表 */
.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.order-list { padding: 0 32rpx 32rpx; }
.order-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.order-card + .order-card { margin-top: 20rpx; }

.order-head { display: flex; align-items: center; justify-content: space-between; }
.order-no { font-size: 24rpx; font-weight: 700; color: #667085; letter-spacing: 0.5rpx; }
.tag {
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.gray { color: #667085; background: #F1F3F5; }

.order-row {
  display: flex;
  gap: 16rpx;
  margin-top: 14rpx;
  font-size: 24rpx;
}
.row-label { color: #98A2B3; width: 120rpx; flex-shrink: 0; }
.row-value { color: #344054; flex: 1; }
.row-value.strong { color: #17212B; font-weight: 700; }

.order-actions {
  display: flex;
  gap: 14rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.order-btn {
  flex: 1;
  height: 62rpx;
  border-radius: 10rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #0A66C2;
  border: 1px solid #C7DEF1;
  background: #fff;
}
.order-btn.primary { color: #fff; background: #0A66C2; border-color: #0A66C2; }

/* 弹层 */
.sheet { padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom)); }
.sheet-head { display: flex; align-items: center; }
.sheet-title { flex: 1; font-size: 32rpx; font-weight: 750; color: #17212B; }
.sheet-close { padding: 8rpx; }
.sheet-x { font-size: 36rpx; color: #98A2B3; line-height: 1; }
.sheet-body { margin-top: 24rpx; }
.sheet-desc { font-size: 24rpx; color: #667085; line-height: 1.6; }
.note-input {
  height: 88rpx;
  margin: 28rpx 0;
  padding: 0 24rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #FAFAFA;
  font-size: 26rpx;
}
.note-ph { color: #98A2B3; }
.sheet-btn {
  height: 88rpx;
  border-radius: 50rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 650;
  background: #0A66C2;
  color: #fff;
  box-shadow: 0 6px 16px rgba(10, 102, 194, 0.28);
}
.sheet-btn.disabled { opacity: 0.5; }

/* 空状态 */
.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 0; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  margin-top: 32rpx;
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}
</style>
