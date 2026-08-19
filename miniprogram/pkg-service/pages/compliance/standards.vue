<template>
  <view
    class="page"
    :class="{ 'no-motion': noMotion }"
    :style="{ paddingTop: (statusBarHeight + 44) + 'px' }"
  >
    <u-nav-bar title="团体标准库" show-back :fixed="true" @back="goBack" />

    <!-- 固定头部：搜索 + 分类 tabs（一体吸顶，避开 fixed 导航栏，参考页模式） -->
    <view class="sticky-head">
      <u-sticky :offset-top="statusBarHeight + 44">
        <view class="sbar">
          <u-search v-model="searchText" placeholder="搜索团体标准" @search="onSearch" />
        </view>
        <u-tabs :active="tabIndex" :titles="tabTitles" @change="onTabChange" />
      </u-sticky>
    </view>

    <!-- Banner 渐变卡（参考页同款深蓝系） -->
    <view class="banner">
      <view class="banner-icon">标</view>
      <view class="banner-info">
        <text class="banner-title">标准引领，规范先行</text>
        <text class="banner-sub">团体标准库 · 为行业提供规范参考</text>
      </view>
    </view>

    <!-- 白色板块：信息行 + 列表 -->
    <view class="section">
      <!-- 信息行：共 N 项 + 当前分类 -->
      <view class="ir">
        <text>共 <text class="irn">{{ filteredList.length }}</text> 项标准</text>
        <text class="ir-suffix">{{ activeCategory || '全部' }}</text>
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

      <!-- 空态 -->
      <view v-else-if="!loading && filteredList.length === 0" class="st">
        <u-empty description="暂无匹配标准">
          <text class="sth">试试调整关键词或分类</text>
          <view class="stb" @tap="clearFilters">清除筛选</view>
        </u-empty>
      </view>

      <!-- 列表：纯文字卡片（左缘分类色条 + 分类 tag + 下载按钮） -->
      <view v-else class="cl">
        <view
          v-for="item in filteredList"
          :key="item.id"
          class="card"
          :class="'cat-' + stdClass(item.category)"
          hover-class="tap-scale"
          hover-start-time="0"
          hover-stay-time="120"
          @tap="openStandard(item)"
        >
          <view class="c-bar"></view>
          <view class="c-top">
            <view class="c-badges">
              <text class="c-tag" :class="'tag-' + stdClass(item.category)">{{ item.category || '其他' }}</text>
            </view>
            <view class="c-dl-btn" aria-role="button" aria-label="下载标准" @tap.stop="downloadStandard(item)">
              <text class="dl-ic">↓</text>
            </view>
          </view>
          <text class="ct">{{ item.title || '--' }}</text>
          <text v-if="item.summary || item.scope" class="c-desc">{{ item.summary || item.scope }}</text>
          <view class="c-meta">
            <text>{{ (item.effective_date ? '实施于' : '发布于') + ' ' + formatDate(item.effective_date || item.created_at) }}</text>
            <text class="c-dot">·</text>
            <text v-if="item.standard_no" class="c-dl">{{ item.standard_no }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="list.length" class="lm">
        <text v-if="loading">— 加载中 —</text>
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
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      searchText: '',
      activeCategory: '',
      loading: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      statusBarHeight: 20,
      showBt: false,
      noMotion: false, // 减弱动效（无障碍）：Options API 直存，避免 setup() 混合触发微信端 props 解析异常
      categoryTabs: [
        { label: '全部', value: '' },
        { label: '国家标准', value: '国家标准' },
        { label: '行业标准', value: '行业标准' },
        { label: '团体标准', value: '团体标准' },
        { label: '企业标准', value: '企业标准' },
      ],
    }
  },
  computed: {
    tabTitles() {
      return this.categoryTabs.map(function (t) { return t.label })
    },
    tabIndex() {
      var idx = this.categoryTabs.findIndex(function (t) { return t.value === this.activeCategory }.bind(this))
      return idx >= 0 ? idx : 0
    },
    // 后端 ListStandards 仅支持 category 过滤，搜索词在前端本地过滤
    filteredList() {
      var kw = (this.searchText || '').trim()
      if (!kw) return this.list
      return this.list.filter(function (item) {
        return (item.title || '').indexOf(kw) !== -1 || (item.standard_no || '').indexOf(kw) !== -1
      })
    },
  },
  onLoad() {
    try {
      const sys = uni.getSystemInfoSync()
      this.statusBarHeight = sys.statusBarHeight || 20
    } catch (e) { /* 保持默认 */ }
    if (typeof this.checkMotion === 'function') this.checkMotion()
    this.fetchList(true)
  },
  onPullDownRefresh() {
    this.fetchList(true).then(function () {
      uni.stopPullDownRefresh()
    })
  },
  onReachBottom() {
    // P1 修复：pageSize=20 有分页但页面此前无触底加载，超过一页即不可达
    if (!this.loading && this.hasMore) {
      this.fetchList(false)
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
      } else {
        this.loading = true
      }
      this.errorMsg = ''

      try {
        var params = {
          page: this.page,
          page_size: this.pageSize,
        }
        if (this.activeCategory) params.category = this.activeCategory

        var res = await request({
          url: '/api/v1/compliance-standards',
          data: params,
        })
        var data = Array.isArray(res) ? res : (res && res.items) || []
        var items = Array.isArray(data) ? data : (data && data.items) || []
        var total = (data && data.total) != null ? data.total : items.length

        if (reset) {
          this.list = items
        } else {
          this.list = this.list.concat(items)
        }
        this.hasMore = this.list.length < total
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
        if (!reset) {
          // 加载更多失败回滚页码，避免跳过一页
          this.page--
          uni.showToast({ title: '加载失败，请重试', icon: 'none' })
        }
      } finally {
        this.loading = false
      }
    },
    onSearch() {
      this.fetchList(true)
    },
    onTabChange(index) {
      var tab = this.categoryTabs[index]
      this.activeCategory = tab ? tab.value : ''
      this.fetchList(true)
    },
    openStandard(item) {
      var content = item.summary || item.scope || item.standard_no || ''
      uni.showModal({
        title: item.title || '标准详情',
        content: content || '暂无详细内容',
        showCancel: false,
        confirmText: '知道了',
      })
    },
    downloadStandard(item) {
      if (item.file_url) {
        uni.downloadFile({
          url: item.file_url,
          success: function (res) {
            if (res.statusCode === 200) {
              uni.openDocument({
                filePath: res.tempFilePath,
                showMenu: true,
              })
            }
          },
          fail: function () {
            uni.showToast({ title: '下载失败', icon: 'none' })
          },
        })
      } else if (item.url) {
        uni.setClipboardData({
          data: item.url,
          success: function () {
            uni.showToast({ title: '链接已复制，请在浏览器打开下载', icon: 'none' })
          },
        })
      } else {
        uni.showToast({ title: '暂无下载资源', icon: 'none' })
      }
    },
    // 标准类型 → 配色类名（纯样式映射，用于左缘色条与分类 tag）
    stdClass(cat) {
      if (cat === '国家标准') return 'national'
      if (cat === '行业标准') return 'industry'
      if (cat === '团体标准') return 'group'
      if (cat === '企业标准') return 'enterprise'
      return 'other'
    },
    clearFilters() {
      this.searchText = ''
      this.activeCategory = ''
      this.fetchList(true)
    },
    goBack() {
      uni.navigateBack()
    },
    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
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

/* ===== 固定头部（u-sticky 吸顶，offset 避开 fixed 导航栏） ===== */
.sticky-head :deep(.u-sticky) { background: #fff; }

/* ===== 搜索（保留 u-search，视觉对齐参考页 b-search：圆角 + 双层投影） ===== */
.sticky-head :deep(.u-search) {
  background: #fff;
  padding: 12px 12px 6px;
}
.sticky-head :deep(.u-search-box) {
  height: 44px;
  padding: 0 11px;
  border: 1px solid #E4E7EC;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.06), 0 4px 12px rgba(16, 24, 40, 0.05);
  box-sizing: border-box;
}
.sticky-head :deep(.u-search-input) { font-size: 13px; color: #17212B; }
.sticky-head :deep(.u-search-ph) { color: #667085; }
.sticky-head :deep(.u-search-clear) { padding: 4px; }

/* ===== 分类 tabs（保留 u-tabs，视觉对齐参考页 fpill：胶囊 + 选中蓝） ===== */
.sticky-head :deep(.u-tabs) {
  background: #fff;
  padding: 4px 12px 12px;
}
.sticky-head :deep(.u-tabs-inner) {
  display: inline-flex;
  gap: 8px;
}
.sticky-head :deep(.u-tabs-item) {
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
  transition: transform .2s ease, border-color .2s ease, background .2s ease, color .2s ease;
}
.sticky-head :deep(.u-tabs-item--active) {
  border-color: #0A66C2;
  color: #0A66C2;
  background: #F4F8FC;
  font-weight: 600;
}
.sticky-head :deep(.u-tabs-item:active) { transform: scale(.96); transition: transform .08s linear; }
.sticky-head :deep(.u-tabs-line) { display: none; } /* 胶囊态不再需要底部指示线 */

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

/* ===== 列表卡片：白上白 + 左缘分类色条 + 分类 tag + 下载按钮 ===== */
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
.c-tag { color: #074D92; background: #EAF3FB; } /* 兜底色；实际按类型色由 tag-* 覆盖 */
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
.c-dl { color: #667085; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* 下载按钮 */
.c-dl-btn {
  width: 32px;
  height: 32px;
  flex: none;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), background .2s ease, opacity .15s ease;
}
.c-dl-btn:active { transform: scale(.9); background: #DCEBF9; transition: transform .08s linear; }
.dl-ic { font-size: 16px; font-weight: 700; line-height: 1; }

/* 标准类型配色（tag 浅底深字 = 左缘色条同源同色，对比度 ≥4.5:1，取自参考页 FIELD_TAG 体系） */
.tag-national { color: #0d47a1; background: #E3EDF9; }
.cat-national .c-bar { background: #0d47a1; }
.tag-industry { color: #B54708; background: #FDEEE4; }
.cat-industry .c-bar { background: #B54708; }
.tag-group { color: #004d40; background: #E4F2EF; }
.cat-group .c-bar { background: #004d40; }
.tag-enterprise { color: #4a148c; background: #F0E9F7; }
.cat-enterprise .c-bar { background: #4a148c; }
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
.page.no-motion .stb:active,
.page.no-motion .bt:active,
.page.no-motion .c-dl-btn:active { transform: none; }
</style>
