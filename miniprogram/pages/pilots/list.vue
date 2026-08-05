<template>
  <view class="page-container">
    <u-nav-bar title="认证飞手" show-back right-text="申请认证" @back="goBack" @right="applyPilot" />

    <!-- 头部品牌区 -->
    <view class="hero">
      <view class="hero-glow" />
      <text class="hero-title">协会认证飞手名录</text>
      <text class="hero-sub">经协会审核认证的持证飞手 · 可承接巡检 / 测绘 / 植保 / 吊运等作业</text>
      <view class="hero-stats">
        <view class="hero-stat">
          <text class="hero-num">{{ list.length }}</text>
          <text class="hero-label">认证飞手</text>
        </view>
        <view class="hero-divider" />
        <view class="hero-stat">
          <text class="hero-num">{{ totalHours }}</text>
          <text class="hero-label">累计飞行小时</text>
        </view>
      </view>
    </view>

    <!-- 搜索 -->
    <u-search v-model="searchText" placeholder="搜索认证飞手" @search="onSearch" />

    <!-- 飞手名片列表 -->
    <view v-for="(item, i) in list" :key="i" class="card" @tap="goDetail(item)">
      <view class="avatar-wrap">
        <image :src="item.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="avatar" />
        <view class="cert-badge" />
      </view>
      <view class="info">
        <view class="name-row">
          <text class="name">{{ item.real_name || '认证飞手' }}</text>
          <text class="rating" v-if="item.rating">★ {{ item.rating }}</text>
        </view>
        <view class="stats">
          <view class="stat-item">
            <text class="stat-num">{{ (item.cert_ids || []).length }}</text>
            <text class="stat-label">证书</text>
          </view>
          <view class="stat-item">
            <text class="stat-num">{{ item.flight_hours || 0 }}</text>
            <text class="stat-label">飞行小时</text>
          </view>
          <view class="stat-item">
            <text class="stat-num">{{ item.completed_jobs || 0 }}</text>
            <text class="stat-label">完成作业</text>
          </view>
        </view>
        <view v-if="item.bio" class="bio-tags">
          <text v-for="(b, bi) in bioList(item.bio)" :key="bi" class="bio-tag">{{ b }}</text>
        </view>
      </view>
      <text class="arrow">›</text>
    </view>

    <view v-if="!loading && !list.length" class="empty-wrap">
      <u-empty description="暂无认证飞手" />
    </view>

    <!-- 加载态 -->
    <view v-if="loading && !list.length" class="loading-wrap">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'

const searchText = ref('')
const list = ref([])
const loading = ref(false)
const goBack = () => uni.navigateBack()

// 头部统计：当前名录累计飞行小时
const totalHours = computed(() => list.value.reduce((s, p) => s + (p.flight_hours || 0), 0))

// 擅长领域按分隔符拆成标签（/、，、空格）
const bioList = (bio) => String(bio || '').split(/[/，,、\s]+/).filter(Boolean).slice(0, 3)

// 飞手无独立详情页：点击展示信息弹层
const goDetail = (item) => {
  const lines = [
    `证书认证 ${(item.cert_ids || []).length} 项`,
    `飞行时长 ${item.flight_hours || 0} 小时`,
    `完成作业 ${item.completed_jobs || 0} 单`,
    `评分 ${item.rating || '-'}`,
  ]
  if (item.bio) lines.push(`擅长领域：${item.bio}`)
  uni.showModal({
    title: item.real_name || '认证飞手',
    content: lines.join('\n'),
    showCancel: false,
    confirmText: '知道了'
  })
}
const onSearch = () => fetchData()

const fetchData = async () => {
  loading.value = true
  try {
    const kw = searchText.value.trim()
    const res = await request({ url: '/api/v1/certified-pilots', data: { page: 1, page_size: 50, keyword: kw } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch { list.value = [] } finally {
    loading.value = false
  }
}

// ---- 申请认证 / 我的状态 ----
const applyPilot = async () => {
  if (!getStoredUser()) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }
  let mine = null
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    mine = res && res.data ? res.data : res
  } catch (e) {}
  if (mine && mine.id) {
    const label = { pending: '待审核', approved: '已认证', rejected: '未通过' }[mine.status] || mine.status
    uni.showModal({ title: '我的飞手认证', content: `当前状态：${label}\n${mine.real_name || ''}`, showCancel: false, confirmText: '知道了' })
    return
  }
  // 未申请：进入完整申请表单页（姓名/身份证/时长/简介）
  uni.navigateTo({ url: '/pages/pilots/apply' })
}

onLoad(() => fetchData())
</script>

<style scoped>
.page-container { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }

/* 头部品牌区：深蓝渐变 + 光晕 */
.hero {
  position: relative;
  overflow: hidden;
  padding: 36rpx 32rpx 40rpx;
  background: var(--color-primary-deep) 100%);
}
.hero-glow {
  position: absolute;
  top: -120rpx;
  right: -80rpx;
  width: 360rpx;
  height: 360rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(29,212,168,.15), transparent 65%);
  pointer-events: none;
}
.hero-title { font-size: 40rpx; font-weight: 700; color: #fff; display: block; position: relative; }
.hero-sub { font-size: 24rpx; color: rgba(255,255,255,.7); margin-top: 12rpx; display: block; line-height: 1.6; position: relative; }
.hero-stats { display: flex; align-items: center; gap: 48rpx; margin-top: 28rpx; position: relative; }
.hero-stat { display: flex; align-items: baseline; gap: 12rpx; }
.hero-num { font-size: 44rpx; font-weight: 800; color: #fff; }
.hero-label { font-size: 22rpx; color: rgba(255,255,255,.65); }
.hero-divider { width: 2rpx; height: 40rpx; background: rgba(255,255,255,.25); }

/* 搜索 */
.u-search-wrap { padding: 20rpx 24rpx 0; }

/* 飞手名片卡 */
.card {
  display: flex;
  gap: 20rpx;
  margin: 20rpx 24rpx;
  background: var(--color-bg-card);
  border-radius: 8px;
  padding: 28rpx 24rpx;
  box-shadow: 0 3px 12px rgba(16,24,40,.05);
}
.avatar-wrap { position: relative; flex-shrink: 0; }
.avatar { width: 112rpx; height: 112rpx; border-radius: 50%; background: var(--color-primary-light); }
.cert-badge {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 32rpx;
  height: 32rpx;
  border-radius: 50%;
  background: var(--color-success);
  border: 4rpx solid #fff;
}
.cert-badge::after {
  content: '';
  position: absolute;
  left: 9rpx;
  top: 7rpx;
  width: 10rpx;
  height: 14rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: rotate(45deg);
}
.info { flex: 1; min-width: 0; }
.name-row { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; }
.name { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.rating { font-size: 24rpx; font-weight: 600; color: var(--color-accent-deep); }
.stats { display: flex; gap: 40rpx; margin-top: 20rpx; }
.stat-item { display: flex; align-items: baseline; gap: 8rpx; }
.stat-num { font-size: 30rpx; font-weight: 700; color: var(--color-primary); }
.stat-label { font-size: 20rpx; color: var(--color-text-secondary); }
.bio-tags { display: flex; flex-wrap: wrap; gap: 10rpx; margin-top: 16rpx; }
.bio-tag {
  font-size: 20rpx;
  padding: 4rpx 16rpx;
  border-radius: 4px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: 600;
}
.arrow { font-size: 28rpx; color: var(--color-text-placeholder); align-self: center; }

/* 状态 */
.loading-wrap { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 80px 0; }
.loading-text { font-size: 24rpx; color: var(--color-text-secondary); }
.empty-wrap { padding-top: 60px; }
</style>
