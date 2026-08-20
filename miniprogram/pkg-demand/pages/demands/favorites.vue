<template>
  <view class="fav-page">
    <!-- 头部 -->
    <view class="page-header" :style="headerStyle">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">我的收藏</text>
      <view class="head-action" :style="{ marginRight: capsuleGap + 'px' }" @tap="goHall">
        <text class="head-action-text">需求大厅</text>
      </view>
    </view>

    <!-- 列表标题 -->
    <view class="list-head">
      <text class="list-title">收藏记录</text>
      <text class="list-count">共 {{ list.length }} 条</text>
    </view>

    <!-- 加载中骨架 -->
    <view v-if="loading" class="post-list">
      <view v-for="i in 3" :key="i" class="mine-card">
        <view class="skl" style="width: 96rpx; height: 36rpx"></view>
        <view class="skl post-title-skl"></view>
        <view class="skl" style="width: 320rpx; height: 26rpx"></view>
      </view>
    </view>

    <!-- 空状态 -->
    <view v-else-if="list.length === 0" class="state-panel">
      <view class="state-mark">♡</view>
      <text class="state-title">{{ loadError ? '加载失败' : '还没有收藏' }}</text>
      <text class="state-desc">{{ loadError ? '网络异常，请稍后重试' : '在需求大厅看到感兴趣的需求，点「收藏」就能在这里找到' }}</text>
      <view class="state-btn" @tap="loadError ? fetchFavorites() : goHall">{{ loadError ? '重新加载' : '去逛逛' }}</view>
    </view>

    <!-- 收藏列表 -->
    <view v-else class="post-list">
      <view v-for="d in list" :key="d.id" class="mine-card" hover-class="mine-card--active" @tap="goDetail(d)">
        <view class="tag-row">
          <text class="tag blue">需求</text>
          <text class="tag" :class="statusTagClass(d.status)">{{ statusText(d.status) }}</text>
        </view>
        <text class="post-title">{{ d.title || '未命名需求' }}</text>
        <view class="post-meta">
          <text class="post-meta-item">{{ bizTypeLabel(d.biz_type) || '其他' }}</text>
          <text class="post-meta-item">{{ d.district || '重庆' }}</text>
          <text class="post-meta-item">{{ d.budget_fen ? '预算 ' + fmtMoney(d.budget_fen) + ' 元' : '预算可协商' }}</text>
          <text class="post-meta-item" v-if="d.publisher_name">发布者 · {{ d.publisher_name }}</text>
        </view>
        <view class="mine-action-row">
          <view class="action-link" @tap.stop="goDetail(d)">查看详情</view>
          <view class="action-link danger" @tap.stop="unfavorite(d)">取消收藏</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { safeNavigateTo } from '../../../utils/nav'
import { request, getErrorMessage, authStorage } from '../../../utils/request'
import { bizTypeLabel } from '../../../utils/enums'
import { useSafeTop } from '../../../utils/safeTop'

const list = ref([])
const loading = ref(false)
const loadError = ref(false)

// 自定义导航：状态栏留白 + 右上角避让微信胶囊
const statusBarH = ref(20)
try { statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20 } catch (e) { /* 默认 20 */ }
const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)
const headerStyle = computed(() => ({
  paddingTop: (topPad.value || statusBarH.value) + 'px',
  height: (56 + (topPad.value || statusBarH.value)) + 'px',
}))

const DEMAND_STATUS = { pending: '待审核', published: '已上架', completed: '已结束', cancelled: '已下架', rejected: '未通过' }
const statusText = (s) => DEMAND_STATUS[s] || s || ''
const statusTagClass = (s) => {
  if (s === 'rejected' || s === 'cancelled') return 'red'
  if (s === 'pending') return 'orange'
  if (s === 'completed') return 'gray'
  return 'green'
}

const fmtMoney = (fen) => {
  const yuan = (Number(fen) || 0) / 100
  return String(Math.round(yuan)).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}
const normalizeList = (res) => {
  if (Array.isArray(res)) return res
  if (res && Array.isArray(res.data)) return res.data
  return []
}

const fetchFavorites = async () => {
  loading.value = true
  loadError.value = false
  try {
    const res = await request({ url: '/api/v1/demands/favorites/mine' })
    list.value = normalizeList(res)
  } catch (e) {
    loadError.value = true
    list.value = []
  } finally {
    loading.value = false
  }
}

onLoad(() => {
  initSafeTop()
  // 未登录不进收藏页：引导登录
  if (!authStorage.getAccessToken()) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  fetchFavorites()
})
onPullDownRefresh(() => {
  fetchFavorites().finally(() => uni.stopPullDownRefresh())
})

/* ================= 跳转 ================= */

const goDetail = (d) => safeNavigateTo('/pages/demands/detail?id=' + encodeURIComponent(d.id))
const goHall = () => safeNavigateTo('/pages/demands/index')
const goBack = () => uni.navigateBack()

/* ================= 操作 ================= */

async function unfavorite(d) {
  const confirm = await new Promise((resolve) => {
    uni.showModal({ title: '取消收藏', content: '确定不再收藏这条需求吗？', success: (r) => resolve(r.confirm) })
  })
  if (!confirm) return
  try {
    await request({
      url: '/api/v1/demands/' + encodeURIComponent(d.id) + '/favorite',
      method: 'POST',
      data: { favorite: false },
    })
    list.value = list.value.filter((x) => x.id !== d.id)
    uni.showToast({ title: '已取消收藏', icon: 'none' })
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' })
  }
}
</script>

<style scoped>
.fav-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* 头部 */
.page-header {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-action { padding: 14rpx; }
.head-action-text { color: #0A66C2; font-size: 26rpx; font-weight: 600; }

/* 列表 */
.list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 28rpx 32rpx 16rpx;
}
.list-title { font-size: 36rpx; font-weight: 750; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

.post-list { padding: 0 32rpx 32rpx; }
.mine-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 26rpx;
  border: 1px solid #EEF1F4;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.mine-card + .mine-card { margin-top: 20rpx; }
.mine-card--active { opacity: .8; }

.tag-row { display: flex; gap: 10rpx; }
.tag {
  border-radius: 8rpx;
  padding: 6rpx 12rpx;
  font-size: 20rpx;
  line-height: 1;
}
.tag.blue { color: #0A66C2; background: #EAF3FB; }
.tag.green { color: #168A55; background: #E9F7F0; }
.tag.orange { color: #DB5F0D; background: #FFF0E6; }
.tag.red { color: #D92D20; background: #FEF3F2; }
.tag.gray { color: #667085; background: #F1F3F5; }

.post-title {
  display: block;
  font-size: 28rpx;
  line-height: 1.45;
  color: #17212B;
  font-weight: 700;
  margin: 16rpx 0 8rpx;
}
.post-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx 20rpx;
}
.post-meta-item { font-size: 22rpx; color: #667085; }

.mine-action-row {
  display: flex;
  gap: 32rpx;
  border-top: 1px solid #EEF1F4;
  margin-top: 22rpx;
  padding-top: 20rpx;
}
.action-link { color: #0A66C2; font-size: 24rpx; font-weight: 600; }
.action-link.danger { color: #D92D20; }

/* 骨架 */
.skl {
  border-radius: 8rpx;
  background: #EDF0F3;
}
.post-title-skl {
  height: 32rpx;
  margin: 18rpx 0;
  width: 70%;
}

/* 空状态 */
.state-panel {
  min-height: 560rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
}
.state-title { font-size: 28rpx; font-weight: 700; color: #17212B; }
.state-desc { margin: 12rpx 0 32rpx; font-size: 22rpx; color: #98A2B3; }
.state-btn {
  height: 72rpx;
  padding: 0 30rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  line-height: 72rpx;
}
</style>
