<template>
  <view class="case-detail-page" v-if="caseItem">
    <view class="media-section">
      <swiper 
        class="media-swiper" 
        circular 
        indicator-dots 
        indicator-color="rgba(255,255,255,0.5)" 
        indicator-active-color="#fff"
      >
        <swiper-item v-for="(media, index) in caseItem.media" :key="index">
          <image v-if="media.type === 'image'" :src="media.url" mode="aspectFill" class="media-content" />
          <video v-else :src="media.url" class="media-content" controls></video>
        </swiper-item>
      </swiper>
    </view>

    <view class="detail-container">
      <view class="case-header">
        <view class="case-title">{{ caseItem.title }}</view>
        <view class="case-meta">
          <text class="tag">{{ caseItem.service }}</text>
          <text class="date">{{ caseItem.date }}</text>
        </view>
      </view>

      <view class="info-card">
        <view class="info-row">
          <text class="label">项目地点：</text>
          <text class="value">{{ caseItem.location }}</text>
        </view>
        <view class="info-row">
          <text class="label">项目类型：</text>
          <text class="value">{{ caseItem.service }}</text>
        </view>
      </view>

      <view class="section">
        <view class="section-title">案例介绍</view>
        <view class="section-content">
          <text class="desc-text">{{ caseItem.description }}</text>
        </view>
      </view>

      <view class="section" v-if="caseItem.highlights">
        <view class="section-title">项目亮点</view>
        <view class="highlight-list">
          <view v-for="(h, i) in caseItem.highlights" :key="i" class="highlight-item">
            <text class="check-icon">✓</text>
            <text>{{ h }}</text>
          </view>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getCaseById } from '../../utils/cases'
import { request } from '../../utils/request'

const caseItem = ref(null)

onLoad(async (options) => {
  const id = options.id || '1'
  try {
    const res = await request({ url: '/api/cases', data: { id } })
    const item = res?.data?.[0] || (Array.isArray(res) ? res[0] : res)
    if (item && item.title) {
      caseItem.value = item
    } else {
      caseItem.value = getCaseById(id)
    }
  } catch (e) {
    caseItem.value = getCaseById(id)
  }
  if (caseItem.value) {
    uni.setNavigationBarTitle({ title: '案例详情' })
  }
})
</script>

<style scoped>
.case-detail-page {
  min-height: 100vh;
  background: #fff;
}

.media-section {
  width: 100%;
  height: 240px;
  background: #000;
}

.media-swiper {
  width: 100%;
  height: 100%;
}

.media-content {
  width: 100%;
  height: 100%;
}

.detail-container {
  padding: 20px;
}

.case-header {
  margin-bottom: 20px;
}

.case-title {
  font-size: 22px;
  font-weight: bold;
  color: #1a1a1a;
  line-height: 1.4;
  margin-bottom: 12px;
}

.case-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tag {
  font-size: 12px;
  padding: 2px 8px;
  background: #f0f4ff;
  color: #667eea;
  border-radius: 4px;
}

.date {
  font-size: 12px;
  color: #969799;
}

.info-card {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
}

.info-row {
  display: flex;
  margin-bottom: 8px;
  font-size: 14px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-row .label {
  color: #969799;
  width: 80px;
}

.info-row .value {
  color: #323233;
  flex: 1;
}

.section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 18px;
  font-weight: bold;
  color: #1a1a1a;
  margin-bottom: 12px;
  padding-left: 12px;
  border-left: 4px solid #667eea;
  line-height: 1;
}

.section-content {
  font-size: 15px;
  color: #4b4c4d;
  line-height: 1.8;
}

.highlight-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.highlight-item {
  display: flex;
  gap: 10px;
  background: #f2f4ff;
  padding: 12px;
  border-radius: 8px;
  font-size: 14px;
  color: #323233;
  line-height: 1.4;
}

.check-icon {
  color: #667eea;
  font-weight: bold;
}
</style>
