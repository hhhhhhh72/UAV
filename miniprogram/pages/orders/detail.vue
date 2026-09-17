<template>
  <view class="order-detail-page">
    <u-nav-bar :title="navTitle" show-back @back="goBack" />

    <view v-if="loading" class="state-panel">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="error || !order" class="state-panel">
      <u-empty description="订单加载失败" />
      <view class="retry-btn" @tap="loadData">
        <text>重新加载</text>
      </view>
    </view>

    <template v-else>
      <!-- 头部商品卡 -->
      <view class="detail-hero" v-if="order.detail && order.detail.hero">
        <text class="hero-kind" :class="kindClass(order.type)">{{ order.kind_label }}</text>
        <text class="hero-title">{{ order.detail.hero.title }}</text>
        <text class="hero-sub">{{ order.detail.hero.sub }}</text>
      </view>

      <!-- 媒体区 + 名称价格（无 hero 时） -->
      <view v-else class="detail-media-card">
        <image v-if="order.image" class="detail-media" :src="order.image" mode="aspectFill" />
        <view v-else class="detail-media detail-media-ph">
          <text>{{ kindMark(order.kind_label) }}</text>
        </view>
        <view class="media-body">
          <text class="media-kind" :class="kindClass(order.type)">{{ order.kind_label }}</text>
          <text class="media-title">{{ order.title }}</text>
          <view class="media-price-row">
            <text class="media-price">¥{{ fmtFen(order.amount_fen) }}</text>
            <text class="media-qty">{{ order.quantity_label }}</text>
          </view>
          <text class="media-sub">{{ order.subtitle }}</text>
        </view>
      </view>

      <!-- 分区明细 -->
      <view
        v-for="(section, i) in order.detail.sections || []"
        :key="i"
        class="detail-card"
      >
        <text class="detail-card-title">{{ section.title }}</text>
        <view v-for="(row, ri) in section.rows" :key="ri" class="data-row">
          <text class="data-label">{{ row.label }}</text>
          <text class="data-value" :class="rowClass(row.status)">{{ row.value }}</text>
        </view>
      </view>

      <!-- 物流入口（待收货） -->
      <view v-if="order.detail && order.detail.logistics" class="detail-list">
        <view class="detail-list-item" @tap="goLogistics">
          <view class="list-mark">物</view>
          <view class="list-copy">
            <text class="list-title">查看物流详情</text>
            <text class="list-desc">查看完整运输节点和配送进度</text>
          </view>
          <text class="list-arrow">›</text>
        </view>
      </view>

      <!-- 主操作 -->
      <view class="primary-action">
        <view class="primary-btn" @tap="handlePrimaryAction">
          <text>{{ order.action }}</text>
        </view>
        <text class="action-note" v-if="actionNote">{{ actionNote }}</text>
      </view>

      <!-- 取消订单（未付款订单：买卖双方均可取消；取消后商品自动重新上架） -->
      <view v-if="showCancelEntry" class="ghost-action" hover-class="ghost-action--active" @tap="cancelOrder()">
        <text>取消订单</text>
      </view>

      <!-- 申请售后（买家：已发货/已完成且无售后记录时） -->
      <view v-if="showAftersaleEntry" class="aftersale-entry" hover-class="aftersale-entry--active" @tap="goRefundApply">
        <view class="aftersale-entry-copy">
          <text class="aftersale-entry-title">{{ order.aftersale ? '重新申请售后' : '申请售后' }}</text>
          <text class="aftersale-entry-hint">{{ order.aftersale ? '上次申请已被驳回，可修改说明后重新提交' : '商品有问题可申请仅退款或退货退款，由卖家审核处理' }}</text>
        </view>
        <text class="aftersale-entry-arrow">›</text>
      </view>
    </template>

    <view class="bottom-spacer"></view>
  </view>

  <!-- 发货：卖家填快递公司 + 单号。单号必填——没有单号的"已发货"买家查不到物流，
       出问题也无从举证；服务端同样强制（POST /trade-orders/{id}/ship）。 -->
  <view v-if="shipVisible" class="ship-mask" @tap="shipVisible = false">
    <view class="ship-sheet" @tap.stop>
      <view class="ship-title">填写发货信息</view>
      <view class="ship-row">
        <text class="ship-label">快递公司</text>
        <input v-model="shipForm.company" class="ship-input" placeholder="如：顺丰速运（选填）" :maxlength="30" />
      </view>
      <view class="ship-row">
        <text class="ship-label">快递单号</text>
        <input v-model="shipForm.tracking" class="ship-input" placeholder="物流/同城必填；自提可留空" :maxlength="60" />
      </view>
      <view class="ship-actions">
        <view class="ship-btn ship-btn--ghost" hover-class="btn-press" @tap="shipVisible = false">取消</view>
        <view class="ship-btn ship-btn--primary" hover-class="btn-press" @tap="submitShip">确认发货</view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
// request 必须显式引入：此前本文件用到 request（支付 / 状态流转）却没有 import，
// 三个核心动作（确认支付、发货、确认收货）全部抛 ReferenceError 被 catch 吞成
// "失败，请稍后重试"，前端根本发不出请求。
import { request } from '../../utils/request'
import { loadOrder, fmtFen } from '../../utils/orderAdapter'

const order = ref(null)
const loading = ref(true)
const error = ref(false)
let orderId = ''

const navTitle = computed(() => {
  if (!order.value) return '订单详情'
  if (order.value.type === 'course') return '培训订单详情'
  if (order.value.type === 'service') return '服务订单详情'
  return '商品订单详情'
})

// 申请售后入口：仅买家、已付款（未发货退款）/已发货/已完成。
// 已有售后记录时不再显示入口，但**被驳回（rejected）的除外**——后端允许驳回后重新申请
//（ApplyAftersale 的判重只拦 pending/returning/returned/approved），
// 此前前端一律 return false，买家被驳回一次就再也点不到入口，与后端契约不一致。
const showAftersaleEntry = computed(() => {
  const o = order.value
  if (!o || o.role === 'seller') return false
  if (o.aftersale && o.aftersale.status_key !== 'rejected') return false
  return o.status === 'paid' || o.status === 'shipped' || o.status === 'completed'
})

// 取消订单入口：仅未付款（pending）——后端只允许 pending 由买卖双方取消；
// 已付款订单的退款必须走「申请售后」（支付后未发货退货），保持单一退款口径。
const showCancelEntry = computed(() => !!(order.value && order.value.status === 'pending'))

const actionNote = computed(() => {
  if (!order.value) return ''
  const o = order.value
  // 卖家视角：发货方文案
  if (o.role === 'seller') {
    if (o.status === 'pending') return '等待买家完成付款，付款后即可发货（超时未付将自动取消）'
    if (o.status === 'paid') return '确认发货后订单将标记为已发货'
    if (o.status === 'shipped') return '等待买家确认收货'
    if (o.status === 'completed') return '交易已完成，感谢你的销售'
    return ''
  }
  // 买家视角：付款方文案
  if (o.aftersale) {
    return o.aftersale.status_key === 'rejected'
      ? '售后申请已被驳回，可重新申请或联系客服'
      : '退款由卖家审核，可在售后详情查看处理进度'
  }
  if (o.status === 'pending') return '确认支付后订单将标记为已支付；暂不购买可直接取消'
  if (o.status === 'paid') return '卖家将在 48 小时内发货，请留意物流更新'
  if (o.status === 'shipped') return '确认收货后订单将完成'
  if (o.status === 'completed') return o.action === '已评价' ? '感谢你的评价，结课凭证已存入「我的报名 / 证书」' : '评价后结课凭证进入「我的报名 / 证书」'
  return ''
})

const loadData = async (query = {}) => {
  const id = query.id || orderId
  if (!id) {
    error.value = true
    loading.value = false
    return
  }
  loading.value = true
  error.value = false
  try {
    order.value = await loadOrder(id)
    if (!order.value) error.value = true
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }
}

onLoad((options) => {
  orderId = (options && options.id) || ''
  loadData(options)
})

// onShow 刷新：从支付/发货/收货/评价等子页返回后重载订单状态（订单详情为低频页，
// 重复请求可接受；onLoad 先执行完成首屏加载，onShow 只负责后续返回时的刷新）
onShow(() => {
  if (orderId) loadData({ id: orderId })
})

const handlePrimaryAction = () => {
  if (!order.value) return
  const o = order.value
  // 有售后记录的订单（含结案单：状态已回 completed）一律看售后详情
  if (o.aftersale) {
    uni.navigateTo({
      url: `/pages/orders/aftersale?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
    return
  }
  // 已取消订单：无任何可操作项
  if (o.status === 'cancelled') {
    uni.showToast({ title: '订单已取消', icon: 'none' })
    return
  }
  // ── 卖家视角：发货 / 等待流转提示 ──
  if (o.role === 'seller') {
    if (o.status === 'paid') {
      // 发货：调后端状态机 paid→shipped
      shipOrder(o)
      return
    }
    const sellerTips = {
      pending: '等待买家完成付款',
      shipped: '等待买家确认收货',
      completed: '交易已完成',
      aftersale: '售后处理中',
    }
    uni.showToast({ title: sellerTips[o.status] || '请等待买家操作', icon: 'none' })
    return
  }
  // ── 买家视角：支付 / 收货 / 评价 ──
  if (o.status === 'pending') {
    // 支付：调后端状态机 pending→paid
    payOrder(o)
    return
  }
  if (o.status === 'paid') {
    // 提醒发货：发送提醒成功反馈，不把状态改为待收货
    uni.showToast({ title: '已向商家发送发货提醒', icon: 'none' })
    return
  }
  if (o.status === 'shipped') {
    // 确认收货：调后端状态机 shipped→completed
    receiveOrder(o)
    return
  }
  if (o.status === 'completed') {
    // 去评价（结案单已在上方 o.aftersale 分支处理，不会走到这里）
    uni.navigateTo({
      url: `/pages/orders/review?id=${encodeURIComponent(o.id)}&type=${o.type}`,
    })
    return
  }
}

// 支付：POST /api/v1/trade-orders/{id}/pay 置 paid（真实微信支付接入后替换此逻辑）
const payOrder = async (o) => {
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '确认支付',
      content: '确认支付 ¥' + fmtFen(o.amount_fen) + '？支付后订单进入待发货。',
      confirmText: '确认支付',
      success: (r) => resolve(!!r.confirm),
      fail: () => resolve(false),
    })
  })
  if (!confirmed) return
  uni.showLoading({ title: '支付中...' })
  try {
    await request({ url: '/api/v1/trade-orders/' + encodeURIComponent(o.id) + '/pay', method: 'POST' })
    uni.hideLoading()
    uni.showToast({ title: '支付成功', icon: 'success' })
    setTimeout(() => { loadData({ id: o.id }) }, 600)
  } catch (e) {
    uni.hideLoading()
    // 402 = 托管金余额不足（后端 phase3.go 把 ErrInsufficientBalance 映射成 402）。
    // 与课程报名页（register.vue 的 402 分支）同一口径：弹窗说明并直接带去充值，
    // 否则买家只看到一句干巴巴的"余额不足"，不知道去哪儿充。
    if (e && e.statusCode === 402) {
      uni.showModal({
        title: '托管金余额不足',
        content: '商城下单走平台担保交易，支付时会从托管金冻结货款。请先充值（模拟通道，单笔上限 20 万元），再回来支付。',
        confirmText: '去充值',
        success: (r) => { if (r.confirm) uni.navigateTo({ url: '/pages/escrow/index' }) },
      })
      return
    }
    const msg = (e && e.data && e.data.error && e.data.error.message) || '支付失败，请稍后重试'
    uni.showToast({ title: msg, icon: 'none' })
  }
}

// 发货：卖家填写快递公司 + 单号，走 POST /trade-orders/{id}/ship。
//
// 不再用 PATCH status=shipped：那条路没有单号可填，会造出"已发货但查不到物流"的订单，
// 服务层也已封掉（actorAllowedTransition）。单号必填。
const shipVisible = ref(false)
const shipTarget = ref(null)
const shipForm = reactive({ company: '', tracking: '' })
const shipSubmitting = ref(false)
const shipOrder = (o) => {
  shipTarget.value = o
  shipForm.company = ''
  shipForm.tracking = ''
  shipVisible.value = true
}
const submitShip = async () => {
  // 单号不在这里强制：自提订单本来就没有快递单号，服务端会按商品的交付方式判断。
  const tracking = String(shipForm.tracking || '').trim()
  const o = shipTarget.value
  if (!o) return
  shipSubmitting.value = true
  uni.showLoading({ title: '发货中...' })
  try {
    await request({
      url: '/api/v1/trade-orders/' + encodeURIComponent(o.id) + '/ship',
      method: 'POST',
      data: { shipping_company: String(shipForm.company || '').trim(), shipping_tracking: tracking },
    })
    uni.hideLoading()
    uni.showToast({ title: '发货成功', icon: 'success' })
    shipVisible.value = false
    await loadData()
  } catch (e) {
    uni.hideLoading()
    const msg = (e && e.data && e.data.error && e.data.error.message) || '发货失败，请稍后重试'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    shipSubmitting.value = false
  }
}

// 确认收货：买家 PATCH 置 completed（真实物流接入后替换此逻辑）
const receiveOrder = async (o) => {
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '确认收货',
      content: '确认已收到商品？确认后订单将完成，可进行评价。',
      confirmText: '确认收货',
      success: (r) => resolve(!!r.confirm),
      fail: () => resolve(false),
    })
  })
  if (!confirmed) return
  patchOrderStatus(o, 'completed', '确认中...', '确认收货成功', '操作失败，请稍后重试')
}

// 取消订单：PATCH /api/v1/trade-orders/{id}/status 置 cancelled。
// 仅未付款订单可取消（后端 pending→cancelled），取消后商品恢复为可售。
const cancelOrder = async (o = order.value) => {
  if (!o) return
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '取消订单',
      content: '确认取消这笔未付款订单？取消后商品会重新上架。',
      confirmText: '确认取消',
      success: (r) => resolve(!!r.confirm),
      fail: () => resolve(false),
    })
  })
  if (!confirmed) return
  patchOrderStatus(o, 'cancelled', '取消中...', '订单已取消', '取消失败，请稍后重试')
}

// 状态流转共用：PATCH /api/v1/trade-orders/{id}/status 置新状态，成功即刷新订单
const patchOrderStatus = async (o, status, loadingText, successText, failText) => {
  uni.showLoading({ title: loadingText })
  try {
    await request({
      url: '/api/v1/trade-orders/' + encodeURIComponent(o.id) + '/status',
      method: 'PATCH',
      data: { status },
    })
    uni.hideLoading()
    uni.showToast({ title: successText, icon: 'success' })
    setTimeout(() => { loadData({ id: o.id }) }, 600)
  } catch (e) {
    uni.hideLoading()
    const msg = (e && e.data && e.data.error && e.data.error.message) || failText
    uni.showToast({ title: msg, icon: 'none' })
  }
}

// 申请售后：进入退款申请页（提交后回这里/售后详情看进度）
const goRefundApply = () => {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/orders/refund-apply?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
}

const goLogistics = () => {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/orders/logistics?id=${encodeURIComponent(order.value.id)}&type=${order.value.type}`,
  })
}

const goBack = () => {
  uni.navigateBack()
}

// ── 展示辅助 ──
const kindMark = (label) => (label ? label.charAt(0) : '单')
const kindClass = (type) => (type === 'course' ? 'course' : type === 'service' ? 'service' : '')
const rowClass = (status) => {
  if (status === 'good') return 'good'
  if (status === 'wait') return 'wait'
  return ''
}
</script>

<style scoped>
/* 发货弹层（卖家填写快递单号） */
.ship-mask {
  position: fixed; left: 0; right: 0; top: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.45); z-index: 999;
  display: flex; align-items: flex-end;
}
.ship-sheet {
  width: 100%; background: #fff;
  border-radius: 28rpx 28rpx 0 0; padding: 36rpx 32rpx calc(36rpx + env(safe-area-inset-bottom));
}
.ship-title { font-size: 32rpx; font-weight: 600; color: #1d2129; margin-bottom: 24rpx; }
.ship-row { display: flex; align-items: center; padding: 20rpx 0; border-bottom: 1rpx solid #f2f3f5; }
.ship-label { width: 160rpx; font-size: 28rpx; color: #4e5969; }
.ship-input { flex: 1; font-size: 28rpx; color: #1d2129; background: #fafafa; border-radius: 16rpx; padding: 14rpx 20rpx; }
.ship-actions { display: flex; gap: 20rpx; margin-top: 32rpx; }
.ship-btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; border-radius: 42rpx; font-size: 30rpx; }
.ship-btn--ghost { background: #f2f3f5; color: #4e5969; }
.ship-btn--primary { background: var(--color-primary, #0A66C2); color: #fff; }
.order-detail-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

.state-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120rpx;
}
.loading-inline {
  display: flex;
  align-items: center;
  gap: 12rpx;
  font-size: 26rpx;
  color: var(--color-text-secondary);
}
.retry-btn {
  margin-top: 12rpx;
  padding: 16rpx 48rpx;
  background: var(--color-primary);
  color: #fff;
  border-radius: 12rpx;
  font-size: 26rpx;
}

/* hero（演示数据） */
.detail-hero {
  margin: 20rpx 24rpx 0;
  padding: 32rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--color-primary-deep), #126DC8);
  border-radius: 16rpx;
  position: relative;
  overflow: hidden;
}
.detail-hero::after {
  content: '';
  position: absolute;
  right: -60rpx;
  top: -80rpx;
  width: 240rpx;
  height: 240rpx;
  border: 1rpx solid rgba(255,255,255,.18);
  border-radius: 50%;
}
.hero-kind {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: rgba(255,255,255,.18);
  font-size: 22rpx;
  font-weight: 600;
  position: relative;
  z-index: 1;
}
.hero-title {
  display: block;
  margin-top: 16rpx;
  font-size: 36rpx;
  font-weight: 700;
  line-height: 1.4;
  position: relative;
  z-index: 1;
}
.hero-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: rgba(255,255,255,.76);
  position: relative;
  z-index: 1;
}

/* media（真实商品订单） */
.detail-media-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
.detail-media {
  width: 100%;
  height: 300rpx;
  background: var(--color-primary-light);
}
.detail-media-ph {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 80rpx;
  font-weight: 700;
  color: var(--color-primary);
}
.media-body { padding: 24rpx; }
.media-kind {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 20rpx;
  font-weight: 600;
}
.media-kind.course { background: #F0EDFF; color: #7B61D1; }
.media-kind.service { background: var(--color-accent-light); color: var(--color-accent-deep); }
.media-title {
  display: block;
  margin-top: 12rpx;
  font-size: 32rpx;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.4;
}
.media-price-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-top: 12rpx;
}
.media-price {
  font-size: 36rpx;
  font-weight: 800;
  color: var(--color-accent-deep);
}
.media-qty { font-size: 22rpx; color: var(--color-text-secondary); }
.media-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

/* 分区卡片 */
.detail-card {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  padding: 28rpx;
}
.detail-card-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 16rpx;
}
.data-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  min-height: 76rpx;
  border-top: 1rpx solid var(--color-divider);
}
.data-row:first-of-type { border-top: none; }
.data-label { font-size: 24rpx; color: var(--color-text-secondary); flex-shrink: 0; }
.data-value { font-size: 24rpx; color: var(--color-text); text-align: right; word-break: break-all; }
.data-value.good { color: var(--color-success); }
.data-value.wait { color: var(--color-warning); }

/* 列表入口（物流） */
.detail-list {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
.detail-list-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  min-height: 104rpx;
  padding: 0 28rpx;
  box-sizing: border-box;
}
.list-mark {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 28rpx;
  font-weight: 700;
}
.list-copy { flex: 1; min-width: 0; }
.list-title {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: var(--color-text);
}
.list-desc {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: var(--color-text-placeholder);
}
.list-arrow { font-size: 30rpx; color: var(--color-text-placeholder); }

/* 主操作 */
.primary-action {
  margin: 32rpx 24rpx 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}.primary-btn {
  width: 100%;
  height: 92rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  box-shadow: 0 8rpx 20rpx rgba(10, 102, 194, 0.22);
}
.action-note {
  font-size: 22rpx;
  color: var(--color-text-placeholder);
  text-align: center;
}

/* 次要操作（取消订单）：描边样式，与主按钮区分层级；不使用危险红——
   取消未付款订单是常规操作，且此时还没有资金往来。 */
.ghost-action {
  margin: 20rpx 24rpx 0;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1rpx solid var(--color-border);
  border-radius: 12rpx;
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  font-size: 28rpx;
  font-weight: 600;
  box-sizing: border-box;
}
.ghost-action--active { opacity: 0.7; }

/* 申请售后入口 */
.aftersale-entry {
  margin: 20rpx 24rpx 0;
  background: #fff;
  border: 1rpx solid var(--color-border);
  border-radius: 16rpx;
  box-shadow: var(--shadow-sm);
  min-height: 104rpx;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 20rpx;
  box-sizing: border-box;
}
.aftersale-entry--active { opacity: 0.7; }
.aftersale-entry-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.aftersale-entry-title { font-size: 26rpx; font-weight: 600; color: var(--color-text); }
.aftersale-entry-hint { font-size: 22rpx; color: var(--color-text-placeholder); }
.aftersale-entry-arrow { font-size: 30rpx; color: var(--color-text-placeholder); }

.bottom-spacer { height: 24rpx; }
</style>
