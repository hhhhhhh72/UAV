<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- 原生导航栏显示"政策资讯"（pages.json 未设 navigationStyle:custom），
         不再自绘 topbar，彻底移除 env(safe-area-inset-top) 安全区问题 -->

    <!-- 固定头部：搜索 + 分类筛选（一体吸顶，位于原生导航栏之下） -->
    <view class="sticky-head">
      <!-- 搜索框（参考页样式） -->
      <view class="sbar">
        <view class="b-search">
          <image class="b-search-ic" src="/static/home/icons/search.svg" mode="aspectFit" />
          <input
            v-model="searchText"
            class="b-sinp"
            placeholder="搜索政策资讯、关键词"
            placeholder-class="b-ph"
            confirm-type="search"
            @input="onSearchInput"
            @confirm="onSearchConfirm"
          />
          <text v-if="searchText" class="b-sclr" aria-role="button" aria-label="清除搜索" @tap="clearSearch">×</text>
          <view class="b-sep"></view>
          <text class="b-sbtn" aria-role="button" @tap="onSearchConfirm">搜索</text>
        </view>
      </view>

      <!-- 分类筛选（胶囊条，覆盖全部分类，选中态对齐参考页 fpill） -->
      <scroll-view scroll-x class="fbar-scroll" :show-scrollbar="false">
        <view class="fbar">
          <view
            v-for="chip in categoryChips"
            :key="chip.value"
            class="fpill"
            :class="{ on: activeCategory === chip.value }"
            @tap="onCategoryChange(chip.value)"
          >
            <text class="fpv">{{ chip.label }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- Banner 渐变卡（参考页同款深蓝系） -->
    <view class="banner">
      <view class="banner-icon">策</view>
      <view class="banner-info">
        <text class="banner-title">政策资讯，一站速览</text>
        <text class="banner-sub">聚焦低空产业政策 · 法规标准速递</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 当前筛选范围 -->
      <view class="ir">
        <text>共 <text class="irn">{{ filteredList.length }}</text> 项资讯</text>
        <text class="ir-suffix">{{ activeCategory ? categoryLabel(activeCategory) : '全部' }}</text>
      </view>

      <!-- 骨架屏 -->
      <view v-if="loading && list.length === 0" class="skl">
        <view v-for="i in 4" :key="'sk' + i" class="skc">
          <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w40"></view></view>
          <view class="sk-bd">
            <view class="sk-l w90"></view>
            <view class="sk-l w80"></view>
            <view class="sk-l w60"></view>
          </view>
        </view>
      </view>

      <!-- 错误态 -->
      <view v-else-if="errorMsg && list.length === 0" class="st">
        <u-empty description="加载失败，请检查网络">
          <view class="stb" @tap="fetchList(true)">重新加载</view>
        </u-empty>
      </view>

      <!-- 空态（无任何数据） -->
      <view v-else-if="!loading && list.length === 0" class="st">
        <u-empty description="暂无相关资讯">
          <text class="sth">试试调整分类</text>
          <view class="stb" @tap="clearFilters">清除筛选</view>
        </u-empty>
      </view>

      <!-- 空态（搜索/筛选无匹配） -->
      <view v-else-if="!loading && filteredList.length === 0" class="st">
        <u-empty description="暂无匹配资讯">
          <text class="sth">试试调整关键词或分类</text>
          <view class="stb" @tap="clearFilters">清除筛选</view>
        </u-empty>
      </view>

      <!-- 列表：纯文字卡片（左缘分类色条 + 分类/置顶 tag 为视觉锚点） -->
      <view v-else class="cl">
        <view
          v-for="item in filteredList"
          :key="item.id"
          class="card"
          :class="'cat-' + (item.category || 'other')"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="openDetail(item)"
        >
          <view class="c-bar"></view>
          <view class="c-top">
            <view class="c-badges">
              <text class="c-tag" :class="'tag-' + (item.category || 'other')">{{ categoryLabel(item.category) }}</text>
              <text v-if="item.is_pinned" class="c-st st-pin">置顶</text>
            </view>
            <text class="c-arrow">›</text>
          </view>
          <text class="ct">{{ item.title }}</text>
          <text v-if="item.summary" class="c-desc">{{ item.summary }}</text>
          <view class="c-meta">
            <text>发布于 {{ formatDate(item.created_at) }}</text>
            <text class="c-dot">·</text>
            <text class="c-src">{{ item.source || '来源未知' }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="list.length" class="lm">
        <text v-if="loadingMore">— 加载中 —</text>
        <text v-else-if="!hasMore">— 没有更多了 —</text>
        <text v-else>— 上拉加载更多 —</text>
      </view>
    </view>

    <!-- 有数据时的错误横幅 -->
    <view v-if="errorMsg && list.length > 0" class="error-banner">
      <text>{{ errorMsg }}</text>
      <text class="error-retry" @tap="fetchList(true)">重试</text>
    </view>

    <!-- 回到顶部 -->
    <view class="bt" :class="{ show: showBt }" aria-role="button" aria-label="回到顶部" @tap="scrollToTop"><text>↑</text></view>

    <!-- 详情底部弹层 -->
    <u-popup
      :show="sheetVisible"
      position="bottom"
      round
      @close="closeSheet"
    >
      <view v-if="selectedItem" class="sheet-body">
        <view class="sheet-handle"></view>
        <view class="sheet-head">
          <text class="sheet-title">{{ selectedItem.title }}</text>
          <view class="sheet-close" aria-role="button" aria-label="关闭" @tap="closeSheet">
            <text class="sheet-close-x">×</text>
          </view>
        </view>
        <scroll-view scroll-y class="sheet-scroll">
          <view class="detail-meta">
            <text class="c-tag" :class="'tag-' + (selectedItem.category || 'other')">
              {{ categoryLabel(selectedItem.category) }}
            </text>
            <text class="detail-text">{{ selectedItem.source || '' }}</text>
            <text class="detail-text">{{ selectedItem.author || '' }}</text>
            <text class="detail-text">{{ formatDate(selectedItem.created_at) }}</text>
          </view>
          <text class="detail-content">{{ selectedItem.content || selectedItem.summary || '暂无详细内容' }}</text>
        </scroll-view>
      </view>
    </u-popup>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

const CATEGORY_CHIPS = [
  { label: '全部', value: '' },
  { label: '低空经济', value: 'low_altitude_policy' },
  { label: '无人机法规', value: 'uav_regulation' },
  { label: '空域管理', value: 'airspace_management' },
  { label: '补贴政策', value: 'subsidy_policy' },
  { label: '行业标准', value: 'industry_standard' },
  { label: '无人机知识', value: 'drone_knowledge' },
]

const CATEGORY_MAP = {
  low_altitude_policy: '低空经济',
  uav_regulation: '无人机法规',
  airspace_management: '空域管理',
  subsidy_policy: '补贴政策',
  industry_standard: '行业标准',
  drone_knowledge: '无人机知识',
}

const PAGE_SIZE = 10

export default {
  data() {
    return {
      searchText: '',
      activeCategory: '',
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: PAGE_SIZE,
      hasMore: true,
      requestId: 0,
      sheetVisible: false,
      selectedItem: null,
      showBt: false,
      noMotion: false, // 减弱动效（无障碍）：Options API 直存，避免 setup() 混合触发微信端 props 解析异常
      categoryChips: CATEGORY_CHIPS,
    }
  },
  computed: {
    filteredList() {
      const q = this.searchText.trim().toLowerCase()
      if (!q) return this.list
      return this.list.filter(function (item) {
        const target = [item.title, item.summary, item.source, item.category]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
        return target.indexOf(q) !== -1
      })
    },
  },
  onLoad() {
    if (typeof this.checkMotion === 'function') this.checkMotion()
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    if (!this.loadingMore && this.hasMore) {
      this.loadMore()
    }
  },
  onPageScroll(e) {
    this.showBt = (e ? e.scrollTop : 0) > 400
  },
  methods: {
    // 减弱动效检测（无障碍）：逻辑同 utils/motion.js，Options API 直实现
    checkMotion() {
      try {
        const sys = uni.getSystemInfoSync()
        if (sys && sys.reduceMotion) this.noMotion = true
      } catch (e) { /* 忽略 */ }
      try {
        if (typeof uni.onAccessibilityInfoChange === 'function') {
          uni.onAccessibilityInfoChange((res) => { this.noMotion = !!(res && res.reduceMotion) })
        }
      } catch (e) { /* 旧基础库无此 API */ }
    },
    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
        this.requestId++
      } else {
        this.loadingMore = true
      }

      const reqId = this.requestId
      this.errorMsg = ''

      try {
        const params = {
          page: this.page,
          page_size: this.pageSize,
        }
        if (this.activeCategory) {
          params.category = this.activeCategory
        }

        const res = await request({
          url: '/api/v1/articles',
          data: params,
        })

        // Ignore stale responses when category changed during fetch
        if (reqId !== this.requestId) return

        const items = Array.isArray(res) ? res : []

        // P1 修复：优先用分页响应 total 判定 hasMore，
        // 避免末页恰好等于 pageSize 时（items.length === pageSize）误判还有更多
        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        const total = typeof res.total === 'number' ? res.total : null
        this.hasMore = total !== null ? this.list.length < total : items.length === this.pageSize
      } catch (e) {
        if (reqId !== this.requestId) return
        if (reset) {
          this.errorMsg = '网络异常，请稍后重试'
        } else {
          // On load-more error, just show a brief message, keep existing data
          uni.showToast({ title: '加载失败', icon: 'none', duration: 2000 })
          this.page--
        }
      } finally {
        if (reqId === this.requestId) {
          this.loading = false
          this.loadingMore = false
        }
      }
    },

    async loadMore() {
      this.page++
      await this.fetchList(false)
    },

    onSearchInput() {
      // 搜索即筛：v-model + filteredList 实时前端过滤（接口不支持 q/keyword）
    },

    onSearchConfirm() {
      // 确认搜索：收起键盘（过滤结果已实时呈现）
      if (typeof uni.hideKeyboard === 'function') {
        try { uni.hideKeyboard() } catch (e) { /* 忽略 */ }
      }
    },

    clearSearch() {
      this.searchText = ''
    },

    onCategoryChange(value) {
      this.activeCategory = value
      this.fetchList(true)
    },

    clearFilters() {
      this.searchText = ''
      this.activeCategory = ''
      this.fetchList(true)
    },

    openDetail(item) {
      this.selectedItem = item
      this.sheetVisible = true
    },

    closeSheet() {
      this.sheetVisible = false
      this.selectedItem = null
    },

    categoryLabel(cat) {
      return CATEGORY_MAP[cat] || cat || '其他'
    },

    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      if (isNaN(d.getTime())) return iso
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },

    scrollToTop() {
      uni.pageScrollTo({ scrollTop: 0, duration: 300 })
    },
  },
}
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

/* ===== 固定头部（吸顶，位于原生导航栏之下） ===== */
.sticky-head {
  position: sticky;
  top: 0;
  z-index: 40;
  background: #fff;
}

/* ===== 搜索框：白上白——纯白填充 + 灰描边 + 双层投影（参考页同款） ===== */
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
.b-sclr { color: #667085; font-size: 15px; padding: 10px; margin: -10px; } /* 热区扩大：视觉 × 外扩 10px */
.b-sep { width: 1px; height: 15px; background: #DDE1E6; margin: 0 9px 0 6px; flex: none; }
.b-sbtn { flex: none; color: #344054; font-size: 13px; line-height: 1; padding: 6px 2px 6px 0; }

/* ===== 分类筛选胶囊（参考页 fpill：白底灰描边淡投影，选中蓝） ===== */
.fbar-scroll { white-space: nowrap; }
.fbar {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 12px;
}
.fpill {
  flex: none;
  min-height: 40px; /* 触控目标：微信 44px 建议值附近 */
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.04), 0 3px 10px rgba(16, 24, 40, 0.04);
  color: #344054;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.fpill.on { border-color: #0A66C2; color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.fpill:active { transform: scale(.96); transition: transform .08s linear; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ===== Banner（参考页同款：深蓝渐变 + 图标 + 标题/副标题） ===== */
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

/* ===== 白色板块 / 信息行 ===== */
.section {
  margin-top: 0;
  padding: 0;
}
.ir {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 14px 4px;
  font-size: 12px;
  color: #667085;
}
.irn { color: #0A66C2; font-weight: 600; }
.ir-suffix { color: #667085; font-size: 12px; }

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
.sk-tag { width: 56px; height: 18px; border-radius: 4px; background: #EDF0F3; flex: none; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; }
.sk-l.w60 { width: 60%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }

/* ===== 错误 / 空态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.sth { font-size: 12px; color: #667085; display: block; margin-bottom: 16px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 列表卡片：白上白 + 左缘分类色条 + 分类/置顶 tag ===== */
.cl {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
}
.card {
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
.tap-scale { transform: scale(0.95); opacity: 0.9; }
.c-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: 10px 0 0 10px;
}
.c-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
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
.c-tag { color: #074D92; background: #EAF3FB; } /* 兜底色；实际按分类色由 tag-* 覆盖 */
.c-st.st-pin { color: #0A66C2; background: #EAF3FB; }
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
  min-width: 0;
}
.c-dot { color: #DDE1E6; flex: none; }
.c-src { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 45%; }
.c-arrow { color: #0A66C2; font-size: 18px; font-weight: 300; line-height: 1; flex: none; }

/* 分类配色（tag 浅底深字 = 左缘色条同源同色，对比度 ≥4.5:1，取自参考页 FIELD_TAG 体系） */
.tag-low_altitude_policy { color: #0d47a1; background: #E3EDF9; }
.cat-low_altitude_policy .c-bar { background: #0d47a1; }
.tag-uav_regulation { color: #b71c1c; background: #FBE9E9; }
.cat-uav_regulation .c-bar { background: #b71c1c; }
.tag-airspace_management { color: #1a237e; background: #E7E9F4; }
.cat-airspace_management .c-bar { background: #1a237e; }
.tag-subsidy_policy { color: #B54708; background: #FDEEE4; }
.cat-subsidy_policy .c-bar { background: #B54708; }
.tag-industry_standard { color: #004d40; background: #E4F2EF; }
.cat-industry_standard .c-bar { background: #004d40; }
.tag-drone_knowledge { color: #4a148c; background: #F0E9F7; }
.cat-drone_knowledge .c-bar { background: #4a148c; }
.tag-other { color: #344054; background: #EEF1F4; }
.cat-other .c-bar { background: #344054; }

/* ===== 加载更多 / 错误横幅 / 回到顶部 ===== */
.lm { text-align: center; padding: 12px; font-size: 12px; color: #667085; }
.error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 14px;
  margin: 8px 12px 0;
  font-size: 12px;
  color: #B42318;
  background: #FDECEC;
  border-radius: 8px;
}
.error-retry { color: #0A66C2; font-weight: 600; padding: 6px 4px; }
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

/* ===== 详情底部弹层 ===== */
.sheet-body { display: flex; flex-direction: column; }
.sheet-handle { width: 38px; height: 4px; margin: 9px auto 3px; background: #D0D5DD; border-radius: 2px; }
.sheet-head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 18px 13px;
  border-bottom: 1px solid #E5EAF1;
}
.sheet-title {
  flex: 1;
  font-size: 19px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
}
.sheet-close {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: background .2s ease;
}
.sheet-close:active { background: #F2F4F7; }
.sheet-close-x { font-size: 20px; color: #667085; line-height: 1; }
.sheet-scroll { max-height: calc(66vh - 54px); padding: 16px 18px 26px; }
.detail-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
}
.detail-text { font-size: 12px; color: #667085; }
.detail-content {
  display: block;
  margin-top: 17px;
  font-size: 14px;
  color: #344054;
  line-height: 1.8;
  white-space: pre-line;
}

/* ===================== 动效规范（对齐参考页） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许）
   时长：微反馈 150-200ms / 松手弹簧回位 .3s（ios-pop）/ 浮层 200-300ms；退场 = 进场 ×0.7 且必须存在
   曲线：ios-pop cubic-bezier(0.16,1,0.3,1) + ios-decel cubic-bezier(.32,.72,0,1)；其余 ease-out / ease-in / linear
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 列表入场：前 6 项每 20ms 依次淡入上移（首屏可见范围；80ms 起 + 100ms 错峰 + 220ms 动画 = 400ms ≤ 400ms） */
.card { animation: none; }
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
/* 分类色条与卡片同拍"点亮"（scaleY 顶部抽出，与 cardIn 同错峰，380ms 内收完） */
.card:nth-child(-n+6) .c-bar { animation: barIn .2s ease-out backwards; }
.card:nth-child(1) .c-bar { animation-delay: 80ms; }
.card:nth-child(2) .c-bar { animation-delay: 100ms; }
.card:nth-child(3) .c-bar { animation-delay: 120ms; }
.card:nth-child(4) .c-bar { animation-delay: 140ms; }
.card:nth-child(5) .c-bar { animation-delay: 160ms; }
.card:nth-child(6) .c-bar { animation-delay: 180ms; }
@keyframes barIn { from { opacity: 0; transform: scaleY(.3); } to { opacity: 1; transform: scaleY(1); } }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
/* 卡片按压（快进慢出）：按下 .1s linear 直接到位，松手 .35s ios-pop 弹簧回位 */
.card { transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.card.tap-scale { transition-duration: .1s; transition-timing-function: linear; }

/* Banner 内部微编排（替代整块 fadeUp）：图标 0ms → 标题 80ms → 装饰圆 120ms → 副文案 140ms，总 340ms ≤ 400ms */
.banner-icon { animation: iconIn .2s ease-out backwards; }
.banner-title { animation: fadeUp .2s ease-out 80ms backwards; }
.banner-sub { animation: fadeUp .2s ease-out 140ms backwards; }
.banner::after { animation: orbIn .3s ease-out 120ms backwards; }
@keyframes iconIn { from { opacity: 0; transform: scale(.92); } to { opacity: 1; transform: scale(1); } }
@keyframes orbIn { from { opacity: 0; transform: scale(1.1); } to { opacity: 1; transform: scale(1); } }
/* Banner 单次扫光（非循环装饰：100ms 起播 280ms 线性，380ms 内收完） */
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

/* 骨架呼吸（加载中环境光；循环动画 1.4s linear，一页仅此 1 处循环） */
.sk-tag, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 2) 交互反馈：可点元素按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位） */
.b-sclr:active { opacity: .6; }
.b-sbtn { transition: opacity .2s ease; }
.b-sbtn:active { opacity: .5; }
.stb { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease; }
.stb:active { transform: scale(.95); opacity: .85; transition: transform .08s linear; }
.error-retry { transition: opacity .2s ease; }
.error-retry:active { opacity: .6; }

/* ===== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===== */
.page.no-motion .card,
.page.no-motion .c-bar,
.page.no-motion .banner,
.page.no-motion .ir { animation: none; }
.page.no-motion .banner-icon,
.page.no-motion .banner-title,
.page.no-motion .banner-sub,
.page.no-motion .banner::before,
.page.no-motion .banner::after { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
.page.no-motion .tap-scale { transform: none !important; }
.page.no-motion .fpill:active,
.page.no-motion .stb:active,
.page.no-motion .bt:active { transform: none; }
</style>
