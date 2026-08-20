<template>
  <Layout :current="2">
    <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
      <!-- 标题区 -->
      <view class="pub-home-head">
        <view class="pub-eyebrow">低空综合服务平台</view>
        <view class="pub-h1">发布</view>
        <view class="pub-sub">发布需求与供给，快速对接业务机会</view>
      </view>

      <!-- 草稿横幅 -->
      <view v-if="draftCount > 0" class="pub-draft-banner">
        <view class="pub-draft-dot"></view>
        <text>你有 {{ draftCount }} 条草稿待完善</text>
        <view class="pub-draft-btn" @tap="resumeDraft">继续编辑</view>
      </view>

      <!-- 四张发布入口卡片 -->
      <view class="pub-grid">
        <view
          v-for="card in typeCards"
          :key="card.key"
          class="pub-type-card"
          hover-class="pub-type-card--active"
          @tap="chooseType(card.key)"
        >
          <view class="pub-type-icon" :class="'pub-type-icon--' + card.key">
            <image :src="card.icon" mode="aspectFit" />
          </view>
          <view class="pub-type-main">
            <text class="pub-type-name">{{ card.name }}</text>
            <text class="pub-type-desc">{{ card.desc }}</text>
          </view>
          <text class="pub-arrow">›</text>
        </view>
      </view>

      <!-- 我的发布 -->
      <view class="pub-section-title">我的发布</view>
      <view class="pub-manage">
        <view class="pub-manage-row" hover-class="pub-manage-row--active" @tap="goMyPosts('all', 'all')">
          <text class="pub-manage-name">全部发布</text>
          <text class="pub-manage-desc pub-manage-count">{{ totalCount }} 条</text>
          <text class="pub-arrow" style="font-size:14px">›</text>
        </view>
        <view class="pub-manage-row" hover-class="pub-manage-row--active" @tap="goMyPosts('demand', 'all')">
          <text class="pub-manage-name">我的需求</text>
          <text class="pub-manage-desc">状态 · 对接意向</text>
          <text class="pub-arrow" style="font-size:14px">›</text>
        </view>
        <view class="pub-manage-row" hover-class="pub-manage-row--active" @tap="goMyPosts('service', 'all')">
          <text class="pub-manage-name">我的服务</text>
          <text class="pub-manage-desc">能力卡 · 线索咨询</text>
          <text class="pub-arrow" style="font-size:14px">›</text>
        </view>
        <view class="pub-manage-row" hover-class="pub-manage-row--active" @tap="goMyPosts('all', 'draft')">
          <text class="pub-manage-name">我的草稿</text>
          <text class="pub-manage-desc">{{ draftCount }} 条待完善</text>
          <text class="pub-arrow" style="font-size:14px">›</text>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import Layout from '../../components/Layout.vue'
import { TYPES, getPosts, draftPosts } from '../../utils/publishData'
import { request } from '../../utils/request'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop()

const typeCards = [
  { key: 'demand', name: '发布需求', desc: '发布具体项目，获得飞手与服务商报价', icon: '/static/publish/demand.svg' },
  { key: 'service', name: '发布服务能力', desc: '展示团队、设备和可承接服务，让需求方主动联系', icon: '/static/publish/service.svg' },
  { key: 'product', name: '发布商品设备', desc: '按商品逻辑补齐型号、成色、价格和交付方式', icon: '/static/publish/product.svg' },
  { key: 'course', name: '发布培训课程', desc: '用课程、证书、日期与招生信息回答学员的核心问题', icon: '/static/publish/course.svg' },
]

const allPosts = ref([])
const backendCount = ref(0)
const draftCount = computed(() => draftPosts().length)
const totalCount = computed(() => allPosts.value.length + backendCount.value)

// 我的发布计数 = 本地记录 + 后端已提交（需求/商品/服务/课程 mine=1；未登录后端返回空）
async function refresh() {
  allPosts.value = getPosts()
  try {
    const [dRes, pRes, sRes, cRes] = await Promise.all([
      request({ url: '/api/v1/demands', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/products', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/service-listings', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/training-courses', data: { mine: '1', page: 1, page_size: 100 } }),
    ])
    const dList = Array.isArray(dRes) ? dRes : dRes?.data || []
    const pList = Array.isArray(pRes) ? pRes : pRes?.data || []
    const sList = Array.isArray(sRes) ? sRes : sRes?.data || []
    const cList = Array.isArray(cRes) ? cRes : cRes?.data || []
    backendCount.value = dList.length + pList.length + sList.length + cList.length
  } catch (e) {
    backendCount.value = 0
  }
}

function chooseType(type) {
  uni.navigateTo({ url: '/pages/publish/form?type=' + type })
}

function resumeDraft() {
  uni.navigateTo({ url: '/pages/publish/form?type=demand&resume=1' })
}

function goMyPosts(kind, tab) {
  uni.navigateTo({ url: '/pages/publish/my-posts?kind=' + kind + '&tab=' + tab })
}

onShow(() => {
  initSafeTop()
  refresh()
})
</script>

<style scoped>
@import './pub-style.css';
.pub-page {
  min-height: 100vh;
  background: #F5F6F8;
  padding: 14px 12px calc(96px + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.pub-manage-desc {
  flex: 1;
  text-align: right;
}
.pub-manage-count {
  margin-left: 0;
}
.pub-manage-row .pub-arrow { font-size: 14px; }
</style>
