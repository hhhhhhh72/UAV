<template>
  <view class="page-container">
    <u-nav-bar title="认证飞手" show-back right-text="申请认证" @back="goBack" @right="applyPilot" />
    <u-search v-model="searchText" placeholder="搜索认证飞手" @search="onSearch" />
    <view v-for="(item, i) in list" :key="i" class="card" @tap="goDetail(item)">
      <image :src="item.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="avatar" />
      <view class="info">
        <text class="name">{{ item.real_name || '认证飞手' }}</text>
        <text class="desc">{{ (item.cert_ids || []).length > 0 ? item.cert_ids.length + ' 项证书认证' : '认证飞手' }}</text>
        <text class="meta">{{ item.flight_hours || 0 }} 小时飞行 · 评分 {{ item.rating || '-' }}</text>
        <text v-if="item.bio" class="bio">{{ item.bio }}</text>
      </view>
      <text class="arrow">›</text>
    </view>
    <u-empty v-if="!list.length" description="暂无认证飞手" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, getStoredUser } from '../../utils/request'

const searchText = ref('')
const list = ref([])
const goBack = () => uni.navigateBack()
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
  try {
    const kw = searchText.value.trim()
    const res = await request({ url: '/api/v1/certified-pilots', data: { page: 1, page_size: 50, keyword: kw } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch { list.value = [] }
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
.card { display: flex; align-items: center; gap: 12px; margin: 8px 12px; background: var(--color-bg-card); border-radius: 10px; padding: 12px; }
.avatar { width: 48px; height: 48px; border-radius: 50%; background: var(--color-primary-light); flex-shrink: 0; }
.info { flex: 1; }
.name { font-size: 15px; font-weight: 600; color: var(--color-text); display: block; }
.desc { font-size: 13px; color: var(--color-primary); margin-top: 2px; display: block; }
.meta { font-size: 12px; color: var(--color-text-secondary); margin-top: 2px; display: block; }
.bio { font-size: 12px; color: var(--color-primary); margin-top: 4rpx; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 420rpx; }
.arrow { font-size: 18px; color: var(--color-text-placeholder); }
</style>
