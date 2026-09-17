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
import { TYPES, getPosts, draftPosts, pruneStalePointers, MINE_PAGE_SIZE } from '../../utils/publishData'
import { request, authStorage, requireLogin } from '../../utils/request'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop()

// 入口已合并：原先「发布服务能力」与「发布商品设备」是两个入口做同一件事
// （服务能力是纯展示、商品可下单，两条路字段与状态机都不一样，用户看到两个入口做同一件事）。
// 现在服务能力就是 prod_type 为服务类目（维修/航拍/试飞/检测标定/空域协调）的商品，
// 统一从「发布商品设备」进——所以这里少一个入口，但可发的东西反而更多了。
const typeCards = [
  { key: 'demand', name: '发布需求', desc: '发布具体项目，获得飞手与服务商报价', icon: '/static/publish/demand.svg' },
  { key: 'product', name: '发布商品与服务', desc: '整机、零部件，或航拍/试飞/检测/维修等服务能力，都可下单交易', icon: '/static/publish/product.svg' },
  { key: 'course', name: '发布培训课程', desc: '用课程、证书、日期与招生信息回答学员的核心问题', icon: '/static/publish/course.svg' },
]

const allPosts = ref([])
const backendCount = ref(0)
// 后端拉取是否成功：失败时退化为纯本地口径，避免整页显示 0
const backendOK = ref(false)
const draftCount = computed(() => draftPosts().length)

// "全部发布"的条数 = 本地记录 + 后端记录，**但两者不能直接相加**。
//
// 本地记录分两类，计数口径不同：
//  1. 带 backendId 的：它只是后端实体的本地指针（发布成功后留的根记录）。
//     后端拉取成功时**一律不本地计数**——后端那次请求已经算过它了。
//  2. 不带 backendId 的：纯本地记录（未提交），后端看不到，必须本地算。
//
// 早先写的是"指针的 id 出现在后端集合里才跳过"，那只解决了"同一条数两遍"，
// 解决不了"实体已被删除"：老指针的 id 不在后端集合里，于是被留下来又数一次。
// 表现就是——删掉一条老需求、再发一条新的，显示 2 条（死指针 1 + 新需求 1）。
const totalCount = computed(() => {
  const localKeep = allPosts.value.filter((p) => !p.backendId || !backendOK.value)
  return localKeep.length + backendCount.value
})

// 我的发布计数 = 本地记录 + 后端已提交（需求/商品/服务/课程 mine=1；未登录后端返回空）
async function refresh() {
  allPosts.value = getPosts()
  try {
    // 服务能力已并入商品表（migration 000110）：商品接口返回的**已经包含**服务类目，
    // 再单独请求一次 /service-listings 会把同一批数据数两遍。
    const [dRes, pRes, cRes] = await Promise.all([
      request({ url: '/api/v1/demands', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/products', data: { mine: '1', page: 1, page_size: 100 } }),
      request({ url: '/api/v1/training-courses', data: { mine: '1', page: 1, page_size: 100 } }),
    ])
    const dList = Array.isArray(dRes) ? dRes : dRes?.data || []
    const pList = Array.isArray(pRes) ? pRes : pRes?.data || []
    const cList = Array.isArray(cRes) ? cRes : cRes?.data || []
    backendCount.value = dList.length + pList.length + cList.length
    backendOK.value = true
    // 后端已确认不存在的实体：本地死指针顺手清掉（只在三个接口都完整返回时清）。
    // 清完重读一次本地，保证下面的本地口径与存储一致。
    const pruned = pruneStalePointers(
      [...dList, ...pList, ...cList].map((x) => String((x && x.id) || '')),
      { complete: [dList, pList, cList].every((l) => l.length < MINE_PAGE_SIZE) }
    )
    if (pruned > 0) allPosts.value = getPosts()
  } catch (e) {
    backendCount.value = 0
    backendOK.value = false
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
  // 未登录每次进入都拦截（登录页返回仍会再次跳转，必须登录后才能使用发布功能）
  if (!requireLogin('请先登录后再发布')) return
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
