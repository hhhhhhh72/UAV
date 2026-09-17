<template>
  <view class="auth-page">
    <u-nav-bar title="飞手认证" show-back @back="goBack" />

    <!-- 加载中 -->
    <view v-if="loading" class="state-panel">
      <u-loading size="28rpx" />
      <text class="state-desc">加载中...</text>
    </view>

    <!-- 错误态 -->
    <view v-else-if="error" class="state-panel">
      <view class="status-icon status-icon--err">!</view>
      <text class="status-title">加载失败</text>
      <text class="status-desc">飞手认证信息获取失败，请检查网络后重试</text>
      <view class="retry-btn" hover-class="tap-fade" @tap="fetchMine">重新加载</view>
    </view>

    <!-- 认证状态卡：已认证 / 审核中 / 未通过 三种真实状态共用一张卡 -->
    <view v-else-if="record" class="status-card">
      <view class="status-icon" :class="stateIconClass">{{ stateIcon }}</view>
      <text class="status-title">{{ stateTitle }}</text>
      <text class="status-desc">{{ stateDesc }}</text>

      <!-- 身份信息：未通过时不展示——此时重点是驳回原因与重新提交 -->
      <view v-if="record.status !== 'rejected'" class="identity-rows">
        <view class="identity-row">
          <text class="identity-label">真实姓名</text>
          <text class="identity-value">{{ record.real_name || '—' }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">认证地区</text>
          <text class="identity-value">{{ record.region || '—' }}</text>
        </view>
        <view class="identity-row">
          <text class="identity-label">认证状态</text>
          <text class="identity-value">{{ statusText }}</text>
        </view>
      </view>

      <view v-if="record.status === 'rejected'" class="apply-btn" hover-class="tap-fade" @tap="goApply">重新提交申请</view>
    </view>

    <!-- 空态：无飞手认证记录，引导提交申请 -->
    <view v-else class="status-card">
      <view class="status-icon">实</view>
      <text class="status-title">尚未完成飞手认证</text>
      <text class="status-desc">提交飞手认证申请并通过审核后，即可展示在认证飞手名录中</text>
      <view class="apply-btn" hover-class="tap-fade" @tap="goApply">去提交飞手认证</view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

// 飞手认证状态页：数据来自 GET /api/v1/certified-pilots/mine。
//
// ⚠️ 本页此前叫「实名认证」，但它展示的**从来就是飞手认证**的结果——
// 平台并没有账号级的实名认证（users 表没有 real_name/id_card，也没有任何身份核验服务，
// 详见 docs/接口文档/实名认证接入方案.md，状态为"方案待评审"、从未实现）。
// 两者不是一回事：实名认证回答"你是你"（姓名+证号比对），飞手认证回答"你有资质"
//（CAAC/AOPA/UTC 证书 + 飞行小时）。沿用旧名字会让没有证书的用户
// 一直被引导去申请一个他做不到的"实名认证"，所以按实际含义改名。
const loading = ref(true)
const error = ref(false)
// 任意状态的飞手认证记录（pending / approved / rejected）。
// 此前这里只让 approved 进 record，另外两种状态被当成"没有记录"——
// 于是提交后待审核的用户（以及被驳回的用户）看到的都是"尚未完成实名认证 + 去提交飞手认证"，
// 会被引导去重复提交；statusText 里的 pending/rejected 分支也因此成了永远走不到的死代码。
const record = ref(null)

const statusText = computed(() => {
  const map = { approved: '已认证', pending: '审核中', rejected: '未通过' }
  return (record.value && map[record.value.status]) || '已认证'
})

// 状态卡三态文案（图标 / 标题 / 说明），与后端 certified_pilots.status 一一对应
const stateIcon = computed(() => {
  const map = { approved: '实', pending: '审', rejected: '!' }
  return map[record.value && record.value.status] || '实'
})
const stateIconClass = computed(() => {
  const s = record.value && record.value.status
  return { 'status-icon--pending': s === 'pending', 'status-icon--err': s === 'rejected' }
})
const stateTitle = computed(() => {
  const map = { approved: '飞手认证已完成', pending: '飞手认证审核中', rejected: '飞手认证未通过' }
  return map[record.value && record.value.status] || '尚未完成飞手认证'
})
const stateDesc = computed(() => {
  const s = record.value && record.value.status
  if (s === 'pending') return '飞手认证申请已提交，协会审核通过后即收录进认证飞手名录'
  if (s === 'rejected') {
    return record.value.reject_reason
      ? '驳回原因：' + record.value.reject_reason
      : '申请未通过审核，可补充材料后重新提交'
  }
  return '资质核验通过，您已收录进认证飞手名录'
})

const fetchMine = async () => {
  loading.value = true
  error.value = false
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const mine = (res && res.data) || res || null
    record.value = mine && mine.id ? mine : null
  } catch (e) {
    record.value = null
    error.value = true
  } finally {
    loading.value = false
  }
}

const goApply = () => {
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

const goBack = () => uni.navigateBack()

onShow(() => { fetchMine() })
</script>

<style scoped>
.auth-page { min-height: 100vh; background: var(--color-bg); padding-bottom: 40rpx; }

/* ── 状态卡 ── */
.status-card {
  margin: 28rpx 24rpx 0;
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 40rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  box-shadow: 0 3px 12px rgba(16,24,40,.05);
}
.status-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  background: var(--color-success, #16A34A);
}
.status-icon--err { background: #D92D20; }
/* 审核中：与项目强调橙同档（--color-accent-deep），区别于已认证的绿、未通过的红 */
.status-icon--pending { background: var(--color-accent-deep, #E96012); }
.status-title { font-size: 32rpx; font-weight: 700; color: var(--color-text); }
.status-desc { font-size: 24rpx; color: var(--color-text-secondary); text-align: center; line-height: 1.6; }

/* 身份信息展示 */
.identity-rows {
  margin-top: 16rpx;
  width: 100%;
  background: #FAFAFA;
  border-radius: 8px;
  padding: 8rpx 24rpx;
}
.identity-row { display: flex; justify-content: space-between; align-items: center; height: 72rpx; }
.identity-row + .identity-row { border-top: 1rpx solid var(--color-border, #EEF0F2); }
.identity-label { font-size: 24rpx; color: var(--color-text-secondary); }
.identity-value { font-size: 26rpx; color: var(--color-text); font-weight: 600; }

/* 加载 / 错误态 */
.state-panel {
  min-height: 480rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  padding: 40rpx;
}
.state-desc { font-size: 24rpx; color: var(--color-text-secondary); }
.retry-btn,
.apply-btn {
  margin-top: 20rpx;
  padding: 14rpx 56rpx;
  border-radius: 999rpx;
  background: var(--color-primary);
  color: #fff;
  font-size: 26rpx;
  font-weight: 600;
}
.tap-fade { opacity: 0.8; }
</style>
