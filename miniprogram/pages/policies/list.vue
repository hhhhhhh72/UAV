<template>
  <view class="page-container">
    <u-nav-bar title="政策法规" show-back @back="goBack" />
    <u-sticky>
      <u-search v-model="searchText" placeholder="搜索政策法规" @search="onSearch" />
    </u-sticky>
    <view v-for="(item, i) in list" :key="i" class="card" @tap="goDetail(item)">
      <text class="card-title">{{ item.title }}</text>
      <text class="card-desc">{{ item.description || '' }}</text>
      <view class="card-foot">
        <text class="card-date">{{ item.publish_date || '' }}</text>
        <text class="card-arrow">›</text>
      </view>
    </view>
    <u-empty v-if="!list.length" description="暂无政策法规" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

const searchText = ref('')
const list = ref([])
const goBack = () => uni.navigateBack()
const onSearch = () => fetchData()
const goDetail = (item) => uni.navigateTo({ url: '/pages/compliance/knowledge?id=' + item.id })

const fetchData = async () => {
  try {
    const res = await request({ url: '/api/v1/compliance-docs', data: { page: 1, page_size: 50, keyword: searchText.value } })
    list.value = (Array.isArray(res) ? res : (res.data || []))
  } catch { list.value = [] }
}
onLoad(() => fetchData())
</script>

<style scoped>
.page-container { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }
.card { margin: 8px 12px; background: var(--color-bg-card); border-radius: 10px; padding: 14px; }
.card-title { font-size: 15px; font-weight: 600; color: var(--color-text); display: block; margin-bottom: 6px; }
.card-desc { font-size: 13px; color: var(--color-text); opacity: 0.8; display: block; line-height: 1.4; margin-bottom: 8px; }
.card-foot { display: flex; justify-content: space-between; font-size: 12px; color: var(--color-text-secondary); }
.card-arrow { font-size: 16px; color: var(--color-text-placeholder); }
</style>
