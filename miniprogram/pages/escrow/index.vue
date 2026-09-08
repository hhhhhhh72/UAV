<template>
  <view class="es-page">
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
          <input class="es-input" v-model="amountYuan" type="digit" placeholder="单笔上限 10000" placeholder-class="es-ph" />
          <view class="es-btn" hover-class="es-btn-hover" @tap="deposit">充值</view>
        </view>
        <text class="es-tip">模拟托管通道：充值即入账（无真实资金流；真实微信支付接入后由支付校验替代）。冻结/退款按订单流程自动处理。</text>
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
          </view>
          <text class="es-tx-amount" :class="{ minus: tx.tx_type === 'freeze' || tx.tx_type === 'release' || tx.tx_type === 'refund' }">
            {{ (tx.amount_fen / 100).toFixed(2) }} 元
          </text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { request, getErrorMessage } from '../../utils/request'

const balanceFen = ref(0)
const frozenFen = ref(0)
const txs = ref([])
const amountYuan = ref('')

const txTypeLabel = (tx) => ({ deposit: '充值', freeze: '冻结', release: '释放', refund: '退款' }[tx.tx_type] || tx.tx_type || '-')

async function load() {
  try {
    const res = await request({ url: '/api/v1/escrow/mine' })
    const acc = (res && res.account) || {}
    balanceFen.value = acc.balance_fen || 0
    frozenFen.value = acc.frozen_fen || 0
    txs.value = Array.isArray(res && res.transactions) ? res.transactions : []
  } catch (e) {
    txs.value = []
  }
}

async function deposit() {
  const yuan = Number(amountYuan.value) || 0
  if (yuan <= 0) { uni.showToast({ title: '请输入金额', icon: 'none' }); return }
  if (yuan > 10000) { uni.showToast({ title: '单笔上限 10000 元', icon: 'none' }); return }
  try {
    await request({
      url: '/api/v1/escrow/deposit',
      method: 'POST',
      data: { amount_fen: Math.round(yuan * 100) },
      // 充值是可重复同参操作：必须唯一幂等键，否则相同金额二次充值被服务端 24h 幂等去重拦截（余额不变）
      header: { 'Idempotency-Key': 'esc-dep-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8) }
    })
    uni.showToast({ title: '充值成功', icon: 'success' })
    amountYuan.value = ''
    load()
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '充值失败，请重试', icon: 'none' })
  }
}

onMounted(load)
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
.es-tip { display: block; font-size: 22rpx; color: #98A2B3; line-height: 1.6; margin-top: 16rpx; }
.es-empty { text-align: center; color: #98A2B3; font-size: 24rpx; padding: 40rpx 0; }
.es-tx { display: flex; justify-content: space-between; align-items: center; padding: 18rpx 0; border-bottom: 1rpx solid #F0F2F5; }
.es-tx:last-child { border-bottom: none; }
.es-tx-type { display: block; font-size: 26rpx; color: #17212B; }
.es-tx-time { display: block; font-size: 22rpx; color: #98A2B3; margin-top: 4rpx; }
.es-tx-amount { font-size: 28rpx; font-weight: 700; color: #0B6B41; }
.es-tx-amount.minus { color: #D92D20; }
</style>
