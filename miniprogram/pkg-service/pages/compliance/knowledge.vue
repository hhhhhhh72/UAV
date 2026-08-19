<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="合规知识库" show-back :fixed="true" @back="goBack" />

    <!-- 固定头部：搜索 + 分类胶囊（一体吸顶） -->
    <view class="sticky-head" :style="{ top: (statusBarHeight + 44) + 'px' }">

      <!-- 搜索框（对齐研发难题广场） -->
      <view class="sbar">
        <view class="b-search">
          <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
          <input class="b-sinp" v-model="q" placeholder="搜索文档标题、关键词" placeholder-class="b-ph" confirm-type="search" @input="onSearch" />
          <text v-if="q" class="b-sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</text>
          <view class="b-sep"></view>
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 分类胶囊：点击展开对应分类的文档列表（手风琴） -->
      <view class="fbar">
        <view v-for="s in sections" :key="s.key" class="fpill" :class="{ on: activeCollapse === s.key }" @tap="onCollapseChange(s.key)">
          <text class="fpv">{{ s.title }}</text><text class="farr">▾</text>
        </view>
      </view>
    </view>

    <!-- Banner 渐变卡 -->
    <view class="banner">
      <view class="banner-icon">规</view>
      <view class="banner-info">
        <text class="banner-title">合规护航，稳健经营</text>
        <text class="banner-sub">政策 · 法规 · 标准 · 指南 一站查阅</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 当前分类 -->
      <view class="ir">
        <text>共 <text class="irn">{{ matchedTotal }}</text> 项合规文档</text>
        <text class="ir-hint">{{ activeLabel }}</text>
      </view>

      <!-- 骨架 -->
      <view v-if="loading" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w70"></view>
          </view>
        </view>
      </view>

      <!-- 错误 -->
      <view v-else-if="errorMsg" class="st">
        <u-empty description="加载失败，请检查网络">
          <view class="stb" @tap="fetchAll">重新加载</view>
        </u-empty>
      </view>

      <!-- 空 -->
      <view v-else-if="allEmpty" class="st">
        <u-empty description="暂无合规知识文档">
          <text class="sth">文档收录后即可在此查阅</text>
        </u-empty>
      </view>

      <!-- 搜索无匹配 -->
      <view v-else-if="q && matchedTotal === 0" class="st">
        <u-empty description="无匹配文档">
          <text class="sth">试试其他关键词</text>
          <view class="stb" @tap="clearSearch">清除搜索</view>
        </u-empty>
      </view>

      <!-- 折叠分类列表 -->
      <view v-else class="cl">
        <view v-for="s in sections" :key="s.key" class="sec-group">
          <view class="sec-head" hover-class="tap-scale" hover-start-time="0" hover-stay-time="120" @tap="onCollapseChange(s.key)">
            <view class="sec-bar" :style="{ background: secColor(s).tagC }" />
            <view class="sec-head-main">
              <text class="sec-title">{{ s.title }}</text>
              <text class="sec-count">{{ visibleDocs(s).length }} 篇</text>
            </view>
            <text class="sec-arrow" :class="{ expanded: activeCollapse === s.key }">▾</text>
          </view>

          <view v-if="activeCollapse === s.key" class="sec-panel">
            <view v-if="s.loading" class="panel-state">
              <view class="sk-mini">
                <view class="sk-l w80"></view>
                <view class="sk-l w60"></view>
              </view>
            </view>
            <view v-else-if="s.error" class="panel-state">
              <text class="panel-err-t">加载失败</text>
              <text class="panel-err-r" @tap="fetchSection(s)">重试</text>
            </view>
            <view v-else-if="visibleDocs(s).length === 0" class="panel-state">
              <text class="panel-err-t">{{ q ? '无匹配文档' : '暂无文档' }}</text>
            </view>
            <view v-else class="d-list">
              <view v-for="doc in visibleDocs(s)" :key="doc.id" class="d-card" hover-class="tap-scale" hover-start-time="0" hover-stay-time="120" @tap="openDoc(doc)">
                <view class="c-bar" :style="{ background: secColor(s).tagC }" />
                <view class="c-badges">
                  <text class="c-tag" :style="{ color: secColor(s).tagC, background: secColor(s).tagBg }">{{ doc.category || s.title }}</text>
                  <text class="c-st" :class="docSt(doc).cls">{{ docSt(doc).label }}</text>
                </view>
                <text class="ct">{{ doc.title || '--' }}</text>
                <text v-if="doc.summary" class="c-desc">{{ doc.summary }}</text>
                <view class="c-meta">
                  <text>{{ doc.publisher || '协会发布' }}</text>
                  <text class="c-dot">·</text>
                  <text>{{ fmtDate(doc.publish_date || doc.created_at) }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPageScroll, onUnload } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { useReduceMotion } from '../../../utils/motion'

const SEARCH_DEBOUNCE_MS = 250 // 搜索防抖：停顿 250ms 后再过滤（防每键整表重渲染）

/* ===== 静态配置 ===== */
const sectionConfigs = [
  { key: 'policy', title: '政策', value: '政策' },
  { key: 'regulation', title: '法规', value: '法规' },
  { key: 'standard', title: '标准', value: '标准' },
  { key: 'guide', title: '指南', value: '指南' },
]
/* 分类强调色：折叠头色条 / 文档左缘色条 / 分类 tag 的视觉锚点（对比度 ≥4.5:1） */
const SEC_COLOR = {
  policy: { tagC: '#0d47a1', tagBg: '#E3EDF9' },
  regulation: { tagC: '#1a237e', tagBg: '#E7E9F4' },
  standard: { tagC: '#004d40', tagBg: '#E4F2EF' },
  guide: { tagC: '#B54708', tagBg: '#FDEEE4' },
}
const SEC_COLOR_DEFAULT = { tagC: '#344054', tagBg: '#EEF1F4' }

/* ===== 状态 ===== */
const sections = ref(sectionConfigs.map(function (cfg) {
  return {
    key: cfg.key,
    title: cfg.title,
    value: cfg.value,
    list: [],
    loading: false,
    error: false,
    loaded: false,
  }
}))
const activeCollapse = ref('')
const loading = ref(false)
const errorMsg = ref('')
const q = ref('')
const statusBarHeight = ref(20)
const showBt = ref(false)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）
let searchT = null // 搜索防抖定时器（onUnload 清理）

/* ===== 派生 ===== */
const allEmpty = computed(() => sections.value.every(function (s) {
  return s.loaded && s.list.length === 0 && !s.error
}))
/* 搜索关键词（标题/摘要/发布方/分类）；空关键词返回全量 */
const visibleDocs = (s) => {
  const kw = (q.value || '').trim().toLowerCase()
  if (!kw) return s.list
  return s.list.filter((d) => ((d.title || '') + ' ' + (d.summary || '') + ' ' + (d.publisher || '') + ' ' + (d.category || '')).toLowerCase().includes(kw))
}
const matchedTotal = computed(() => sections.value.reduce((n, s) => n + visibleDocs(s).length, 0))
const activeLabel = computed(() => {
  const s = sections.value.find((x) => x.key === activeCollapse.value)
  return s ? s.title + ' · ' + visibleDocs(s).length + ' 篇' : '展开分类查看'
})
const secColor = (s) => SEC_COLOR[s.key] || SEC_COLOR_DEFAULT
const fmtDate = (d) => (d ? String(d).slice(0, 10) : '')
const docSt = (doc) => {
  const s = String(doc.status || '').toLowerCase()
  if (!s || s === 'published') return { label: '已发布', cls: 'st-open' }
  if (s === 'draft') return { label: '草稿', cls: 'st-closed' }
  return { label: doc.status, cls: 'st-closed' }
}

/* ===== 数据（逻辑保留：并行拉取四分类 / 懒加载 / 弹窗展示） ===== */
const fetchAll = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    await Promise.all(sections.value.map((section) => fetchSection(section)))
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}
const fetchSection = async (section) => {
  if (section.loaded && section.list.length > 0) return
  section.loading = true
  section.error = false
  try {
    const res = await request({ url: '/api/v1/compliance-docs', data: { category: section.value } })
    const data = Array.isArray(res) ? res : (res && res.items) || []
    section.list = Array.isArray(data) ? data : (data && data.items) || []
    section.loaded = true
  } catch (e) {
    section.error = true
  } finally {
    section.loading = false
  }
}
/* 折叠面板（手风琴）：点选即展开；首次展开时懒加载该分类 */
const onCollapseChange = (name) => {
  activeCollapse.value = name
  if (name && name.length > 0) {
    const section = sections.value.find((s) => s.key === name)
    if (section && !section.loaded && !section.loading) {
      fetchSection(section)
    }
  }
}
const openDoc = (doc) => {
  uni.showModal({
    title: doc.title || '文档内容',
    content: doc.summary || '暂无详细内容',
    showCancel: false,
    confirmText: '知道了',
  })
}

/* ===== 搜索（客户端过滤；搜索时补齐未加载分类，保证全量检索） ===== */
const onSearch = () => {
  clearTimeout(searchT)
  searchT = setTimeout(() => {
    if ((q.value || '').trim()) {
      sections.value.forEach((s) => {
        if (!s.loaded && !s.loading) fetchSection(s)
      })
    }
  }, SEARCH_DEBOUNCE_MS)
}
const clearSearch = () => { clearTimeout(searchT); q.value = '' }

/* ===== 其他 ===== */
const goBack = () => uni.navigateBack()
const scrollToTop = () => uni.pageScrollTo({ scrollTop: 0, duration: 300 })

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  fetchAll()
})
onPageScroll((e) => {
  showBt.value = (e?.scrollTop ?? 0) > 400
})
onUnload(() => {
  clearTimeout(searchT)
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
  padding-bottom: 40px;
}

/* ===== 固定头部（一体吸顶） ===== */
.sticky-head {
  position: sticky;
  z-index: 40;
  background: #fff;
}

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 极淡灰投影 ===== */
.sbar {
  padding: 12px 12px 8px;
  background: #fff;
}
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
.b-search-ic { width: 15px; height: 15px; flex: none; }
.b-sinp { flex: 1; min-width: 0; background: transparent; font-size: 13px; color: #17212B; }
.b-ph { color: #667085; }
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; }
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 分类胶囊 ===== */
.fbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 12px;
  background: #fff;
}
.fpill {
  flex: 1;
  min-width: 0;
  min-height: 40px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  overflow: hidden;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.farr { font-size: 11px; color: #667085; flex: none; }

/* ===== Banner 渐变卡 ===== */
.banner {
  margin: 12px 14px;
  padding: 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #0A66C2 0%, #074D92 100%);
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
  position: relative;
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
.banner::after {
  content: '';
  position: absolute;
  top: -30%;
  right: -20%;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.08) 0%, transparent 70%);
}
.banner-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}
.banner-info { flex: 1; min-width: 0; position: relative; z-index: 1; }
.banner-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; display: block; line-height: 1.3; color: #fff; }
.banner-sub { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: block; }

/* ===== 白色板块（信息行 + 列表） ===== */
.section {
  margin-top: 0;
  padding: 0;
}

/* ===== 信息行 ===== */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { color: #667085; font-weight: 500; padding: 8px 4px 8px 12px; }

/* ===== 折叠分类列表 ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.sec-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.sec-head {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 13px 14px 13px 18px;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.sec-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  border-radius: 10px 0 0 10px;
}
.sec-head-main { display: flex; align-items: baseline; gap: 8px; min-width: 0; }
.sec-title { font-size: 15px; font-weight: 700; color: #17212B; }
.sec-count { font-size: 12px; color: #667085; flex: none; }
.sec-arrow {
  font-size: 12px;
  color: #667085;
  flex: none;
  transition: transform .25s cubic-bezier(0.16, 1, 0.3, 1);
}
.sec-arrow.expanded { transform: rotate(180deg); color: #0A66C2; }

/* 展开面板：文档卡片列表 */
.sec-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  animation: panelIn .28s cubic-bezier(.32, .72, 0, 1);
}
.panel-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 18px 0;
}
.sk-mini { width: 100%; display: flex; flex-direction: column; gap: 8px; padding: 6px 4px; box-sizing: border-box; }
.panel-err-t { font-size: 12px; color: #667085; }
.panel-err-r { font-size: 13px; color: #0A66C2; font-weight: 500; padding: 6px 12px; }

/* ===== 文档卡片（参考页卡片体系：左缘色条 + 顶部徽章 + 标题 + 描述 + 元信息） ===== */
.d-list { display: flex; flex-direction: column; gap: 8px; }
.d-card {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 14px;
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
}
.c-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  border-radius: 10px 0 0 10px;
}
.c-badges { display: flex; gap: 6px; }
.c-tag, .c-st {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
}
.c-tag { color: #074D92; background: #EAF3FB; }
.c-st.st-open { color: #0B6B41; background: #E9F7F0; }
.c-st.st-closed { color: #5D6B82; background: #EEF1F4; }
.ct {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-desc {
  font-size: 12.5px;
  color: #667085;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.c-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #667085;
}
.c-dot { color: #DDE1E6; }

/* ===== 骨架 ===== */
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
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w70 { width: 70%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 回到顶部 ===== */
.bt {
  position: fixed;
  bottom: 90px;
  right: 16px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 35;
  opacity: 0;
  transform: scale(0.5);
  pointer-events: none;
  transition: opacity 0.2s, transform .35s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 20px;
  color: #666;
}
.bt.show { opacity: 1; transform: scale(1); pointer-events: auto; }
.bt:active { transform: scale(.92); transition: transform .08s linear; }

/* ===================== 动效规范（对齐全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸 color/background 过渡允许）
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) 松手柔顺减速 + ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* Banner 内部微编排：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，总 340ms ≤ 400ms */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }
/* Banner 单次扫光（非循环装饰，100ms 起播 280ms 线性，380ms 内收完） */
.banner::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(100deg, transparent 0%, rgba(255, 255, 255, 0.22) 50%, transparent 100%);
  transform: translateX(-150%) skewX(-20deg);
  animation: shineOnce .28s linear 100ms backwards;
  pointer-events: none;
}
@keyframes shineOnce {
  from { transform: translateX(-150%) skewX(-20deg); }
  to { transform: translateX(320%) skewX(-20deg); }
}

/* 信息行：卡片入场前落位 */
.ir { animation: fadeUp .25s ease-out backwards; animation-delay: 60ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* 分类折叠头入场：前 4 组依次淡入上移（每 30ms 错峰，总 ≤280ms） */
.sec-group .sec-head { animation: cardIn .22s ease-out backwards; }
.sec-group:nth-child(1) .sec-head { animation-delay: 60ms; }
.sec-group:nth-child(2) .sec-head { animation-delay: 90ms; }
.sec-group:nth-child(3) .sec-head { animation-delay: 120ms; }
.sec-group:nth-child(4) .sec-head { animation-delay: 150ms; }
/* 分类色条与头部同拍"点亮"（scaleY 顶部抽出） */
.sec-group .sec-bar { animation: barIn .2s ease-out backwards; }
.sec-group:nth-child(1) .sec-bar { animation-delay: 60ms; }
.sec-group:nth-child(2) .sec-bar { animation-delay: 90ms; }
.sec-group:nth-child(3) .sec-bar { animation-delay: 120ms; }
.sec-group:nth-child(4) .sec-bar { animation-delay: 150ms; }
@keyframes barIn { from { opacity: 0; transform: scaleY(.3); } to { opacity: 1; transform: scaleY(1); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
@keyframes panelIn { from { opacity: 0; transform: translateY(-6px); } to { opacity: 1; transform: translateY(0); } }

/* 文档卡片入场：展开面板内前 6 项错峰淡入 */
.d-card { animation: docIn .2s ease-out backwards; }
.d-card:nth-child(1) { animation-delay: 30ms; }
.d-card:nth-child(2) { animation-delay: 50ms; }
.d-card:nth-child(3) { animation-delay: 70ms; }
.d-card:nth-child(4) { animation-delay: 90ms; }
.d-card:nth-child(5) { animation-delay: 110ms; }
.d-card:nth-child(6) { animation-delay: 130ms; }
@keyframes docIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 交互反馈：可点元素按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位） */
.sec-head { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.d-card { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.tap-scale { transform: scale(.97); opacity: .92; transition-duration: .1s; transition-timing-function: linear; }
.b-sclr:active { opacity: .6; }
.b-sbtn { transition: opacity .2s ease; }
.b-sbtn:active { opacity: .5; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }
.panel-err-r { transition: opacity .2s ease; }
.panel-err-r:active { opacity: .6; }

/* ===== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用 ===== */
.page.no-motion .sec-head,
.page.no-motion .sec-bar,
.page.no-motion .d-card,
.page.no-motion .banner,
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after,
.page.no-motion .ir { animation: none; }
.page.no-motion .sec-panel { animation: panelFadeIn .22s ease-out; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .tap-scale { transform: none !important; }
.page.no-motion .sec-arrow { transition: none; }
.page.no-motion .stb:active,
.page.no-motion .panel-err-r:active,
.page.no-motion .bt:active,
.page.no-motion .b-sclr:active,
.page.no-motion .b-sbtn:active { transform: none; }
@keyframes panelFadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
