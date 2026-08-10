<template>
  <view class="ent-list-page">
    <u-nav-bar title="入驻企业" show-back @back="goBack" />

    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="28rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <view v-else-if="list.length === 0" class="empty-state">
      <u-empty description="暂无入驻企业" />
      <text class="empty-note">企业完成入驻审核后将在此公示</text>
    </view>

    <view v-else class="list-body">
      <u-cell-group inset>
        <u-cell
          v-for="e in list"
          :key="e.id"
          :title="e.name"
          :label="'入驻时间 ' + formatDate(e.created_at)"
        >
          <template #value>
            <u-tag type="success" size="mini">已认证</u-tag>
          </template>
        </u-cell>
      </u-cell-group>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'

const loading = ref(false)
const list = ref([])

const goBack = () => uni.navigateBack()

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

onLoad(async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/enterprises/public' })
    list.value = Array.isArray(res) ? res : []
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.ent-list-page { min-height: 100vh; background: var(--color-bg); }
.loading-state { display: flex; justify-content: center; padding: 80rpx 0; }
.loading-inline { display: flex; align-items: center; gap: 16rpx; color: var(--color-text-secondary); }
.empty-state { padding: 100rpx 0; text-align: center; }
.empty-note { display: block; margin-top: 16rpx; font-size: 24rpx; color: var(--color-text-secondary); }
.list-body { padding: 24rpx; }
</style>
