<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarH + 44) + 'px' }">
    <u-nav-bar title="入驻企业" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起；左侧 CSS 放大镜，右侧"搜索"文字按钮） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="keyword"
            placeholder="搜索企业名称、能力、行业"
            placeholder-class="b-ph"
            confirm-type="search"
          />
          <text v-if="keyword" class="b-sclr" hover-class="tap-fade" :hover-stay-time="100" @tap="keyword = ''">×</text>
          <view class="b-sep" />
          <text class="b-sbtn">搜索</text>
        </view>
      </view>

      <!-- 筛选分段：行业（下划线 tab，对齐科技成果库；动态分类超宽 → 单行横向滚动） -->
      <scroll-view v-if="cats.length > 1" scroll-x :show-scrollbar="false" class="stages-scroll">
        <view class="stage-wrap">
          <view class="stages">
            <view
              v-for="c in cats"
              :key="c"
              class="stg"
              :class="{ on: activeCat === c }"
              @tap="pickStageTab(c)"
            >
              <text>{{ c === '' ? '全部' : c }}</text>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- ② 信息行：共 N 家 + 统计提示 -->
    <view class="ir">
      <text>共 <text class="irn">{{ filteredList.length }}</text> 家企业</text>
      <text class="ir-hint">{{ activeCat || '全部行业' }} · 会员 {{ memberCount }} · 覆盖 {{ industryCount }} 行业</text>
    </view>

    <!-- ③ 骨架屏：首次加载 -->
    <view v-if="loading && list.length === 0" class="skl">
      <view v-for="i in 4" :key="'sk' + i" class="skc">
        <view class="sk-row">
          <view class="sk-tag"></view>
          <view class="sk-bd">
            <view class="sk-l w60"></view>
            <view class="sk-l w40"></view>
          </view>
        </view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ④ 加载失败（无旧数据）：错误态 + 重试 -->
    <view v-else-if="errorMsg && list.length === 0" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="load">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑤ 空数据（全量无企业） -->
    <view v-else-if="list.length === 0" class="st">
      <u-empty description="暂无入驻企业">
        <text class="sth">企业完成入驻审核后将在此公示</text>
        <view class="stb" @tap="goRegister">申请入驻</view>
      </u-empty>
    </view>

    <!-- ⑥ 搜索/筛选无结果 -->
    <view v-else-if="filteredList.length === 0" class="st">
      <u-empty description="未找到相关企业">
        <text class="sth">换个关键词试试，或浏览全部入驻企业</text>
        <view class="stb" @tap="resetFilter">清除筛选</view>
      </u-empty>
    </view>

    <!-- ⑦ 企业列表：logo / 名称 / 认证状态 / 标签 / 简介 / 入驻时间 -->
    <view v-else class="cl">
      <view
        v-for="e in filteredList"
        :key="e.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @tap="openDetail(e)"
      >
        <view class="cell-top">
          <view class="ent-logo">
            <image v-if="e.logo" :src="resolveUrl(e.logo)" mode="aspectFill" class="ent-logo-img" @error="e.logo = ''" />
            <view v-else class="ent-logo-fallback">{{ e.name ? e.name.charAt(0) : '企' }}</view>
          </view>
          <view class="cell-main">
            <view class="cell-title-row">
              <text class="cell-title">{{ e.name }}</text>
              <text v-if="e.is_member" class="member-badge">会员</text>
            </view>
            <view class="cell-verified">
              <view class="verified-dot" />
              <text class="verified-text">协会已认证</text>
            </view>
            <view v-if="displayTags(e).length" class="tag-row">
              <text v-for="t in displayTags(e)" :key="t.label" class="type-tag" :class="t.blue ? 'tag--blue' : 'tag--gray'">{{ t.label }}</text>
              <text v-if="tagMore(e) > 0" class="tag-more">+{{ tagMore(e) }}</text>
            </view>
          </view>
        </view>

        <text v-if="e.description" class="c-desc">{{ e.description }}</text>

        <view class="cell-foot">
          <text class="cell-org">入驻 {{ formatDate(e.created_at) }}</text>
        </view>
      </view>

      <!-- 有旧数据时刷新失败：错误横幅 + 重试（保留旧数据） -->
      <view v-if="errorMsg" class="error-banner">
        <text>{{ errorMsg }}</text>
        <text class="error-retry" @tap="load">重试</text>
      </view>

      <!-- 入驻引导 -->
      <view class="foot-join" hover-class="tap-fade" :hover-stay-time="100" @tap="goRegister">
        <text class="foot-join-text">您的企业也想入驻？立即申请</text>
      </view>
    </view>

    <!-- ⑧ 底部固定申请入驻条 -->
    <view class="join-bar">
      <view class="join-btn" hover-class="join-btn-hover" hover-stay-time="100" @tap="goRegister">
        <text class="join-btn-text">申请企业入驻</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'
import { useReduceMotion } from '../../../utils/motion'

const loading = ref(false)
const list = ref([])
const errorMsg = ref('')
const statusBarH = ref(20)
const activeCat = ref('')
const keyword = ref('')
const { noMotion, checkMotion } = useReduceMotion()

const goBack = () => uni.navigateBack()

const goRegister = () => uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })

// 清除搜索与行业筛选
const resetFilter = () => {
  keyword.value = ''
  activeCat.value = ''
}

// 行业分段 tab（方案 A：非全部再点取消回全部）
const pickStageTab = (c) => {
  activeCat.value = activeCat.value === c && c !== '' ? '' : c
}

const openDetail = (e) => {
  uni.navigateTo({ url: '/pkg-eco/pages/enterprise/detail?id=' + encodeURIComponent(e.id) })
}

const load = async () => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/enterprises/public' })
    list.value = Array.isArray(res) ? res : []
    errorMsg.value = ''
  } catch (e) {
    // P1 修复：失败不再静默清空——空列表展示错误态+重试；有旧数据时保留并降级统计
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

onPullDownRefresh(async () => {
  await load()
  uni.stopPullDownRefresh()
})

const splitTags = (str) => {
  if (!str) return []
  return String(str).split(',').map((t) => t.trim()).filter(Boolean)
}
// 相对路径（存库格式）→ 完整 URL（预览格式）
const resolveUrl = (u) => {
  if (!u) return ''
  if (u.indexOf('http') === 0) return u
  return BASE_URL + u
}
const categoryList = (e) => splitTags(e.industry_category)
const tagList = (e) => splitTags(e.capability_tags)

// 卡片标签：分类（蓝）优先，不足 2 个补能力标签（灰），超出计数
const displayTags = (e) => {
  const cats = categoryList(e)
  const tags = tagList(e)
  const shown = []
  for (let i = 0; i < Math.min(2, cats.length); i++) shown.push({ label: cats[i], blue: true })
  for (let i = 0; shown.length < 2 && i < tags.length; i++) shown.push({ label: tags[i], blue: false })
  return shown
}
const tagMore = (e) => Math.max(0, categoryList(e).length + tagList(e).length - 2)

// Hero 统计：会员数 + 行业覆盖（industry_category 首个分类去重）
const memberCount = computed(() => list.value.filter((e) => e.is_member).length)
const industryCount = computed(() => {
  const set = new Set()
  list.value.forEach((e) => {
    const first = categoryList(e)[0]
    if (first) set.add(first)
  })
  return set.size
})

// 筛选 chips：全部 + 各企业行业分类去重（保留原始顺序）
const cats = computed(() => {
  const set = new Set()
  list.value.forEach((e) => categoryList(e).forEach((c) => set.add(c)))
  return ['', ...set]
})

// 关键词匹配：企业名称 / 简介 / 行业 / 能力 任一命中
const matchKeyword = (e) => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return true
  const hay = [e.name, e.description, e.industry_category, e.capability_tags]
    .filter(Boolean).join(' ').toLowerCase()
  return hay.includes(kw)
}

const filteredList = computed(() => {
  const base = activeCat.value
    ? list.value.filter((e) => categoryList(e).includes(activeCat.value))
    : list.value
  return base.filter(matchKeyword)
})

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
}

onLoad(async () => {
  try {
    statusBarH.value = uni.getSystemInfoSync().statusBarHeight || 20
  } catch (e) {
    // 默认 20
  }
  checkMotion()
  await load()
})
</script>

<style>
page {
  background: #fff;
}
</style>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: calc(84px + env(safe-area-inset-bottom));
}

.tap-fade { opacity: 0.85; }

/* ===== 白底头部 ===== */
.head-zone { background: #fff; }

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 双层投影 ===== */
.sbar { padding: 12px 12px 8px; background: #fff; }
.b-search {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
  display: flex;
  align-items: center;
  gap: 7px;
  box-sizing: border-box;
}
.b-search-ic { position: relative; width: 15px; height: 15px; flex: none; }
.ic-ring {
  width: 9px; height: 9px;
  border: 1.5px solid #98A2B3;
  border-radius: 50%;
  position: absolute; top: 0; left: 0;
}
.ic-bar {
  position: absolute; right: 0; bottom: 1px;
  width: 5px; height: 1.5px;
  background: #98A2B3;
  transform: rotate(45deg);
}
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; }
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 行业筛选分段（对齐科技成果库：下划线 tab；动态分类超宽单行横滑，不换行） ===== */
.stages-scroll { white-space: nowrap; width: 100%; }
.stage-wrap { position: relative; z-index: 42; }
.stages { display: flex; gap: 40rpx; padding: 4rpx 28rpx 16rpx; white-space: nowrap; }
.stg {
  position: relative;
  flex-shrink: 0;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 0 8rpx;
  font-size: 24rpx;
  color: #667085;
}
.stg.on { color: #074D92; font-weight: 600; }
.stg.on::after {
  content: '';
  position: absolute;
  left: 8rpx;
  right: 8rpx;
  bottom: 16rpx;
  height: 3rpx;
  border-radius: 2rpx;
  background: #074D92;
  animation: toc-in 0.22s ease-out;
}
@keyframes toc-in { from { transform: scaleX(0); } to { transform: scaleX(1); } }

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ===== 骨架屏 ===== */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
}
.sk-row { display: flex; align-items: center; gap: 8px; }
.sk-tag { width: 40px; height: 40px; border-radius: 8px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; flex: 1; min-width: 0; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 企业卡片：白上白——灰描边 + 柔和环境阴影，错峰入场 ===== */
.cl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px 12px; }
.card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease;
}
.card:nth-child(-n+6) { animation: cardIn .22s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

.cell-top { display: flex; align-items: flex-start; gap: 10px; }
.ent-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  overflow: hidden;
  background: #EAF3FB;
  flex: none;
}
.ent-logo-img { width: 100%; height: 100%; }
.ent-logo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: #0A66C2;
  background: #EAF3FB;
}
.cell-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.cell-title-row { display: flex; align-items: center; gap: 6px; }
.cell-title {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.member-badge {
  flex: none;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #0B6B41;
  background: #E9F7F0;
}
.cell-verified { display: flex; align-items: center; gap: 4px; }
.verified-dot { width: 6px; height: 6px; border-radius: 50%; background: #0B6B41; }
.verified-text { font-size: 11px; color: #0B6B41; font-weight: 500; }

/* 标签行：最多 2 个 + 溢出计数（分类蓝 / 能力灰，与详情页 chip 同规格：22rpx/8rpx 18rpx/8rpx 圆角/同色 token） */
.tag-row { display: flex; flex-wrap: wrap; gap: 14rpx; align-items: center; }
.type-tag { padding: 8rpx 18rpx; border-radius: 8rpx; font-size: 22rpx; line-height: 1.4; }
.tag--blue { background: #EAF3FB; color: #0A66C2; border: 1rpx solid rgba(10, 102, 194, 0.12); }
.tag--gray { background: #F1F3F5; color: #667085; border: 1rpx solid rgba(102, 112, 133, 0.1); }
.tag-more { padding: 8rpx 0 8rpx 8rpx; font-size: 22rpx; color: #98A2B3; }

/* 描述：两行截断 */
.c-desc {
  font-size: 12.5px;
  color: #667085;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

/* 底部入驻时间 */
.cell-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 2px;
  padding-top: 10px;
  border-top: 1px solid #F0F1F3;
}
.cell-org { font-size: 12px; color: #98A2B3; }

/* 已有数据时加载失败横幅 */
.error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 14px;
  margin: 0 12px;
  font-size: 12px;
  color: #B42318;
  background: #FEF0EF;
  border-radius: 8px;
}
.error-retry { color: #0A66C2; font-weight: 600; }

/* 入驻引导 */
.foot-join {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 14px 0 4px;
}
.foot-join-text { font-size: 13px; color: #0A66C2; font-weight: 600; }

/* ===== 底部固定申请入驻条（按钮规范：radius 50rpx + box-shadow，扁平主色） ===== */
.join-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 20;
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -1px 0 #F0F1F3;
  pointer-events: none;
}
.join-btn {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 44px;
  border-radius: 22px;
  background: #0A66C2;
  box-shadow: 0 6px 16px rgba(10, 102, 194, 0.24);
}
.join-btn-hover { opacity: 0.88; transform: scale(0.985); }
.join-btn-text {
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 1px;
}

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .stg.on::after { animation: none; }

@media (prefers-reduced-motion: reduce) {
  .stg { animation: none !important; transition: none !important; }
}
</style>
