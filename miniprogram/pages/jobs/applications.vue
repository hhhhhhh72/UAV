<template>
  <view class="page-container">
    <u-nav-bar title="我的投递" show-back @back="goBack" />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>

    <!-- Empty -->
    <view v-else-if="!list.length" class="empty-state">
      <u-empty description="暂无投递记录" />
    </view>

    <!-- 投递列表 -->
    <view v-else class="apply-list">
      <view v-for="item in list" :key="item.id" class="apply-card">
        <view class="apply-row1">
          <text class="apply-job">职位 {{ item.job_id }}</text>
          <u-tag :type="statusTag(item.status)" size="mini">{{ statusLabel[item.status] || item.status }}</u-tag>
        </view>
        <text class="apply-time">{{ formatDate(item.created_at) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { request } from '../../utils/request'

const goBack = () => uni.navigateBack()
const list = ref([])
const loading = ref(false)

const statusLabel = { submitted: '已投递', viewed: '已查看', interviewing: '面试中', offered: '已录用', rejected: '未通过', withdrawn: '已撤回' }
const statusTag = (s) => ({ submitted: 'primary', viewed: 'warning', interviewing: 'warning', offered: 'success', rejected: 'danger', withdrawn: 'info' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const load = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/applications' })
    list.value = Array.isArray(res) ? res : ((res && res.data) || [])
  } catch (e) {
    uni.showToast({ title: '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.page-container { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 60px 0; }
.loading-text { font-size: 24rpx; color: var(--color-text-secondary); }
.empty-state { padding-top: 60px; }
.apply-list { padding: 8px 12px; }
.apply-card { background: var(--color-bg-card); border-radius: 12rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 1px 3px rgba(0,0,0,.05); }
.apply-row1 { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; }
.apply-job { font-size: 28rpx; font-weight: 600; color: var(--color-text); }
.apply-time { font-size: 22rpx; color: var(--color-text-placeholder); margin-top: 12rpx; display: block; }
</style>
