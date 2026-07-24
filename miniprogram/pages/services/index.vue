<template>
  <Layout :current="1">
    <view class="services-page">
    <van-sticky>
      <van-search v-model="searchText" placeholder="搜索业务分类" shape="round" />
    </van-sticky>

    <view class="content-wrapper">
      <view v-if="filteredCategories.length === 0" class="empty-state">
        <van-icon name="search" size="48" color="#c8c9cc" />
        <text class="empty-text">未找到匹配的业务分类</text>
      </view>

      <view
        v-for="cat in filteredCategories"
        :key="cat.id"
        class="category-card"
      >
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
  </Layout>
</template>

<script>
import Layout from '@/components/Layout.vue'
export default {
  components: { Layout },
  data() {
    return {
      searchText: '',
      categories: [
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
          id: 'innovation',
          title: '产学研协同',
          subtitle: '科技成果 创新驱动',
          icon: 'certificate',
          gradient: 'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)',
          subColor: '#ede9fe',
          subItems: [
            { id: 'achievements', name: '成果库', icon: 'star-o' },
            { id: 'challenges', name: '研发难题', icon: 'fire-o' },
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
          id: 'brand',
          title: '活动与品牌',
          subtitle: '协会活动 会员展示',
          icon: 'star-o',
          gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
          subColor: '#d1fae5',
          subItems: [
            { id: 'events', name: '协会活动', icon: 'calendar-o' },
            { id: 'portfolios', name: '品牌展示', icon: 'shop-o' },
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
    }
  },
  computed: {
    filteredCategories() {
      if (!this.searchText) return this.categories
      const q = this.searchText.toLowerCase()
      return this.categories
        .map(cat => {
          const matchedSubs = cat.subItems.filter(sub => sub.name.includes(q) || cat.title.includes(q))
          if (cat.title.includes(q)) return { ...cat }
          if (matchedSubs.length > 0) return { ...cat, subItems: matchedSubs }
          return null
        })
        .filter(Boolean)
    }
  },
  methods: {
    goToCategory(cat) {
      if (cat.id === 'supply-demand') {
        uni.navigateTo({ url: '/pages/demands/list' })
      } else if (cat.id === 'training') {
        uni.navigateTo({ url: '/pages/training/courses' })
      } else if (cat.id === 'innovation') {
        uni.navigateTo({ url: '/pages/achievements/list' })
      } else if (cat.id === 'trade') {
        uni.navigateTo({ url: '/pages/services/detail?id=trade' })
      } else if (cat.id === 'brand') {
        uni.navigateTo({ url: '/pages/events/list' })
      } else if (cat.id === 'insurance') {
        uni.navigateTo({ url: '/pages/services/detail?id=insurance' })
      } else if (cat.id === 'emergency') {
        uni.navigateTo({ url: '/pages/emergency/resources' })
      }
    },
    goToSubService(cat, sub) {
      const nav = {
        'demand-hall': '/pages/demands/list',
        'bid-quote': '/pages/demands/bid',
        'caac': '/pages/study/index', 'utc': '/pages/study/index', 'hr-cert': '/pages/study/index',
        'pilot': '/pages/training/courses',
        'achievements': '/pages/achievements/list',
        'challenges': '/pages/challenges/list',
        'events': '/pages/events/list',
        'portfolios': '/pages/portfolios/list',
        'rescue-case': '/pages/cases/index',
        'resource-dispatch': '/pages/emergency/resources',
        'policy': '/pages/services/detail?id=insurance',
        'annual': '/pages/services/detail?id=insurance',
        'loan': '/pages/services/detail?id=insurance',
        'drone-unit': '/pages/services/detail?id=trade',
        'repair': '/pages/services/detail?id=repair',
        'parts': '/pages/services/detail?id=trade',
      }
      if (nav[sub.id]) {
        uni.navigateTo({ url: nav[sub.id] })
      } else {
        uni.navigateTo({ url: '/pages/services/detail?id=' + sub.id })
      }
    }
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
  color: #969799;
}

.empty-text {
  margin-top: 12px;
  font-size: 14px;
}

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

.category-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.category-info {
  flex: 1;
}

.category-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
}

.category-subtitle {
  font-size: 12px;
  color: #969799;
  margin-top: 2px;
}

.sub-service-list {
  margin-top: 12px;
}

.sub-service-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: #f8f9fa;
  border-radius: 10px;
  margin-top: 6px;
}

.sub-service-item:active {
  opacity: 0.7;
}

.sub-icon {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sub-name {
  flex: 1;
  font-size: 14px;
  color: #333;
}
</style>
