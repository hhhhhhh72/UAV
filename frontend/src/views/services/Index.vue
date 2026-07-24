<template>
  <div class="services-page page-container">
    <van-sticky>
      <van-search v-model="searchText" placeholder="搜索业务分类" shape="round" />
    </van-sticky>

    <div class="content-wrapper">
      <div v-if="filteredCategories.length === 0" class="empty-state">
        <van-icon name="search" size="48" color="#c8c9cc" />
        <p>未找到匹配的业务分类</p>
      </div>

      <div
        v-for="cat in filteredCategories"
        :key="cat.id"
        class="category-card"
      >
        <div class="category-header" @click="goToCategory(cat)">
          <div class="category-icon-wrap" :style="{ background: cat.gradient }">
            <van-icon :name="cat.icon" size="24" color="#ffffff" />
          </div>
          <div class="category-info">
            <div class="category-title">{{ cat.title }}</div>
            <div class="category-subtitle">{{ cat.subtitle }}</div>
          </div>
          <van-icon name="arrow" size="14" color="#c8c9cc" />
        </div>

        <div class="sub-service-list">
          <div
            v-for="sub in cat.subItems"
            :key="sub.id"
            class="sub-service-item"
            @click="goToSubService(cat, sub)"
          >
            <div class="sub-icon" :style="{ background: cat.subColor || cat.accent }">
              <van-icon :name="sub.icon" size="16" color="#ffffff" />
            </div>
            <span class="sub-name">{{ sub.name }}</span>
            <van-icon name="arrow" size="12" color="#c8c9cc" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const searchText = ref('')

const categories = [
  {
    id: 'supply-demand',
    title: '产业供需对接',
    subtitle: '资源匹配 高效对接',
    icon: 'exchange',
    gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    accent: '#7c3aed',
    subColor: '#ede9fe',
    path: '/demands',
    subItems: [
      { id: 'demand-hall', name: '需求大厅', icon: 'bullhorn-o', path: '/demands' },
      { id: 'supply-show', name: '供应展示', icon: 'shop-o', path: '/supplies' },
      { id: 'bid-quote', name: '竞标报价', icon: 'records-o', path: '/bids' },
    ]
  },
  {
    id: 'training',
    title: '培训认证',
    subtitle: '专业考证 技能提升',
    icon: 'medal-o',
    gradient: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)',
    accent: '#0e7490',
    subColor: '#cffafe',
    path: '/training',
    subItems: [
      { id: 'caac', name: 'CAAC执照', icon: 'certificate', path: '/training/caac' },
      { id: 'utc', name: 'UTC认证', icon: 'medal-o', path: '/training/utc' },
      { id: 'hr-cert', name: '人社认证', icon: 'records-o', path: '/training/hr' },
      { id: 'pilot', name: '飞手培训', icon: 'user-o', path: '/service-detail/6' },
    ]
  },
  {
    id: 'trade',
    title: '无人机交易',
    subtitle: '整机配件 一站购齐',
    icon: 'shopping-cart-o',
    gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
    accent: '#b45309',
    subColor: '#fef3c7',
    path: '/trade',
    subItems: [
      { id: 'drone-unit', name: '整机购买', icon: 'gem-o', path: '/trade/unit' },
      { id: 'repair', name: '维修服务', icon: 'setting-o', path: '/service-detail/12' },
      { id: 'parts', name: '配件商城', icon: 'more-o', path: '/trade/parts' },
    ]
  },
  {
    id: 'contract',
    title: '合同签约',
    subtitle: '电子签章 安全合规',
    icon: 'edit',
    gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    accent: '#047857',
    subColor: '#d1fae5',
    path: '/contract',
    subItems: [
      { id: 'template', name: '合同模板', icon: 'description', path: '/contract/template' },
      { id: 'signature', name: '在线签章', icon: 'sign', path: '/contract/sign' },
      { id: 'void', name: '合同作废', icon: 'delete-o', path: '/contract/void' },
    ]
  },
  {
    id: 'insurance',
    title: '保险金融',
    subtitle: '全面保障 资金支持',
    icon: 'shield-o',
    gradient: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
    accent: '#b91c1c',
    subColor: '#fee2e2',
    path: '/insurance',
    subItems: [
      { id: 'policy', name: '无人机保单', icon: 'shield-o', path: '/insurance/policy' },
      { id: 'annual', name: '年审服务', icon: 'clock-o', path: '/insurance/annual' },
      { id: 'loan', name: '金融贷款', icon: 'gold-coin-o', path: '/insurance/loan' },
    ]
  },
  {
    id: 'emergency',
    title: '应急资源协同',
    subtitle: '快速响应 资源调度',
    icon: 'fire-o',
    gradient: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)',
    accent: '#c2410c',
    subColor: '#ffedd5',
    path: '/emergency',
    subItems: [
      { id: 'rescue-case', name: '救援案例', icon: 'info-o', path: '/cases' },
      { id: 'resource-dispatch', name: '资源调度', icon: 'send-gift-o', path: '/emergency/dispatch' },
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

function goToCategory(cat) {
  if (cat.id === 'supply-demand') {
    router.push('/demands')
  } else {
    router.push(cat.path)
  }
}

function goToSubService(cat, sub) {
  // 供应展示跳转到需求大厅（供需一体）
  if (sub.id === 'supply-show') {
    router.push('/demands')
    return
  }
  router.push(sub.path)
}
</script>

<style scoped>
.services-page {
  background: #f7f8fa;
  min-height: 100vh;
}

.services-page :deep(.van-search) {
  background: #fff;
  border-bottom-left-radius: 24px;
  border-bottom-right-radius: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
  padding: 12px 16px;
  margin-bottom: 4px;
}

.services-page :deep(.van-search__content) {
  background: #f7f8fa;
  border-radius: 12px;
}

.content-wrapper {
  padding: 12px;
}

.empty-state {
  text-align: center;
  padding: 80px 0;
  color: #969799;
}

.empty-state p {
  margin-top: 12px;
  font-size: 14px;
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
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
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
}

.category-subtitle {
  font-size: 12px;
  color: #969799;
}

/* 子服务列表 */
.sub-service-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
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
  cursor: pointer;
  transition: background 0.2s;
  -webkit-tap-highlight-color: transparent;
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

.sub-service-item .van-icon-arrow {
  flex-shrink: 0;
}
</style>
