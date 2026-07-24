<template>
  <div class="services-page page-container">
    <van-sticky>
      <van-search v-model="searchText" placeholder="搜索服务" shape="round" />
    </van-sticky>

    <div class="content-wrapper">
      <!-- 遍历服务分组 -->
      <div 
        v-for="group in serviceGroups" 
        :key="group.title" 
        class="service-group-card"
      >
        <div class="group-header">
          <div class="group-title">{{ group.title }}</div>
          <div class="group-subtitle">{{ group.subtitle }}</div>
        </div>
        
        <div class="service-grid">
          <div
            v-for="service in group.items"
            :key="service.id"
            class="service-grid-item"
            @click="goToDetail(service.id)"
          >
            <div class="service-icon-large" :style="{ background: service.color }">
              <van-icon :name="service.icon" size="28" color="#ffffff" />
            </div>
            <h3 class="service-title">{{ service.name }}</h3>
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

// 基础服务数据池
const rawServices = [
  { id: 'flight', name: '飞行服务', description: '空域查询、飞行申报', icon: '/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 2, name: '政务服务', description: '环保监测、安全巡查', icon: 'eye-o', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 8, name: '无人机外卖', description: '即时配送、在线下单', icon: '/icons/delivery.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 1, name: '无人机物流', description: '城市配送、物资运输', icon: '/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 4, name: '无人机吊运', description: '高空吊运、设备安装', icon: '/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 5, name: '无人机表演', description: '活动表演、编队飞行', icon: '/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 3, name: '无人机托管', description: '专业托管、保养维护', icon: '/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
  { id: 7, name: '无人机租赁', description: '设备租赁、配件租赁', icon: 'coupon-o', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 6, name: '飞手培训', description: 'CAAC执照、技能培训', icon: '/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 9, name: '低空研学', description: '科普教育、实践体验', icon: '/icons/study.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 10, name: '无人机销售', description: '设备买卖、以旧换新', icon: '/icons/shop.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 11, name: '金融服务', description: '设备保险、飞行护航', icon: '/icons/finance.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 12, name: '维修服务', description: '故障维修、定期保养', icon: 'setting-o', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 13, name: '无人机赛事', description: '竞技比赛、赛事组织', icon: '/icons/competition.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' },
  { id: 14, name: '医疗配送', description: '无人机医疗物资配送', icon: 'shield-o', color: 'linear-gradient(135deg, #34d399 0%, #059669 100%)' },
  { id: 'reviews', name: '服务评价', description: '用户评价、服务反馈', icon: 'comment-o', color: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }
]

// 定义分组结构
const groupsConfig = [
  {
    title: '核心服务',
    subtitle: 'Core Services',
    ids: ['flight', 2, 8, 1, 14]
  },
  {
    title: '商业应用',
    subtitle: 'Business Application',
    ids: [4, 5, 3, 7, 13]
  },
  {
    title: '教育培训',
    subtitle: 'Education & Training',
    ids: [6, 9]
  },
  {
    title: '增值服务',
    subtitle: 'Value-added Services',
    ids: [10, 11, 12]
  },
  {
    title: '互动反馈',
    subtitle: 'Feedback',
    ids: ['reviews']
  }
]

// 计算分组后的数据
const serviceGroups = computed(() => {
  const all = rawServices
  // 如果有搜索词，只返回包含搜索词的单一列表（或过滤后的分组）
  if (searchText.value) {
    const filtered = all.filter(s => s.name.includes(searchText.value))
    return [{ title: '搜索结果', subtitle: 'Search Results', items: filtered }]
  }

  return groupsConfig.map(group => {
    return {
      ...group,
      items: group.ids.map(id => all.find(s => s.id === id)).filter(Boolean)
    }
  })
})

const goToDetail = (id) => {
  if (id === 'flight') {
    window.location.href = 'https://wx.zndkfx.com'
  } else if (id === 8) {
    window.location.href = 'https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130&title=%E6%97%A0%E4%BA%BA%E6%9C%BA%E5%A4%96%E5%8D%96%E9%85%8D%E9%80%81&jyauthcode='
  } else if (id === 9) {
    router.push('/study')
  } else if (id === 14) {
    router.push('/medical/order/create')
  } else if (id === 'reviews') {
    router.push('/reviews')
  } else {
    router.push(`/service-detail/${id}`)
  }
}
</script>

<style scoped>
.services-page {
  background: #f7f8fa; /* 统一浅灰背景 */
  min-height: 100vh;
}

/* 搜索栏优化 */
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
  border-radius: 12px; /* 搜索框圆角 */
}

.content-wrapper {
  padding: 12px;
}

/* 分组卡片 */
.service-group-card {
  background: #fff;
  border-radius: 16px;
  padding: 16px 12px; /* 减小内边距 */
  margin-bottom: 12px; /* 减小卡片间距 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02); /* 减淡投影 */
}

.group-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px; /* 减小标题与网格间距 */
  padding-left: 4px;
}

.group-title {
  font-size: 15px; /* 稍微减小标题 */
  font-weight: 600;
  color: #1a1a1a;
}

.group-subtitle {
  display: none; /* 隐藏英文副标题 */
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px 4px; /* 紧凑行间距 */
}

.service-grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  cursor: pointer;
  transition: opacity 0.2s;
}

.service-grid-item:active {
  opacity: 0.6;
}

.service-icon-large {
  width: 44px; /* iOS 标准紧凑尺寸 */
  height: 44px;
  border-radius: 14px; /* 适配小尺寸的 Squircle */
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-icon-large :deep(.van-icon) {
  font-size: 24px !important; /* 强制图标大小 */
}

/* 强制 SVG 图片图标变白 */
.service-icon-large :deep(.van-icon__image) {
  filter: brightness(0) invert(1);
}

.service-title {
  font-size: 12px; /* 精致小字体 */
  font-weight: 400;
  color: #333;
  margin: 0;
  line-height: 1.3;
  white-space: nowrap; /* 防止换行 */
  transform: scale(0.95); /* 视觉上更小一点 */
}
</style>

