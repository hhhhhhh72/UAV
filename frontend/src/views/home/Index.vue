<template>
  <div class="home-page">
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <!-- 沉浸式背景 -->
      <div class="video-header">
        <img
          class="bg-video"
          src="https://www-cdn.djiits.com/cms/uploads/4d6128a30991074b6bad20e7e13a0c62.png"
          alt="background"
        />
        <div class="video-mask"></div>
      </div>

      <div class="video-placeholder"></div>

      <!-- 顶部：定位+搜索 -->
      <div class="header-section float-header">
        <div class="location-bar">
          <span class="location-text">重庆市 <van-icon name="arrow-down" size="12" /></span>
        </div>
        <div class="search-bar">
          <van-search placeholder="搜索需求/服务" shape="round" background="transparent" />
        </div>
      </div>

      <!-- 7 大业务系统入口 -->
      <div class="system-section">
        <div class="system-grid">
          <div
            v-for="sys in homeStore.systemEntries"
            :key="sys.id"
            class="system-card"
            :style="{ borderTopColor: sys.color }"
            @click="onSystemTap(sys.id)"
          >
            <van-icon :name="sys.icon" :color="sys.color" size="28" />
            <span class="system-name">{{ sys.name }}</span>
          </div>
        </div>
      </div>

      <!-- Banner 轮播 -->
      <div class="banner-section" v-if="homeStore.banners.length">
        <van-swipe class="my-swipe" :autoplay="5000" indicator-color="white">
          <van-swipe-item v-for="(item, index) in homeStore.banners" :key="index">
            <img :src="item.image" class="banner-image" alt="banner" />
          </van-swipe-item>
        </van-swipe>
      </div>

      <!-- 需求大厅 -->
      <div class="demand-section">
        <div class="section-header">
          <h3 class="section-title">需求大厅</h3>
          <span class="section-more" @click="router.push('/demands')">更多 <van-icon name="arrow" size="12" /></span>
        </div>

        <van-list
          v-model:loading="homeStore.loading"
          :finished="demandFinished"
          :finished-text="homeStore.demandFeed.length ? '没有更多了' : ''"
          @load="onLoadDemands"
        >
          <div v-if="!homeStore.demandFeed.length && !homeStore.loading" class="empty-state">
            <van-icon name="info-o" size="48" color="#c8c9cc" />
            <p>暂无需求信息</p>
          </div>

          <div
            v-for="d in homeStore.demandFeed"
            :key="d.id"
            class="demand-card"
            @click="goDemand(d.id)"
          >
            <div class="demand-top">
              <span class="demand-title">{{ d.title }}</span>
              <van-tag v-if="d.biz_type" plain type="primary" size="small">{{ d.biz_type }}</van-tag>
            </div>
            <p class="demand-desc">{{ d.description }}</p>
            <div class="demand-meta">
              <span v-if="d.budget_fen" class="demand-price">{{ formatPrice(d.budget_fen) }}</span>
              <span class="demand-time">{{ d.created_at || '' }}</span>
            </div>
          </div>
        </van-list>
      </div>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useHomeStore } from '@/stores/home'

const router = useRouter()
const homeStore = useHomeStore()

const refreshing = ref(false)
const demandFinished = ref(false)
const demandPage = ref(1)

onMounted(() => {
  homeStore.fetchHome()
})

async function onLoadDemands() {
  try {
    const data = await homeStore.fetchDemands({ page: demandPage.value, limit: 10 })
    demandPage.value++
    if (!data.items?.length && !data.data?.length) {
      demandFinished.value = true
    }
    if (data.total && data.total <= homeStore.demandFeed.length) {
      demandFinished.value = true
    }
  } catch {
    showToast('加载需求列表失败')
  }
}

async function onRefresh() {
  demandPage.value = 1
  demandFinished.value = false
  homeStore.demandFeed = []
  await homeStore.fetchHome()
  refreshing.value = false
  showToast('刷新成功')
}

function onSystemTap(id) {
  const routes = {
    1: '/members',
    2: '/demands',
    3: '/innovation',
    4: '/compliance',
    5: '/training',
    6: '/events',
    7: '/emergency',
  }
  router.push(routes[id] || '/services')
}

function goDemand(id) {
  router.push(`/demand-detail/${id}`)
}

function formatPrice(budgetFen) {
  if (budgetFen == null) return ''
  const yuan = (Number(budgetFen) / 100).toFixed(2)
  return `￥${yuan}`
}
</script>

<style scoped>
.home-page {
  background: #f5f6fa;
  min-height: 100vh;
  padding-bottom: var(--tabbar-height);
  width: 100%;
  max-width: var(--page-max-width);
  margin: 0 auto;
  --home-bg-height: 200px;
}

/* 沉浸式背景 */
.video-header {
  height: var(--home-bg-height);
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);
  z-index: 0;
  overflow: hidden;
  background: #0d2137;
}

.video-placeholder {
  height: var(--home-bg-height);
  width: 100%;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

.bg-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.video-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.4) 0%, rgba(0, 0, 0, 0) 60%, rgba(245, 246, 250, 1) 100%);
  pointer-events: none;
}

/* 悬浮顶部 */
.float-header {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);
  z-index: 100;
  padding: calc(12px + env(safe-area-inset-top)) 16px 12px;
  color: #fff;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.4) 0%, transparent 100%);
}

.location-bar {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
}

.location-text {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.search-bar :deep(.van-search) {
  padding: 0;
  background: transparent;
}

.search-bar :deep(.van-search__content) {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.search-bar :deep(.van-field__control) {
  color: #fff;
}

.search-bar :deep(.van-icon) {
  color: rgba(255, 255, 255, 0.8);
}

.search-bar :deep(input::placeholder) {
  color: rgba(255, 255, 255, 0.6);
}

/* 7 大业务系统入口 */
.system-section {
  margin: -20px 12px 12px;
  position: relative;
  z-index: 10;
}

.system-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 20px 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  border-top: 3px solid transparent;
}

.system-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  cursor: pointer;
  border-top: 3px solid transparent;
  border-radius: 12px;
  transition: background 0.2s;
}

.system-card:active {
  background: rgba(0, 0, 0, 0.04);
}

.system-name {
  font-size: 12px;
  color: #333;
  font-weight: 500;
  text-align: center;
}

/* Banner */
.banner-section {
  margin: 0 12px 16px;
  border-radius: 16px;
  overflow: hidden;
}

.banner-image {
  width: 100%;
  height: 150px;
  object-fit: cover;
  display: block;
}

/* 需求大厅 */
.demand-section {
  margin: 0 12px 20px;
  background: #fff;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0;
}

.section-more {
  font-size: 13px;
  color: #969799;
  display: flex;
  align-items: center;
  gap: 2px;
  cursor: pointer;
}

.demand-card {
  padding: 14px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
}

.demand-card:last-child {
  border-bottom: none;
}

.demand-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 6px;
}

.demand-title {
  font-size: 15px;
  font-weight: 500;
  color: #1a1a1a;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.demand-desc {
  font-size: 13px;
  color: #969799;
  margin: 0 0 8px;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.demand-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.demand-price {
  font-size: 15px;
  font-weight: 600;
  color: #ee0a24;
}

.demand-time {
  font-size: 12px;
  color: #c8c9cc;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
  color: #969799;
}

.empty-state p {
  margin: 8px 0 0;
  font-size: 14px;
}
</style>
