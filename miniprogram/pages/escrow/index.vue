<template>
  <view class="es-page">
    <!-- 加载失败必须显式报错：此前只在 catch 里清空流水，余额停留在初始 0，
         资金页会显示"余额 ¥0.00"，用户会以为钱没了。 -->
    <view v-if="loadError" class="es-error">
      <text class="es-error-title">账户信息加载失败</text>
      <text class="es-error-tip">当前余额不可信，请勿据此判断账户金额；点击下方按钮重试。</text>
      <view class="es-btn es-error-btn" hover-class="es-btn-hover" @tap="load">重新加载</view>
    </view>
    <template v-else>
    <view class="es-balance">
      <text class="es-balance-label">托管金余额（元）</text>
      <text class="es-balance-num">{{ (balanceFen / 100).toFixed(2) }}</text>
      <view class="es-frozen" v-if="frozenFen > 0">冻结中 ¥{{ (frozenFen / 100).toFixed(2) }}</view>
    </view>

    <view class="es-section">
      <text class="es-section-title">充值</text>
      <view class="es-form">
        <view class="es-row">
          <text class="es-row-label">金额（元）</text>
          <input class="es-input" v-model="amountYuan" type="digit" :placeholder="'单笔上限 ' + maxYuan" placeholder-class="es-ph" />
          <view class="es-btn" hover-class="es-btn-hover" :class="{ 'es-btn-loading': paying }" @tap="deposit">{{ paying ? '支付中…' : '充值' }}</view>
        </view>
        <text class="es-tip" v-if="payState === 'real'">微信支付：付款成功由微信通知服务端入账，余额稍后自动刷新（以服务端到账为准）。冻结/退款按订单流程自动处理。</text>
        <text class="es-tip" v-else-if="payState === 'sim'">模拟托管通道：充值即入账（无真实资金流；微信支付商户号配置后自动切换为真实支付）。冻结/退款按订单流程自动处理。</text>
        <text class="es-tip es-tip-warn" v-else>支付状态获取失败。为避免误入模拟通道，充值已暂停，请下拉刷新页面重试。</text>
      </view>
    </view>

    <view class="es-section">
      <text class="es-section-title">资金流水</text>
      <view v-if="!txs.length" class="es-empty">暂无流水记录</view>
      <view v-else>
        <view v-for="tx in txs" :key="tx.id" class="es-tx">
          <view class="es-tx-main">
            <text class="es-tx-type">{{ txTypeLabel(tx) }}</text>
            <text class="es-tx-time">{{ (tx.created_at || '').slice(0, 16).replace('T', ' ') }}</text>
            <text class="es-tx-sub">{{ txSub(tx) }}</text>
          </view>
          <text class="es-tx-amount" :class="{ minus: tx.tx_type === 'freeze' || tx.tx_type === 'release' }">
            {{ (tx.amount_fen / 100).toFixed(2) }} 元
          </text>
        </view>
      </view>
    </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request, getErrorMessage } from '../../utils/request'

const balanceFen = ref(0)
const frozenFen = ref(0)
const txs = ref([])
const amountYuan = ref('')
// loadError：加载失败时进入显式错误态，而不是把"没加载出来"渲染成余额 0。
const loadError = ref(false)
// realPayEnabled：后端是否已开通真实微信支付。由 GET /api/v1/payments/mine 探得：
// 未开通时该接口回 503，这是「未配置商户号」的确定信号，比在页面上写死文案可靠
// ——商户号一旦配上，页面无需改动即自动切成真实支付。
// payState 三态，**不能用布尔**：
//   'real'    服务端明确回答已开通微信支付
//   'sim'     服务端明确回答未开通（此时页面走模拟托管通道）
//   'unknown' 还没探到 / 探测失败 —— fail-closed，见 probeRealPay
// 为什么必须是三态：模拟通道（POST /api/v1/escrow/deposit）是**不加钱就能给自己加余额**的，
// 一旦真实支付已开通，把"探测失败"当成"未开通"就等于一次网络抖动就把用户送到免费通道上。
const payState = ref('unknown')
const realPayEnabled = computed(() => payState.value === 'real')
const paying = ref(false)
// 真实支付单笔上限 ¥50,000（服务端 service.MaxRechargeFen）；模拟通道沿用旧的 ¥200,000。
const maxYuan = computed(() => (payState.value === 'real' ? 50000 : 200000))

const txTypeLabel = (tx) => ({ deposit: '充值', freeze: '冻结', release: '学费结算', refund: '退款' }[tx.tx_type] || tx.tx_type || '-')
/* 流水副说明：让每笔钱的去向一目了然（结算=转给课程机构，退款=钱已回账） */
const txSub = (tx) => {
  if (tx.tx_type === 'deposit') {
    // 真实资金与平台内部记账必须能一眼分开，否则对账时说不清钱从哪来。
    return tx.channel === 'wechat' ? '微信支付到账' : '模拟充值入账'
  }
  return ({ freeze: '报名时冻结学费', release: '结业结算，已划转至课程机构', refund: '订单取消或报名失败，已退回余额' }[tx.tx_type] || '')
}

/* probeRealPay 探测后端是否已开通微信支付。
   服务端恒回 200 + enabled 字段（未开通不再用 5xx——5xx 会被后端统一脱敏成
   "internal server error"，客户端分不清"没开通"和"服务炸了"）。
   探测失败时**保持 unknown 而不是降级为 sim**：降级会让用户落到免费模拟通道。 */
async function probeRealPay() {
  try {
    const res = await request({ url: '/api/v1/payments/mine' })
    payState.value = res && res.enabled ? 'real' : 'sim'
  } catch (e) {
    payState.value = 'unknown'
  }
}

async function load() {
  try {
    const res = await request({ url: '/api/v1/escrow/mine' })
    const acc = (res && res.account) || {}
    balanceFen.value = acc.balance_fen || 0
    frozenFen.value = acc.frozen_fen || 0
    txs.value = Array.isArray(res && res.transactions) ? res.transactions : []
    loadError.value = false
  } catch (e) {
    // 失败时不再静默把余额留成 0（资金页显示"¥0.00"会误导用户以为钱丢了），
    // 而是进入错误态并给出重试入口。
    loadError.value = true
    txs.value = []
    uni.showToast({ title: getErrorMessage(e) || '账户信息加载失败', icon: 'none' })
  }
}

/* requestPayment 调起微信收银台（Promise 化：uni 的 API 是回调式的）。 */
function requestPayment(params) {
  return new Promise((resolve, reject) => {
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: params.timeStamp,
      nonceStr: params.nonceStr,
      package: params.package,
      signType: params.signType,
      paySign: params.paySign,
      success: resolve,
      // 用户主动取消也走 fail（errMsg 含 cancel）：不算失败，不弹报错。
      fail: reject
    })
  })
}

/* refreshAfterPay 付款成功后刷新余额。
   注意：**入账不在这里做**，由微信回调打到服务端完成，前端只负责稍后取一次真相。
   微信通知可能比 requestPayment 的 success 晚几十毫秒，故重试几次而不是只查一次。 */
async function refreshAfterPay() {
  for (let i = 0; i < 5; i++) {
    await load()
    if (balanceFen.value > 0) return
    await new Promise((r) => setTimeout(r, 800))
  }
}

/* payWithWeChat 走真实微信支付。返回 true=已处理，false=后端未开通，需回退模拟通道。 */
async function payWithWeChat(amountFen) {
  let prepay
  try {
    prepay = await request({
      url: '/api/v1/payments/wechat/prepay',
      method: 'POST',
      data: { amount_fen: amountFen },
      header: { 'Idempotency-Key': 'wx-pay-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8) }
    })
  } catch (e) {
    // 409 = 服务端明确回「微信支付未开通」（见 internal/httpapi/payments.go 的注释：
    // 预期状态不用 5xx，否则消息被脱敏且刷错误日志）。503 兼容旧部署。
    if (e && (e.statusCode === 409 || e.statusCode === 503)) return false
    throw e
  }
  try {
    await requestPayment((prepay && prepay.pay_params) || {})
  } catch (e) {
    const msg = (e && (e.errMsg || e.message)) || ''
    // 用户取消：静默返回，订单会由微信侧在有效期内自动关闭，不需要额外处理。
    if (msg.indexOf('cancel') >= 0) { uni.showToast({ title: '已取消支付', icon: 'none' }); return true }
    throw new Error(msg || '支付未完成')
  }
  uni.showToast({ title: '支付成功', icon: 'success' })
  amountYuan.value = ''
  await refreshAfterPay()
  return true
}

async function deposit() {
  const yuan = Number(amountYuan.value) || 0
  if (yuan <= 0) { uni.showToast({ title: '请输入金额', icon: 'none' }); return }
  if (yuan > maxYuan.value) { uni.showToast({ title: '单笔上限 ' + maxYuan.value + ' 元', icon: 'none' }); return }
  const amountFen = Math.round(yuan * 100)
  // fail-closed：支付能力未知时拒绝充值，绝不"猜一个通道"。
  if (payState.value === 'unknown') {
    uni.showToast({ title: '支付状态获取失败，请下拉刷新后重试', icon: 'none' })
    return
  }
  if (paying.value) return
  paying.value = true
  try {
    if (payState.value === 'real' && await payWithWeChat(amountFen)) return
    // 只有服务端**明确**回答"未开通"（payState==='sim'）才会走到这里。
    // 回退：模拟托管通道（后端未开通微信支付时）
    await request({
      url: '/api/v1/escrow/deposit',
      method: 'POST',
      data: { amount_fen: amountFen },
      // 充值是可重复同参操作：必须唯一幂等键，否则相同金额二次充值被服务端 24h 幂等去重拦截（余额不变）
      header: { 'Idempotency-Key': 'esc-dep-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8) }
    })
    uni.showToast({ title: '充值成功', icon: 'success' })
    amountYuan.value = ''
    load()
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '充值失败，请重试', icon: 'none' })
  } finally {
    paying.value = false
  }
}

onMounted(async () => {
  await probeRealPay()
  load()
})
</script>

<style>
page { background: var(--color-bg); }
.es-page { min-height: 100vh; padding: 24rpx; box-sizing: border-box; }
.es-balance { background: linear-gradient(160deg, #0a3a6b, #074d92); border-radius: 10px; padding: 40rpx 32rpx; color: #fff; }
.es-balance-label { display: block; font-size: 24rpx; opacity: .85; }
.es-balance-num { display: block; font-size: 64rpx; font-weight: 800; margin-top: 12rpx; }
.es-frozen { display: inline-block; margin-top: 12rpx; font-size: 22rpx; opacity: .85; }
.es-section { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 10px; padding: 24rpx; margin-top: 20rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.es-section-title { display: block; font-size: 28rpx; font-weight: 600; color: #17212B; margin-bottom: 16rpx; }
.es-row { display: flex; align-items: center; gap: 16rpx; }
.es-row-label { font-size: 26rpx; color: #667085; flex-shrink: 0; }
.es-input { flex: 1; background: #FAFAFA; border-radius: 24rpx; padding: 16rpx 24rpx; font-size: 28rpx; }
.es-ph { color: #98A2B3; }
.es-btn { flex-shrink: 0; height: 72rpx; border-radius: 36rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 600; display: flex; align-items: center; padding: 0 32rpx; }
.es-btn-hover { opacity: .9; }
/* 支付中：置灰并禁用点击（防重复拉起收银台，同一笔钱不能被发起两次） */
.es-btn-loading { background: #9aa7b5; }
/* 支付能力探测失败：橙色系＝需要注意且影响操作（与发布页的比例提示同一语义） */
.es-tip-warn { color: #B54708; }
.es-tip { display: block; font-size: 22rpx; color: #98A2B3; line-height: 1.6; margin-top: 16rpx; }
.es-empty { text-align: center; color: #98A2B3; font-size: 24rpx; padding: 40rpx 0; }
.es-tx { display: flex; justify-content: space-between; align-items: center; padding: 18rpx 0; border-bottom: 1rpx solid #F0F2F5; }
.es-tx:last-child { border-bottom: none; }
.es-tx-type { display: block; font-size: 26rpx; color: #17212B; }
.es-tx-time { display: block; font-size: 22rpx; color: #98A2B3; margin-top: 4rpx; }
.es-tx-sub { display: block; font-size: 20rpx; color: #98A2B3; margin-top: 2rpx; }
.es-tx-amount { font-size: 28rpx; font-weight: 700; color: #0B6B41; }
.es-tx-amount.minus { color: #D92D20; }
.es-error { background: #fff; border: 1rpx solid #FECACA; border-radius: 10px; padding: 32rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.es-error-title { display: block; font-size: 30rpx; font-weight: 700; color: #B42318; }
.es-error-tip { display: block; font-size: 24rpx; color: #667085; line-height: 1.6; margin-top: 12rpx; }
.es-error-btn { display: inline-flex; margin-top: 20rpx; }
</style>
