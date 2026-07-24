<template>
  <Layout :current="1">
    <view class="services-page">
      <!-- 搜索栏 -->
      <van-sticky>
        <van-search
          v-model="searchText"
          placeholder="搜索业务分类"
          shape="round"
        />
      </van-sticky>

      <view class="content-wrapper">
        <!-- 空状态 -->
        <view v-if="filteredCategories.length === 0" class="empty-state">
          <text class="empty-text">未找到匹配的业务分类</text>
        </view>

        <!-- 分类卡片列表 -->
        <view
          v-for="cat in filteredCategories"
          :key="cat.id"
          class="category-card"
        >
          <!-- 分类头部 -->
          <view class="category-header" @tap="goToCategory(cat)">
            <view class="category-icon-wrap" :style="{ background: cat.gradient }">
              <van-icon :name="cat.icon" size="24" color="#ffffff" />
            </view>
            <view class="category-info">
              <text class="category-title">{{ cat.title }}</text>
              <text class="category-subtitle">{{ cat.subtitle }}</text>
            </view>
            <van-icon name="arrow" size="14" color="#c8c9cc" />
          </view>

          <!-- 子服务列表 -->
          <view class="sub-service-list">
            <view
              v-for="sub in cat.subItems"
              :key="sub.id"
              class="sub-service-item"
              @tap="goToSubService(cat, sub)"
            >
              <view class="sub-icon" :style="{ background: cat.subColor }">
                <van-icon :name="sub.icon" size="16" color="#ffffff" />
              </view>
              <text class="sub-name">{{ sub.name }}</text>
              <van-icon name="arrow" size="12" color="#c8c9cc" />
            </view>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import Layout from '@/components/Layout.vue'

const searchText = ref('')

const categories = [
  {
    id: 'supply-demand',
    title: '产业供需对接',
    subtitle: '资源匹配 高效对接',
    icon: 'exchange',
    gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    subColor: '#ede9fe',
    subItems: [
      { id: 'demand-hall', name: '需求大厅', icon: 'bullhorn-o' },
      { id: 'supply-show', name: '供应展示', icon: 'shop-o' },
      { id: 'bid-quote', name: '竞标报价', icon: 'records-o' },
    ]
  },
  {
    id: 'training',
    title: '培训认证',
    subtitle: '专业考证 技能提升',
    icon: 'medal-o',
    gradient: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)',
    subColor: '#cffafe',
    subItems: [
      { id: 'caac', name: 'CAAC执照', icon: 'certificate' },
      { id: 'utc', name: 'UTC认证', icon: 'medal-o' },
      { id: 'hr-cert', name: '人社认证', icon: 'records-o' },
      { id: 'pilot', name: '飞手培训', icon: 'user-o' },
    ]
  },
  {
    id: 'trade',
    title: '无人机交易',
    subtitle: '整机配件 一站购齐',
    icon: 'shopping-cart-o',
    gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
    subColor: '#fef3c7',
    subItems: [
      { id: 'drone-unit', name: '整机购买', icon: 'gem-o' },
      { id: 'repair', name: '维修服务', icon: 'setting-o' },
      { id: 'parts', name: '配件商城', icon: 'more-o' },
    ]
  },
  {
    id: 'contract',
    title: '合同签约',
    subtitle: '电子签章 安全合规',
    icon: 'edit',
    gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    subColor: '#d1fae5',
    subItems: [
      { id: 'template', name: '合同模板', icon: 'description' },
      { id: 'signature', name: '在线签章', icon: 'sign' },
      { id: 'void', name: '合同作废', icon: 'delete-o' },
    ]
  },
  {
    id: 'insurance',
    title: '保险金融',
    subtitle: '全面保障 资金支持',
    icon: 'shield-o',
    gradient: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
    subColor: '#fee2e2',
    subItems: [
      { id: 'policy', name: '无人机保单', icon: 'shield-o' },
      { id: 'annual', name: '年审服务', icon: 'clock-o' },
      { id: 'loan', name: '金融贷款', icon: 'gold-coin-o' },
    ]
  },
  {
    id: 'emergency',
    title: '应急资源协同',
    subtitle: '快速响应 资源调度',
    icon: 'fire-o',
    gradient: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)',
    subColor: '#ffedd5',
    subItems: [
      { id: 'rescue-case', name: '救援案例', icon: 'info-o' },
      { id: 'resource-dispatch', name: '资源调度', icon: 'send-gift-o' },
    ]
  },
]

const filteredCategories = computed(() => {
  if (!searchText.value) return categories
  const q = searchText.value.toLowerCase()
  return categories
    .map(cat => {
      const matchedSubs = cat.subItems.filter(
        sub => sub.name.includes(q) || cat.title.includes(q)
      )
      if (cat.title.includes(q)) return { ...cat, subItems: cat.subItems }
      if (matchedSubs.length > 0) return { ...cat, subItems: matchedSubs }
      return null
    })
    .filter(Boolean)
})

const goToCategory = (cat) => {
  uni.showToast({ title: cat.title + ' - 即将上线', icon: 'none', duration: 1500 })
}

const goToSubService = (cat, sub) => {
  if (sub.id === 'demand-hall') {
    uni.navigateTo({ url: '/pages/demands/list' })
  } else if (sub.id === 'bid-quote') {
    uni.navigateTo({ url: '/pages/demands/bid' })
  } else if (sub.id === 'pilot') {
    uni.navigateTo({ url: '/pages/services/detail?id=6' })
  } else if (sub.id === 'repair') {
    uni.navigateTo({ url: '/pages/services/detail?id=12' })
  } else if (sub.id === 'rescue-case') {
    uni.navigateTo({ url: '/pages/cases/index' })
  } else if (sub.id === 'caac' || sub.id === 'utc' || sub.id === 'hr-cert') {
    uni.navigateTo({ url: '/pages/study/index' })
  } else {
    uni.showToast({ title: sub.name + ' - 即将上线', icon: 'none', duration: 1500 })
  }
}
</script>

<style scoped>
.services-page {
  background: #f7f8fa;
  min-height: 100vh;
}

.content-wrapper {
  padding: 12px;
}

.empty-state {
  text-align: center;
  padding: 80px 0;
}

.empty-text {
  font-size: 14px;
  color: #969799;
}

/* 分类卡片 */
.category-card {
  background: #fff;
  border-radius: 16px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
  overflow: hidden;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-header:active {
  opacity: 0.7;
}

.category-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
}

.category-info {
  flex: 1;
  min-width: 0;
}

.category-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 2px;
  display: block;
}

.category-subtitle {
  font-size: 12px;
  color: #969799;
  display: block;
}

/* 子服务列表 */
.sub-service-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f5f5f5;
}

.sub-service-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f7f8fa;
  border-radius: 10px;
  width: calc(50% - 4px);
  box-sizing: border-box;
}

.sub-service-item:active {
  background: #f0f0f0;
}

.sub-icon {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sub-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: #323233;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
