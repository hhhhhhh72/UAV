<template>
  <Layout :current="3">
    <view class="services-page">
      <view class="page-head" :style="{ paddingTop: statusBarHeight + 'px' }">
        <text class="page-title">生态服务</text>
        <text class="page-sub">六大板块 · 产业服务一站式入口</text>
      </view>

      <view v-for="cat in categories" :key="cat.id" class="cat-section">
        <view class="cat-head">
          <view class="cat-dot" :style="{ background: cat.color }"></view>
          <text class="cat-title">{{ cat.title }}</text>
          <text class="cat-sub">{{ cat.subtitle }}</text>
        </view>
        <view class="cat-grid">
          <view
            v-for="s in cat.items"
            :key="s.name"
            class="cat-item"
            hover-class="tap-fade"
            @tap="go(s.path)"
          >
            <view class="cat-icon" :style="{ background: cat.bg }">
              <image class="cat-icon-img" :src="s.icon" mode="aspectFit" />
            </view>
            <text class="cat-name">{{ s.name }}</text>
          </view>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import Layout from '@/components/Layout.vue'

// 状态栏高度：custom 导航下自绘标题区需 JS 接管
const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 20

const iconRoot = '/static/home/icons/'
const go = (path) => {
  if (path === 'tab-services') return uni.switchTab({ url: '/pages/services/index' })
  if (path) uni.navigateTo({ url: path })
}

const categories = [
  {
    id: 'policy',
    title: '政策合规',
    subtitle: '资讯 · 标准 · 申报',
    color: '#0A66C2',
    bg: '#EAF3FB',
    items: [
      { name: '政策资讯', icon: iconRoot + 'policy.svg', path: '/pkg-service/pages/compliance/news' },
      { name: '合规知识库', icon: iconRoot + 'policy.svg', path: '/pkg-service/pages/compliance/knowledge' },
      { name: '团体标准', icon: iconRoot + 'policy.svg', path: '/pkg-service/pages/compliance/standards' },
      { name: '项目申报', icon: iconRoot + 'edit.svg', path: '/pkg-app/pages/applications/index' },
      { name: '企业案例', icon: iconRoot + 'shop.svg', path: '/pkg-eco/pages/cases/index' },
    ],
  },
  {
    id: 'innovation',
    title: '产学研用',
    subtitle: '成果 · 难题 · 课题',
    color: '#6941C6',
    bg: '#F6F4FF',
    items: [
      { name: '科技成果库', icon: iconRoot + 'demand.svg', path: '/pkg-eco/pages/achievements/list' },
      { name: '研发难题', icon: iconRoot + 'edit.svg', path: '/pkg-eco/pages/challenges/list' },
      { name: '课题攻关', icon: iconRoot + 'demand.svg', path: '/pkg-eco/pages/projects/list' },
      { name: '场地预约', icon: iconRoot + 'trade.svg', path: '/pkg-service/pages/testsites/list' },
    ],
  },
  {
    id: 'talent',
    title: '人才培育',
    subtitle: '培训 · 考证 · 招聘',
    color: '#E96012',
    bg: '#FFF0E6',
    items: [
      { name: '培训课程', icon: iconRoot + 'training.svg', path: '/pkg-talent/pages/training/courses' },
      { name: '考证管理', icon: iconRoot + 'training.svg', path: '/pkg-talent/pages/training/certificates' },
      { name: '招聘求职', icon: iconRoot + 'pilot.svg', path: '/pkg-talent/pages/jobs/list' },
      { name: '院校展示', icon: iconRoot + 'training.svg', path: '/pkg-eco/pages/colleges/list' },
    ],
  },
  {
    id: 'brand',
    title: '活动品牌',
    subtitle: '活动 · 展会 · 报告',
    color: '#168A55',
    bg: '#E9F7F0',
    items: [
      { name: '协会活动', icon: iconRoot + 'message-blue.svg', path: '/pkg-eco/pages/activities/list' },
      { name: '赛事活动', icon: iconRoot + 'pilot.svg', path: '/pkg-eco/pages/competitions/list' },
      { name: '品牌展示', icon: iconRoot + 'shop.svg', path: '/pkg-eco/pages/portfolios/list' },
      { name: '展会排期', icon: iconRoot + 'ecoservice.svg', path: '/pkg-eco/pages/exhibitions/list' },
      { name: '行业报告', icon: iconRoot + 'policy.svg', path: '/pkg-service/pages/reports/list' },
    ],
  },
  {
    id: 'emergency',
    title: '应急调度',
    subtitle: '资源 · 调度 · 演练',
    color: '#D92D20',
    bg: '#FEF3F2',
    items: [
      { name: '应急资源', icon: iconRoot + 'trade.svg', path: '/pkg-emergency/pages/emergency/resources' },
      { name: '调度记录', icon: iconRoot + 'trade.svg', path: '/pkg-emergency/pages/emergency/dispatches' },
      { name: '救援案例', icon: iconRoot + 'ecoservice.svg', path: '/pkg-emergency/pages/emergency/cases' },
      { name: '部门对接', icon: iconRoot + 'policy.svg', path: '/pkg-emergency/pages/emergency/depts' },
      { name: '联合演练', icon: iconRoot + 'pilot.svg', path: '/pkg-emergency/pages/emergency/dispatches' },
    ],
  },
  {
    id: 'resources',
    title: '资源总库',
    subtitle: '智库 · 台账 · 人才',
    color: '#0A66C2',
    bg: '#F4F8FC',
    items: [
      { name: '专家智库', icon: iconRoot + 'pilot.svg', path: '/pkg-talent/pages/experts/list' },
      { name: '产业资源台账', icon: iconRoot + 'demand.svg', path: '/pkg-service/pages/resources/list' },
      { name: '认证飞手', icon: iconRoot + 'pilot.svg', path: '/pkg-talent/pages/pilots/list' },
      { name: '低空研学', icon: iconRoot + 'training.svg', path: '/pkg-talent/pages/study/index' },
      { name: '入驻企业', icon: iconRoot + 'ecoservice.svg', path: '/pkg-eco/pages/enterprise/list' },
    ],
  },
]
</script>

<style scoped>
.services-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding: 16px 12px calc(24px + env(safe-area-inset-bottom));
}

.page-head {
  padding: 8px 4px 16px;
}

.page-title {
  display: block;
  font-size: 21px;
  font-weight: 700;
  color: #17212B;
}

.page-sub {
  display: block;
  font-size: 12px;
  color: #667085;
  margin-top: 4px;
}

/* 分类 */
.cat-section {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #EEF1F4;
  padding: 14px;
  margin-bottom: 10px;
}

.cat-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.cat-dot {
  width: 8px;
  height: 8px;
  border-radius: 4px;
}

.cat-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
}

.cat-sub {
  font-size: 11px;
  color: #98A2B3;
}

.cat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.cat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.cat-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cat-icon-img {
  width: 26px;
  height: 26px;
}

.cat-name {
  font-size: 11px;
  color: #344054;
  text-align: center;
}

.tap-fade {
  opacity: 0.7;
}
</style>
