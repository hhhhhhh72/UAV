<template>
  <view class="page-container">
    <u-nav-bar title="认证飞手" show-back @back="goBack" />
    <u-search v-model="searchText" placeholder="搜索认证飞手" @search="onSearch" />
    <view v-for="(item, i) in list" :key="i" class="card" @tap="goDetail(item)">
      <image :src="item.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="avatar" />
      <view class="info">
        <text class="name">{{ item.name || item.title }}</text>
        <text class="desc">{{ item.cert_type || item.specialty || '认证飞手' }}</text>
        <text class="meta">{{ item.experience || '经验丰富' }} · {{ item.district || '' }}</text>
      </view>
      <text class="arrow">›</text>
    </view>
    <u-empty v-if="!list.length" description="暂无认证飞手" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

const searchText = ref('')
const list = ref([])
const goBack = () => uni.navigateBack()
const goDetail = (item) => uni.navigateTo({ url: '/pages/experts/detail?id=' + item.id })
const onSearch = () => fetchData()

const fetchData = async () => {
  try {
    const res = await request({ url: '/api/v1/experts', data: { page: 1, page_size: 50 } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch { list.value = [] }
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
.arrow { font-size: 18px; color: var(--color-text-placeholder); }
</style>
