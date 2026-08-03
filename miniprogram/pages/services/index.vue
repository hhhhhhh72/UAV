<template>
  <Layout :current="1">
    <view class="services-page">
      <u-sticky>
        <u-search v-model="searchText" placeholder="搜索业务分类" />
      </u-sticky>

      <view class="content-wrapper">
        <view v-if="filteredCategories.length === 0" class="empty-state">
          <u-icon name="search" size="48rpx" color="#c8c9cc" />
          <text class="empty-text">未找到匹配的业务分类</text>
        </view>

        <view v-for="cat in filteredCategories" :key="cat.id" class="category-card">
          <view class="category-header" @tap="goToCategory(cat)">
            <view class="category-icon-wrap" :style="{ background: cat.gradient }">
              <text class="category-icon-text">{{ cat.icon }}</text>
            </view>
            <view class="category-info">
              <text class="category-title">{{ cat.title }}</text>
              <text class="category-subtitle">{{ cat.subtitle }}</text>
            </view>
            <text class="arrow">›</text>
          </view>

          <view class="sub-service-list">
            <view v-for="sub in cat.subItems" :key="sub.id" class="sub-service-item" @tap="goToSubService(sub)">
              <view class="sub-icon" :style="{ background: cat.subColor }">
                <text class="sub-icon-text">{{ sub.icon }}</text>
              </view>
              <text class="sub-name">{{ sub.name }}</text>
              <text class="arrow">›</text>
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
  { id: 'supply-demand', title: '产业供需对接', subtitle: '资源匹配 高效对接', icon: '供', gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)', subColor: '#ede9fe', subItems: [
    { id: 'demand-hall', name: '需求大厅', icon: '求' },
    { id: 'bid-quote', name: '竞标报价', icon: '标' },
    { id: 'supply-show', name: '供应展示', icon: '展' },
  ]},
  { id: 'training', title: '培训认证', subtitle: '专业考证 技能提升', icon: '训', gradient: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)', subColor: '#cffafe', subItems: [
    { id: 'courses', name: '培训课程', icon: '课' },
    { id: 'certificates', name: '我的证书', icon: '证' },
    { id: 'competitions', name: '赛事中心', icon: '赛' },
  ]},
  { id: 'innovation', title: '产学研协同', subtitle: '科技成果 创新驱动', icon: '研', gradient: 'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)', subColor: '#ede9fe', subItems: [
    { id: 'achievements', name: '成果库', icon: '果' },
    { id: 'challenges', name: '研发难题', icon: '难' },
    { id: 'reports', name: '行业报告', icon: '报' },
  ]},
  { id: 'trade', title: '无人机交易', subtitle: '整机配件 一站购齐', icon: '易', gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)', subColor: '#fef3c7', subItems: [
    { id: 'products', name: '整机/配件', icon: '机' },
    { id: 'repair', name: '维修服务', icon: '修' },
  ]},
  { id: 'brand', title: '活动与品牌', subtitle: '协会活动 会员展示', icon: '品', gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)', subColor: '#d1fae5', subItems: [
    { id: 'events', name: '协会活动', icon: '活' },
    { id: 'portfolios', name: '品牌展示', icon: '牌' },
    { id: 'exhibitions', name: '展会排期', icon: '展' },
  ]},
  { id: 'insurance', title: '保险金融', subtitle: '全面保障 资金支持', icon: '保', gradient: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)', subColor: '#fee2e2', subItems: [
    { id: 'insurance-detail', name: '保险/年审/贷款', icon: '保' },
  ]},
  { id: 'emergency', title: '应急资源协同', subtitle: '快速响应 资源调度', icon: '急', gradient: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)', subColor: '#ffedd5', subItems: [
    { id: 'emergency-resources', name: '应急资源', icon: '资' },
    { id: 'rescue-cases', name: '救援案例', icon: '案' },
    { id: 'depts', name: '部门对接', icon: '部' },
  ]},
]

const filteredCategories = computed(() => {
  if (!searchText.value) return categories
  const q = searchText.value.toLowerCase()
  return categories.map(cat => {
    const matched = cat.subItems.filter(s => s.name.includes(q) || cat.title.includes(q))
    if (cat.title.includes(q)) return { ...cat }
    if (matched.length) return { ...cat, subItems: matched }
    return null
  }).filter(Boolean)
})

const navMap = {
  'demand-hall': '/pages/demands/list', 'bid-quote': '/pages/demands/bid',
  courses: '/pages/training/courses', certificates: '/pages/training/certificates',
  competitions: '/pages/events/list',
  achievements: '/pages/achievements/list', challenges: '/pages/challenges/list',
  reports: '/pages/reports/list',
  products: '/pages/services/detail?id=trade', repair: '/pages/services/detail?id=repair',
  events: '/pages/events/list', portfolios: '/pages/portfolios/list',
  exhibitions: '/pages/exhibitions/list',
  'emergency-resources': '/pages/emergency/resources',
  'rescue-cases': '/pages/emergency/cases', depts: '/pages/emergency/depts',
}

function goToCategory(cat) {
  if (cat.id === 'supply-demand') uni.navigateTo({ url: '/pages/demands/list' })
  else if (cat.id === 'training') uni.navigateTo({ url: '/pages/training/courses' })
  else if (cat.id === 'innovation') uni.navigateTo({ url: '/pages/achievements/list' })
  else if (cat.id === 'trade') uni.navigateTo({ url: '/pages/services/detail?id=trade' })
  else if (cat.id === 'brand') uni.navigateTo({ url: '/pages/events/list' })
  else if (cat.id === 'emergency') uni.navigateTo({ url: '/pages/emergency/resources' })
}

function goToSubService(sub) {
  uni.navigateTo({ url: navMap[sub.id] || '/pages/services/detail?id=' + sub.id })
}
</script>

<style scoped>
.services-page { background: var(--color-bg); }
.content-wrapper { padding: 12px; }
.empty-state { text-align: center; padding: 80px 0; color: var(--color-text-secondary); }
.empty-text { margin-top: 12px; font-size: 14px; }
.category-card { background: var(--color-bg-card); border-radius: 16px; padding: 16px; margin-bottom: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.02); overflow: hidden; }
.category-header { display: flex; align-items: center; gap: 12px; }
.category-icon-wrap { width: 44px; height: 44px; border-radius: 14px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.category-icon-text { font-size: 20px; font-weight: 600; color: #ffffff; }
.category-info { flex: 1; }
.category-title { font-size: 16px; font-weight: 600; color: var(--color-text); }
.category-subtitle { font-size: 12px; color: var(--color-text-secondary); margin-top: 2px; }
.arrow { font-size: 18px; color: var(--color-text-placeholder); }
.sub-service-list { margin-top: 12px; }
.sub-service-item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; background: var(--color-bg); border-radius: 10px; margin-top: 6px; }
.sub-service-item:active { opacity: 0.7; }
.sub-icon { width: 32px; height: 32px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.sub-icon-text { font-size: 15px; font-weight: 600; color: #ffffff; }
.sub-name { flex: 1; font-size: 14px; color: var(--color-text); }
</style>
