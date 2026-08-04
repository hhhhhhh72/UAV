<template>
  <view class="service-detail-page" v-if="service">
    <u-nav-bar
      :title="service.title || '服务详情'"
      show-back
      @back="goBack"
    />

    <view class="detail-content">
      <!-- 服务头部信息 -->
      <view class="service-header">
        <view class="service-icon-big" :style="{ background: service.color || service.gradient || 'var(--color-primary)' }">
          <text class="service-icon-text">{{ service.icon || '服' }}</text>
        </view>
        <text class="service-title">{{ service.title || service.name }}</text>
        <text class="service-desc">{{ service.description || service.subtitle || service.slogan || '' }}</text>
      </view>

      <!-- 服务介绍区域 -->
      <view class="section-card" v-if="service.intro || service.description">
        <text class="section-label">服务介绍</text>
        <text class="section-text">{{ service.intro || service.description }}</text>
      </view>

      <!-- 服务项目区域 -->
      <view class="section-card" v-if="service.projects && service.projects.length > 0">
        <text class="section-label">服务项目</text>
        <view class="project-grid">
          <view
            v-for="(item, index) in service.projects"
            :key="index"
            class="project-item"
          >
            <text class="project-name">{{ item.name || item }}</text>
          </view>
        </view>
      </view>

      <!-- 服务优势区域 -->
      <view class="section-card" v-if="service.advantages && service.advantages.length > 0">
        <text class="section-label">服务优势</text>
        <view class="advantage-list">
          <view
            v-for="(adv, index) in service.advantages"
            :key="index"
            class="advantage-item"
          >
            <u-icon name="check" size="26rpx" color="var(--color-success)" />
            <text class="adv-text">{{ adv }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="action-bar">
      <u-button
        type="primary"
        block
        round
        @click="onApply"
      >
        {{ actionButtonText }}
      </u-button>
    </view>
  </view>

  <!-- 空/错误状态 -->
  <view v-else-if="!loading" class="empty-page">
    <u-nav-bar title="服务详情" show-back @back="goBack" />
    <u-empty description="服务内容待发布" />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'

const service = ref(null)
const loading = ref(true)

onLoad(async (options) => {
  const id = String(options.id || '')
  try {
    const res = await request({ url: '/api/services/config' })
    const config = (res && res.data) || res || {}
    const entry = config[id]
    if (entry && typeof entry === 'object' && Object.keys(entry).length > 0) {
      service.value = entry
    } else if (config._home && Array.isArray(config._home.displayServices)) {
      const s = config._home.displayServices.find((x) => String(x.id) === id)
      if (s) {
        service.value = { id: s.id, title: s.name, name: s.name, description: s.description || '' }
      }
    }
  } catch (e) {
    service.value = null
  } finally {
    loading.value = false
  }
  if (service.value) {
    uni.setNavigationBarTitle({ title: service.value.title || service.value.name || '服务详情' })
  }
})

const actionButtonText = computed(() => {
  if (!service.value) return '立即申请'
  return '立即申请'
})

const onApply = () => {
  uni.showToast({ title: '申请功能即将上线', icon: 'none', duration: 1500 })
}

const goBack = () => {
  uni.navigateBack()
}
</script>

<style scoped>
.service-detail-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 80px;
}

.empty-page {
  min-height: 100vh;
  background: var(--color-bg);
}

.detail-content {
  padding: 12px 0;
}

.service-header {
  background: var(--color-bg-card);
  padding: 32px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.service-icon-big {
  width: 72px;
  height: 72px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.service-icon-text {
  font-size: 32px;
  font-weight: 600;
  color: #ffffff;
}

.service-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-text);
  display: block;
  margin-bottom: 8px;
}

.service-desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  display: block;
}

.section-card {
  background: var(--color-bg-card);
  margin: 12px 16px;
  padding: 16px;
  border-radius: 12px;
}

.section-label {
  font-size: 16px;
  font-weight: 700;
  color: var(--color-text);
  display: block;
  margin-bottom: 12px;
}

.section-text {
  font-size: 14px;
  color: var(--color-text);
  opacity: 0.85;
  line-height: 1.8;
  display: block;
}

.project-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.project-item {
  background: var(--color-bg);
  border-radius: 8px;
  padding: 10px 16px;
}

.project-name {
  font-size: 13px;
  color: var(--color-text);
}

.advantage-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.advantage-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.adv-text {
  font-size: 14px;
  color: var(--color-text);
  opacity: 0.85;
}

.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: var(--color-bg-card);
  border-top: 1px solid var(--color-border);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}
</style>
