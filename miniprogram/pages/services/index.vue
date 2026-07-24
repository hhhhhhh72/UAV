<template>
  <Layout :current="1">
    <view class="services-page">
      <view class="sticky-search">
        <view class="search-box">
          <image class="search-icon" src="/static/icons/search.svg" mode="aspectFit" />
          <input
            class="search-input"
            v-model="searchText"
            placeholder="搜索服务"
            placeholder-class="search-placeholder"
          />
        </view>
      </view>

      <view class="content-wrapper">
        <!-- 遍历服务分组 -->
        <view v-for="group in serviceGroups" :key="group.title" class="service-group-card">
          <view class="group-header">
            <view class="group-title">{{ group.title }}</view>
            <view class="group-subtitle">{{ group.subtitle }}</view>
          </view>

          <view class="service-grid">
            <view
              v-for="service in group.items"
              :key="service.id"
              class="service-grid-item"
              @tap="goToDetail(service.id)"
            >
              <view class="service-icon-large" :style="{ background: service.color }">
                <image class="service-icon-img" :src="service.icon" mode="aspectFit" />
              </view>
              <view class="service-title">{{ service.name }}</view>
            </view>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import { serviceGroupsConfig, serviceList } from '../../utils/services'

const searchText = ref('')

onLoad((options) => {
  if (options.keyword) {
    searchText.value = decodeURIComponent(options.keyword)
  }
})

const serviceGroups = computed(() => {
  const all = serviceList
  if (searchText.value) {
    const keyword = searchText.value.trim()
    const filtered = all.filter((s) => s.name.includes(keyword))
    return [{ title: '搜索结果', subtitle: 'Search Results', items: filtered }]
  }
  return serviceGroupsConfig.map((group) => ({
    ...group,
    items: group.ids.map((id) => all.find((s) => String(s.id) === String(id))).filter(Boolean)
  }))
})

const openExternal = (url) => {
  uni.navigateTo({ url: `/pages/webview/index?src=${encodeURIComponent(url)}` })
}

const goToDetail = (id) => {
  if (String(id) === 'flight') {
    openExternal('https://wx.zndkfx.com')
    return
  }
  if (String(id) === '8') {
    openExternal(
      'https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130&title=%E6%97%A0%E4%BA%BA%E6%9C%BA%E5%A4%96%E5%8D%96%E9%85%8D%E9%80%81&jyauthcode='
    )
    return
  }
  if (String(id) === '9') {
    uni.navigateTo({ url: '/pages/study/index' })
    return
  }
  uni.navigateTo({ url: `/pages/services/detail?id=${encodeURIComponent(String(id))}` })
}
</script>

<style scoped>
.services-page {
  background: #f7f8fa;
  min-height: 100vh;
}

.sticky-search {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #fff;
  padding: 12px 16px;
  border-bottom-left-radius: 24px;
  border-bottom-right-radius: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f7f8fa;
  border-radius: 12px;
  padding: 8px 12px;
}

.search-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
}

.search-input {
  flex: 1;
  font-size: 14px;
  color: #323233;
}

.search-placeholder {
  color: #969799;
}

.content-wrapper {
  padding: 12px;
}

.service-group-card {
  background: #fff;
  border-radius: 16px;
  padding: 16px 12px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.group-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding-left: 4px;
}

.group-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
}

.group-subtitle {
  display: none;
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px 4px;
}

.service-grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  transition: opacity 0.2s;
}

.service-grid-item:active {
  opacity: 0.6;
}

.service-icon-large {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-icon-img {
  width: 26px;
  height: 26px;
  filter: brightness(0) invert(1);
}

.service-title {
  font-size: 12px;
  font-weight: 400;
  color: #333;
  line-height: 1.3;
  white-space: nowrap;
  transform: scale(0.95);
}
</style>
