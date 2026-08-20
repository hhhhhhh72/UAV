<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="救援案例库" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="searchText"
            placeholder="搜索救援案例"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
          <view class="b-sep" />
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 筛选胶囊：全部 / 山火 / 洪水 / 地震 / 搜救 / 其他 -->
      <view class="fbar">
        <view
          v-for="(t, i) in typeTabs"
          :key="t.value"
          class="fpill"
          :class="{ on: activeType === t.value }"
          hover-class="fpill-press"
          :hover-stay-time="100"
          @tap="onTypeChange(i)"
        >
          <text class="fpv">{{ t.label }}</text>
        </view>
      </view>
    </view>

    <!-- ② 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 个案例</text>
      <text class="ir-hint">{{ activeType || '全部类型' }}</text>
    </view>

    <!-- ③ 骨架屏：首次加载 -->
    <view v-if="loading && list.length === 0" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w70"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w50"></view>
        </view>
      </view>
    </view>

    <!-- ④ 空 / 错误 -->
    <view v-else-if="!loading && list.length === 0" class="st">
      <u-empty :description="errorMsg || '暂无案例'">
        <view v-if="errorMsg" class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑤ 案例列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
      >
        <view class="c-top">
          <view class="c-icon" :style="eventIconStyle(item.event_type)"><text>{{ eventIcon(item.event_type) }}</text></view>
          <text class="c-title">{{ item.title || '未命名案例' }}</text>
        </view>

        <view class="c-meta">
          <text class="type-tag" :class="eventTagCls(item.event_type)">{{ eventTypeLabel(item.event_type) }}</text>
          <text v-if="item.date" class="c-text">{{ item.date }}</text>
        </view>

        <view v-if="item.location || item.drone_model" class="c-extra">
          <text v-if="item.location" class="c-text">{{ item.location }}</text>
          <text v-if="item.drone_model" class="c-text c-text--model">{{ item.drone_model }}</text>
        </view>

        <view v-if="item.result" class="c-result">
          <text class="c-result-label">处置结果</text>
          <text class="result-tag" :class="resultTagCls(item.result)">{{ resultLabel(item.result) }}</text>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="list.length > 0" class="load-more">
        <view v-if="loadingMore" class="loading-inline">
          <u-loading size="24rpx" />
          <text>加载更多...</text>
        </view>
        <text v-else-if="!hasMore" class="no-more">没有更多了</text>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../../utils/request'

export default {
  data() {
    return {
      noMotion: false,
      statusBarHeight: 20,
      searchText: '',
      activeType: '',
      typeTabs: [
        { label: '全部', value: '' },
        { label: '山火', value: '山火' },
        { label: '洪水', value: '洪水' },
        { label: '地震', value: '地震' },
        { label: '搜救', value: '搜救' },
        { label: '其他', value: '其他' },
      ],
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
    }
  },
  onLoad() {
    this.checkMotion()
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
  methods: {
    // 减弱动效（无障碍）+ 状态栏高度
    checkMotion() {
      try {
        const sys = uni.getSystemInfoSync()
        if (sys && sys.statusBarHeight) this.statusBarHeight = sys.statusBarHeight
        if (sys && sys.reduceMotion) this.noMotion = true
      } catch (e) {}
      try {
        if (typeof uni.onAccessibilityInfoChange === 'function') {
          uni.onAccessibilityInfoChange((res) => { this.noMotion = !!(res && res.reduceMotion) })
        }
      } catch (e) {}
    },

    async fetchList(reset) {
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.loading = true
      } else {
        this.loadingMore = true
      }
      this.errorMsg = ''

      try {
        var params = {}
        if (this.activeType) params.event_type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/rescue-cases', data: params })
        var data = Array.isArray(res) ? res : (res && res.data) || res || {}
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
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    loadMore() {
      this.page++
      this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    clearSearch() {
      this.searchText = ''
      this.fetchList(true)
    },
    onTypeChange(index) {
      var t = this.typeTabs[index]
      if (!t || this.activeType === t.value) return
      this.activeType = t.value
      this.fetchList(true)
    },

    /* 事件类型归一：后端可能下发英文（mountain_fire/flood/earthquake/search_rescue），统一映射为中文 */
    eventTypeLabel(type) {
      var map = {
        'mountain_fire': '山火',
        'flood': '洪水',
        'earthquake': '地震',
        'search_rescue': '搜救',
        '山火': '山火',
        '洪水': '洪水',
        '地震': '地震',
        '搜救': '搜救',
        '其他': '其他',
      }
      var key = String(type || '').toLowerCase()
      return map[key] || (type ? '其他' : '未知')
    },

    /* 事件类型字符图标（低饱和色块，非 emoji）；输入先归一，中英文均可 */
    eventIcon(type) {
      type = this.eventTypeLabel(type)
      var map = {
        '山火': '火',
        '洪水': '水',
        '地震': '地',
        '搜救': '救',
        '其他': '卫',
      }
      return map[type] || '卫'
    },
    eventIconStyle(type) {
      type = this.eventTypeLabel(type)
      var map = {
        '山火': { background: '#FFF0E6', color: '#E96012' },
        '洪水': { background: '#EAF3FB', color: '#0A66C2' },
        '地震': { background: '#F6F4FF', color: '#667085' },
        '搜救': { background: '#E9F7F0', color: '#168A55' },
      }
      return map[type] || { background: '#F4F6F8', color: '#667085' }
    },
    eventTagCls(type) {
      type = this.eventTypeLabel(type)
      var map = {
        '山火': 'tag--orange',
        '洪水': 'tag--blue',
        '地震': 'tag--gray',
        '搜救': 'tag--green',
        '其他': 'tag--gray',
      }
      return map[type] || 'tag--gray'
    },
    resultLabel(result) {
      var map = {
        '成功': '成功',
        '部分': '部分成功',
        '失败': '失败',
      }
      return map[result] || result || '未知'
    },
    resultTagCls(result) {
      var map = { '成功': 'tag--green', '部分': 'tag--orange', '失败': 'tag--red' }
      return map[result] || 'tag--gray'
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: env(safe-area-inset-bottom);
}

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

/* ===== 筛选胶囊 ===== */
.fbar { display: flex; gap: 8px; padding: 10px 12px 4px; background: #fff; }
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
.fpill-press { transform: scale(0.95); opacity: 0.85; }
.fpv { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ===== 信息行 ===== */
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
.sk-tag { width: 32px; height: 32px; border-radius: 8px; background: #EDF0F3; flex: none; animation: skPulse 1.4s linear infinite; }
.sk-bd { display: flex; flex-direction: column; gap: 8px; }
.sk-l { height: 12px; background: #EDF0F3; border-radius: 4px; animation: skPulse 1.4s linear infinite; }
.sk-l.w50 { width: 50%; }
.sk-l.w70 { width: 70%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 案例卡片 ===== */
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
.card:nth-child(-n+6) { animation: cardIn .22s ease-out backwards; }
.card:nth-child(1) { animation-delay: 80ms; }
.card:nth-child(2) { animation-delay: 100ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 140ms; }
.card:nth-child(5) { animation-delay: 160ms; }
.card:nth-child(6) { animation-delay: 180ms; }
@keyframes cardIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tap-scale { transform: scale(0.97); opacity: 0.9; }

.c-top {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}
.c-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}
.c-title {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.c-meta { display: flex; align-items: center; gap: 8px; }
.type-tag {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.tag--blue { background: #EAF3FB; color: #0A66C2; }
.tag--orange { background: #FFF4EC; color: #E96012; }
.tag--green { background: #E9F7F0; color: #0B6B41; }
.tag--gray { background: #EEF1F4; color: #5D6B82; }
.tag--red { background: #FDECEC; color: #B42318; }

.c-text { font-size: 12px; color: #667085; }
.c-text--model { color: #98A2B3; }

.c-extra { display: flex; align-items: center; gap: 12px; }

.c-result {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #F7F9FC;
  border-radius: 6px;
  padding: 7px 10px;
}
.c-result-label { font-size: 12px; color: #98A2B3; }
.result-tag {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

/* 加载更多 */
.load-more { text-align: center; padding: 10px 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; color: #667085; }
.no-more { font-size: 12px; color: #98A2B3; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
