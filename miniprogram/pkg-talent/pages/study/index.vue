<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- ① 搜索框（白上白：灰描边 + 双层投影浮起；左侧 CSS 搜索图标，右侧文字按钮） -->
    <view class="sbar">
      <view class="b-search">
        <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
        <input
          class="b-sinp"
          v-model="keyword"
          placeholder="搜索研学活动"
          placeholder-class="b-ph"
          confirm-type="search"
          @confirm="onSearch"
          @input="onSearch"
        />
        <text v-if="keyword" class="b-sclr" @click="clearKeyword">×</text>
        <view class="b-sep" />
        <text class="b-sbtn" @click="onSearch">搜索</text>
      </view>
    </view>

    <!-- ② 位置筛选（原生 picker：重庆市 + 区县，CSS 箭头，无定位图标） -->
    <picker
      mode="selector"
      :range="chongqingDistricts"
      :value="districtIndex"
      @change="onDistrictChange"
      class="dist-picker"
    >
      <view class="dist-bar">
        <view class="dist-left">
          <text class="dist-city">重庆市</text>
          <view class="dist-divider" />
          <text class="dist-value" :class="{ placeholder: !selectedDistrict }">{{ selectedDistrict || '全部区县' }}</text>
          <view class="dist-arr" />
        </view>
      </view>
    </picker>

    <!-- ③ 主题分类胶囊（fpill） -->
    <view class="fbar">
      <scroll-view scroll-x :show-scrollbar="false" class="fscroll">
        <view class="finner">
          <view
            v-for="pill in themePills"
            :key="pill.value"
            class="fpill"
            :class="{ on: activeTheme === pill.value }"
            hover-class="fpill-press"
            :hover-stay-time="100"
            @click="switchTheme(pill.value)"
          >
            <text class="fpv">{{ pill.label }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- ④ 信息行：共 N 个活动 + 当前筛选 -->
    <view class="ir">
      <text>共 <text class="irn">{{ displayList.length }}</text> 个活动</text>
      <text class="ir-hint">{{ filterHint }}</text>
    </view>

    <!-- ⑤ 骨架屏：首次加载 -->
    <view v-if="loading && displayList.length === 0" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-tag"></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w60"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ⑥ 空 / 错误 -->
    <view v-else-if="!loading && displayList.length === 0" class="st">
      <u-empty :description="errorMsg || '暂无研学活动'">
        <view v-if="errorMsg" class="stb" @tap="loadData(true)">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑦ 研学卡片列表 -->
    <view v-else class="cl">
      <view
        v-for="item in displayList"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
        @click="openDetail(item)"
      >
        <!-- 封面：有真实图显示图，无图兜底主题色 -->
        <view class="card-cover" :style="!item.cover_image ? { background: themeInfo(item).gradient } : {}">
          <image
            v-if="item.cover_image"
            :src="item.cover_image"
            mode="aspectFill"
            class="cover-img"
            lazy-load
          />
          <image v-else :src="coverByTheme(item)" mode="aspectFill" class="cover-img" lazy-load />
          <view class="status-badge" :style="{ background: statusStyle(item).bg, color: statusStyle(item).color }">
            {{ statusStyle(item).text }}
          </view>
        </view>

        <!-- 卡片信息区 -->
        <view class="card-body">
          <text class="card-name">{{ item.title || '未知活动' }}</text>

          <view class="card-info">
            <view class="info-row">
              <text class="info-label">时间</text>
              <text class="info-value">{{ dateRange(item) }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">地点</text>
              <text class="info-value ellipsis">{{ item.location || '待定' }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">时长</text>
              <text class="info-value">{{ item.duration || '待定' }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">名额</text>
              <text class="info-value"><text class="capacity-value">{{ capacityText(item) }}</text></text>
            </view>
          </view>
        </view>

        <!-- 卡片底部 -->
        <view class="card-footer">
          <text class="detail-hint">点击查看详情</text>
          <view class="foot-arr" />
        </view>
      </view>

      <!-- 列表底部 -->
      <view v-if="!loading && displayList.length > 0" class="list-footer">
        <view class="footer-line" />
        <text class="footer-text">没有更多了</text>
        <view class="footer-line" />
      </view>
      <view style="height: 40rpx" />
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { useReduceMotion } from '../../../utils/motion'

const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关

const keyword = ref('')
const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const list = ref([])
const page = ref(1)
const pageSize = 50
const hasMore = ref(true)

// ── 地区筛选（重庆 38 个区县） ──────────────────────────
const chongqingDistricts = ['全部区县', '渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区', '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县', '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县']
const districtIndex = ref(0)
const selectedDistrict = computed(() => chongqingDistricts[districtIndex.value])
const onDistrictChange = (e) => { districtIndex.value = Number(e.detail.value); page.value = 1; loadData(true) }

// ── 主题分类 Pills ────────────────────
const themePills = [
  { label: '全部', value: 'all' },
  { label: '科普研学', value: 'science' },
  { label: '产业研学', value: 'industry' },
  { label: '实践研学', value: 'practice' },
  { label: '职业研学', value: 'career' },
]
const activeTheme = ref('all')
const switchTheme = (v) => { activeTheme.value = v; page.value = 1; loadData(true) }

// ── 主题推断（title/description 关键词）──
const THEMES = [
  { key: ['职业', '院校', '学院', '专业', '开放日'], value: 'career', label: '职业研学', short: '职业', icon: '职', gradient: 'linear-gradient(135deg,#FF8E3C,#FFB36B)' },
  { key: ['实践', '实训', '训练营', '穿越机', '巡检', '应急救援', '测绘', '实战', '试飞'], value: 'practice', label: '实践研学', short: '实践', icon: '实', gradient: 'linear-gradient(135deg,#065F46,#06B6D4)' },
  { key: ['科普', '科技', '科学', '航模', '体验', '组装'], value: 'science', label: '科普研学', short: '科普', icon: '学', gradient: 'linear-gradient(135deg,#0A1F44,#1E5EFF)' },
  { key: ['产业', '低空经济', '企业', '龙头'], value: 'industry', label: '产业研学', short: '产业', icon: '产', gradient: 'linear-gradient(135deg,#6D28D9,#DB2777)' },
]
const themeInfo = (item) => {
  const title = item.title || ''
  const desc = item.description || ''
  // 先匹配标题（更权威），再匹配描述
  for (const t of THEMES) {
    if (t.key.some((k) => title.includes(k))) return t
  }
  for (const t of THEMES) {
    if (t.key.some((k) => desc.includes(k))) return t
  }
  return { value: 'science', label: '科普研学', short: '研学', icon: '研', gradient: 'linear-gradient(135deg,#0A1F44,#1E5EFF)' }
}

// ── 本地兜底封面：无 cover_image 时按主题映射静态图 ──
const COVER_BY_THEME = { career: 1, practice: 2, science: 3, industry: 1 }
const coverByTheme = (item) => '/static/images/study/cover-' + (COVER_BY_THEME[themeInfo(item).value] || 1) + '.jpg'

// ── 状态徽章 ──────────────────────────
const statusStyle = (item) => {
  const s = item.status || 'active'
  if (s === 'closed') return { text: '已结束', bg: '#EEF0F4', color: '#ADB8C7' }
  if (s === 'draft') return { text: '即将开始', bg: 'rgba(239,68,68,.1)', color: '#EF4444' }
  return { text: '招募中', bg: 'rgba(255,142,60,.12)', color: '#FF8E3C' }
}

// ── 列表：过滤 + 筛选 ──────────────────
const baseList = computed(() => list.value.filter((it) => it.status === 'active' || it.status === ''))
const filteredByTheme = computed(() => {
  if (activeTheme.value === 'all') return baseList.value
  return baseList.value.filter((it) => themeInfo(it).value === activeTheme.value)
})
const displayList = computed(() => {
  if (!selectedDistrict.value || selectedDistrict.value === '全部区县') return filteredByTheme.value
  return filteredByTheme.value.filter((it) => (it.location || '').includes(selectedDistrict.value.slice(0, 2)))
})

// ── 信息行右侧提示（仅展示当前筛选状态，不改任何筛选逻辑）──
const filterHint = computed(() => {
  const d = selectedDistrict.value === '全部区县' ? '' : selectedDistrict.value
  const t = activeTheme.value === 'all' ? '' : (themePills.find((p) => p.value === activeTheme.value) || {}).label
  return [d, t].filter(Boolean).join(' · ') || '全部研学'
})

// ── 卡片字段 ──────────────────────────
const capacityText = (item) => {
  const c = item.capacity
  if (c == null || c <= 0) return '不限'
  return `${c} 人`
}

// 时间范围：start_date ~ end_date（ISO 时间戳；零值时间视为未设置）
function dateRange(item) {
  const s = item.start_date
  const e = item.end_date
  const fmt = (v) => {
    if (!v) return ''
    const d = new Date(v)
    if (isNaN(d.getTime())) return ''
    if (d.getFullYear() <= 1) return ''
    const p = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}年${p(d.getMonth() + 1)}月${p(d.getDate())}日`
  }
  const ss = fmt(s)
  const ee = fmt(e)
  if (ss && ee && ss !== ee) return `${ss}-${ee.replace(/^\d{4}年/, '')}`
  return ss || ee || '待定'
}

function openDetail(item) {
  uni.setStorageSync('study_tour_detail', item)
  uni.navigateTo({ url: '/pkg-talent/pages/study/detail?id=' + encodeURIComponent(item.id) })
}

// ── 搜索 ──────────────────────────────
const clearKeyword = () => { keyword.value = ''; onSearch() }
var searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(function () { page.value = 1; loadData(true) }, 300)
}

// ── 数据加载 ──────────────────────────
async function loadData(reset) {
  if (reset === undefined) reset = true
  if (reset) { page.value = 1; hasMore.value = true; loading.value = true }
  else { loadingMore.value = true }
  errorMsg.value = ''
  try {
    var params = { page: page.value, page_size: pageSize }
    if (keyword.value) params.keyword = keyword.value
    var res = await request({ url: '/api/v1/study/tours', data: params })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || []
    var total = (data && data.total) != null ? data.total : items.length
    if (reset) { list.value = items } else { list.value = list.value.concat(items) }
    hasMore.value = list.value.length < total
  } catch (e) {
    if (reset) errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

onLoad(() => { checkMotion(); loadData(true) })
onPullDownRefresh(function () {
  loadData(true).then(function () { uni.stopPullDownRefresh() })
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
  padding-bottom: 40rpx;
}

/* ═══ ① 搜索框：白上白——纯白填充 + 灰描边 + 双层投影 ═══ */
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

/* ═══ ② 位置筛选条：fpill 同款白底胶囊条 ═══ */
.dist-picker { display: block; margin: 0 12px 8px; }
.dist-bar {
  display: flex;
  align-items: center;
  min-height: 40px;
  padding: 0 12px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  box-sizing: border-box;
}
.dist-left { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.dist-city { font-size: 13px; font-weight: 600; color: #0A66C2; flex: none; }
.dist-divider { width: 1px; height: 14px; background: #DDE1E6; flex: none; }
.dist-value { font-size: 13px; color: #344054; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dist-value.placeholder { color: #98A2B3; }
.dist-arr {
  width: 7px; height: 7px;
  border-right: 1.5px solid #667085;
  border-bottom: 1.5px solid #667085;
  transform: rotate(45deg);
  flex: none; margin-left: auto;
}

/* ═══ ③ 主题分类胶囊（fpill） ═══ */
.fbar { padding: 0 12px 8px; background: #fff; }
.fscroll { width: 100%; white-space: nowrap; }
.finner { display: inline-flex; gap: 8px; padding-bottom: 4px; }
.fpill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 18px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpill-press { transform: scale(0.95); opacity: 0.85; }
.fpv { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ═══ ④ 信息行 ═══ */
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
  animation: fadeUp .25s ease-out backwards;
  animation-delay: 60ms;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-hint { font-size: 12px; color: #98A2B3; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ═══ ⑤ 骨架屏 ═══ */
.skl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px; }
.skc {
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  overflow: hidden;
}
.sk-tag { height: 140rpx; background: #EDF0F3; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; padding: 12px 14px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ═══ ⑥ 空 / 错误 ═══ */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ═══ ⑦ 研学卡片 ═══ */
.cl { display: flex; flex-direction: column; gap: 8px; padding: 0 12px 12px; }
.card {
  position: relative;
  background: #fff;
  border: 1px solid #E4E7EC;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(16, 24, 40, 0.06);
  overflow: hidden;
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

/* 封面 */
.card-cover { position: relative; height: 140rpx; overflow: hidden; background: #F4F6F8; }
.cover-img { position: absolute; inset: 0; width: 100%; height: 100%; }
.status-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.5;
}

/* 信息区 */
.card-body { padding: 12px 14px 8px; }
.card-name {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.45;
  margin-bottom: 10px;
}
.card-info { display: flex; flex-direction: column; gap: 6px; }
.info-row { display: flex; align-items: center; }
.info-label { width: 44px; flex: none; font-size: 12px; color: #98A2B3; }
.info-value { flex: 1; min-width: 0; font-size: 13px; color: #344054; }
.info-value.ellipsis { overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.capacity-value { color: #0B6B41; font-weight: 600; }

/* 卡片底部 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 14px;
  padding: 10px 0;
  border-top: 1px solid #F0F1F3;
}
.detail-hint { font-size: 12px; color: #0A66C2; font-weight: 600; }
.foot-arr {
  width: 7px; height: 7px;
  border-top: 1.5px solid #0A66C2;
  border-right: 1.5px solid #0A66C2;
  transform: rotate(45deg);
}

/* ═══ 列表底部 ═══ */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 16px 0 4px;
}
.footer-line { width: 40px; height: 1px; background: #E4E7EC; }
.footer-text { font-size: 12px; color: #98A2B3; }

/* ═══ 减弱动效（无障碍） ═══ */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
