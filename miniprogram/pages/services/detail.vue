<template>
  <view class="service-detail-page" v-if="service">
    <van-nav-bar
      :title="service.title || '服务详情'"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <view class="detail-content">
      <!-- 服务头部信息 -->
      <view class="service-header">
        <view class="service-icon-big" :style="{ background: service.color || service.gradient || '#1677ff' }">
          <van-icon :name="service.icon || 'shop-o'" size="28" color="#ffffff" />
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
            <van-icon name="checked" size="14" color="#07c160" />
            <text class="adv-text">{{ adv }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="action-bar">
      <van-button
        type="primary"
        block
        round
        @tap="onApply"
      >
        {{ actionButtonText }}
      </van-button>
    </view>
  </view>

  <!-- 空/错误状态 -->
  <view v-else class="empty-page">
    <van-nav-bar title="服务详情" fixed placeholder left-arrow @click-left="goBack" />
    <van-empty description="服务信息加载失败" image="error" />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const service = ref(null)

// 服务数据映射
const serviceDataMap = {
  '6': {
    id: '6',
    title: '飞手培训服务',
    name: '飞手培训服务',
    subtitle: '专业培训 证书认证',
    color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)',
    icon: 'medal-o',
    intro: '提供CAAC执照培训、UTC认证、人社认证等无人机操控员资格培训服务，浙南闽北地区最早具备民航局认定资质的培训机构。',
    projects: [
      { name: 'CAAC执照培训' },
      { name: 'UTC认证培训' },
      { name: '人社认证' },
      { name: '实操教学' },
    ],
    advantages: [
      '民航局官方授权考点',
      '3000平米标准训练场地',
      '高精度模拟飞行系统',
      '资深教员团队精准施教',
      '通过率行业领先',
    ],
  },
  '12': {
    id: '12',
    title: '维修服务',
    name: '维修服务',
    subtitle: '专业维修 原厂配件',
    color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)',
    icon: 'setting-o',
    intro: '提供专业的无人机维修、定期保养、故障诊断等服务，使用正品配件，确保设备安全可靠。',
    projects: [
      { name: '故障诊断' },
      { name: '硬件维修' },
      { name: '定期保养' },
      { name: '配件更换' },
    ],
    advantages: [
      '官方授权维修中心',
      '正品原厂配件',
      '专业技术团队',
      '维修质保承诺',
      '快速响应服务',
    ],
  },
}

onLoad((options) => {
  const id = String(options.id || '1')
  service.value = serviceDataMap[id]
  if (service.value) {
    uni.setNavigationBarTitle({ title: service.value.title })
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
  background: #f7f8fa;
  padding-bottom: 80px;
}

.empty-page {
  min-height: 100vh;
  background: #f7f8fa;
}

.detail-content {
  padding: 12px 0;
}

.service-header {
  background: #fff;
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

.service-title {
  font-size: 20px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 8px;
}

.service-desc {
  font-size: 14px;
  color: #969799;
  display: block;
}

.section-card {
  background: #fff;
  margin: 12px 16px;
  padding: 16px;
  border-radius: 12px;
}

.section-label {
  font-size: 16px;
  font-weight: 700;
  color: #323233;
  display: block;
  margin-bottom: 12px;
}

.section-text {
  font-size: 14px;
  color: #646566;
  line-height: 1.8;
  display: block;
}

.project-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.project-item {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 10px 16px;
}

.project-name {
  font-size: 13px;
  color: #323233;
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
  color: #646566;
}

.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: #fff;
  border-top: 1px solid #f2f3f5;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}
</style>
