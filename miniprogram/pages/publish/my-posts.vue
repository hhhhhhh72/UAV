<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">我的发布</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="filterOpen = true">筛选</view>
    </view>

    <!-- 三张概览卡 -->
    <view class="pub-summary">
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('pending')">
        <text class="pub-summary-num">{{ counts.pending }}</text>
        <text class="pub-summary-label">审核中</text>
      </view>
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('live')">
        <text class="pub-summary-num">{{ counts.live }}</text>
        <text class="pub-summary-label">已发布</text>
      </view>
      <view class="pub-summary-item" hover-class="pub-summary-item--active" @tap="setTab('draft')">
        <text class="pub-summary-num">{{ counts.draft }}</text>
        <text class="pub-summary-label">待完善</text>
      </view>
    </view>

    <!-- 状态筛选胶囊 -->
    <scroll-view scroll-x class="pub-filter-scroll" :show-scrollbar="false">
      <view class="pub-filter-inner">
        <view
          v-for="tab in tabOrder"
          :key="tab"
          class="pub-filter-chip"
          :class="{ 'pub-filter-chip--active': postsTab === tab }"
          @tap="setTab(tab)"
        >{{ TAB_LABEL[tab] }}</view>
      </view>
    </scroll-view>

    <!-- 列表 / 空状态 -->
    <view class="pub-posts">
      <view v-if="filtered.length === 0" class="pub-empty">
        <view class="pub-empty-mark">⌁</view>
        <view class="pub-empty-title">暂无符合条件的发布</view>
        <view class="pub-empty-desc">更换筛选条件，或去发布新内容。</view>
      </view>
      <view
        v-for="post in filtered"
        :key="post.id"
        class="pub-post-card"
        hover-class="pub-post-card--active"
        @tap="openDetail(post)"
      >
        <view class="pub-post-top">
          <text class="pub-post-type">{{ post.label }}</text>
          <text class="pub-post-status" :class="statusClass(post.statusKey)">{{ post.status }}</text>
        </view>
        <view class="pub-post-title">{{ post.title }}</view>
        <view class="pub-post-meta">
          <text v-for="(m, i) in post.meta" :key="i" class="pub-meta-item">{{ m }}</text>
        </view>
        <view class="pub-post-foot">
          <text>{{ post.date }}</text>
          <text class="pub-post-foot-strong">{{ post.leads || post.note }}</text>
        </view>
      </view>
    </view>

    <!-- 类型筛选抽屉 -->
    <view v-if="filterOpen" class="pub-overlay" @tap="filterOpen = false">
      <view class="pub-sheet" @tap.stop>
        <view class="pub-grab"></view>
        <view class="pub-sheet-head">
          <view class="pub-sheet-head-title">筛选发布类型</view>
          <view class="pub-sheet-cancel" @tap="filterOpen = false">完成</view>
        </view>
        <view
          v-for="kind in kindOrder"
          :key="kind"
          class="pub-option"
          :class="{ 'pub-option--selected': postKind === kind }"
          @tap="pickKind(kind)"
        >
          <text>{{ KIND_LABEL[kind] }}</text>
          <text v-if="postKind === kind" class="pub-option-check">✓</text>
        </view>
      </view>
    </view>

    <!-- 底部黑色 toast -->
    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import {
  getPosts, TAB_ORDER, TAB_LABEL, KIND_ORDER, KIND_LABEL,
} from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const postsTab = ref('all')
const postKind = ref('all')
const filterOpen = ref(false)
const toast = ref('')
const toastTimer = ref(null)

const tabOrder = TAB_ORDER
const kindOrder = KIND_ORDER

const allPosts = ref([])

function refresh() {
  allPosts.value = getPosts()
}

const filtered = computed(() => {
  return allPosts.value.filter(
    (p) =>
      (postsTab.value === 'all' || p.statusKey === postsTab.value) &&
      (postKind.value === 'all' || p.type === postKind.value)
  )
})

const counts = computed(() => {
  const list = allPosts.value
  return {
    pending: list.filter((p) => p.statusKey === 'pending').length,
    live: list.filter((p) => p.statusKey === 'live').length,
    draft: list.filter((p) => p.statusKey === 'draft').length,
  }
})

function statusClass(key) {
  return 'status-' + key
}

function setTab(tab) {
  postsTab.value = tab
}

function pickKind(kind) {
  postKind.value = kind
  filterOpen.value = false
}

function openDetail(post) {
  uni.navigateTo({
    url: '/pages/publish/detail?id=' + post.id +
      '&tab=' + postsTab.value + '&kind=' + postKind.value,
  })
}

function goBack() {
  uni.navigateBack()
}

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

onLoad((options) => {
  initSafeTop()
  if (options && options.tab && TAB_LABEL[options.tab]) postsTab.value = options.tab
  if (options && options.kind && KIND_LABEL[options.kind]) postKind.value = options.kind
  refresh()
})

onShow(() => {
  // 从详情页返回后刷新（撤回/下架/删除后列表状态变化）
  refresh()
})
</script>

<style scoped>
@import './pub-style.css';
.pub-fade { opacity: 0.6; }
.pub-filter-inner {
  display: inline-flex;
  gap: 8px;
}
</style>
