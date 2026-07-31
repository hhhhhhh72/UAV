<template>
  <view class="page-container">
    <van-nav-bar title="认证飞手" left-arrow @click-left="goBack" />
    <van-search v-model="searchText" placeholder="搜索认证飞手" shape="round" @search="onSearch" />
    <view v-for="(item, i) in list" :key="i" class="card" @tap="goDetail(item)">
      <image :src="item.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="avatar" />
      <view class="info">
        <text class="name">{{ item.name || item.title }}</text>
        <text class="desc">{{ item.cert_type || item.specialty || '认证飞手' }}</text>
        <text class="meta">{{ item.experience || '经验丰富' }} · {{ item.district || '' }}</text>
      </view>
      <text class="arrow">›</text>
    </view>
    <van-empty v-if="!list.length" description="暂无认证飞手" />
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
.page-container { min-height: 100vh; background: #f5f6f7; padding-bottom: 20px; }
.card { display: flex; align-items: center; gap: 12px; margin: 8px 12px; background: #fff; border-radius: 10px; padding: 12px; }
.avatar { width: 48px; height: 48px; border-radius: 50%; background: #e8f2fc; flex-shrink: 0; }
.info { flex: 1; }
.name { font-size: 15px; font-weight: 600; color: #1a1a1a; display: block; }
.desc { font-size: 13px; color: #1989fa; margin-top: 2px; display: block; }
.meta { font-size: 12px; color: #999; margin-top: 2px; display: block; }
.arrow { font-size: 18px; color: #ccc; }
</style>
