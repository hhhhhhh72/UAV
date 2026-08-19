<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="行业报告" show-back :fixed="true" @back="goBack" />

    <!-- ① 白底头部：搜索 + 筛选 -->
    <view class="head-zone">
      <!-- 搜索框（白上白：双层投影浮起） -->
      <view class="sbar">
        <view class="b-search">
          <view class="b-search-ic"><view class="ic-ring" /><view class="ic-bar" /></view>
          <input
            class="b-sinp"
            v-model="searchText"
            placeholder="搜索报告名称"
            placeholder-class="b-ph"
            confirm-type="search"
            @confirm="onSearch"
          />
          <text v-if="searchText" class="b-sclr" @tap="clearSearch">×</text>
          <view class="b-sep" />
          <text class="b-sbtn" @tap="onSearch">搜索</text>
        </view>
      </view>

      <!-- 筛选胶囊：全部 / 白皮书 / 调研报告 / 年度报告 -->
      <view class="fbar">
        <view
          v-for="t in typeTabs"
          :key="t.value"
          class="fpill"
          :class="{ on: activeType === t.value }"
          hover-class="fpill-press"
          :hover-stay-time="100"
          @tap="onTabChange(t)"
        >
          <text class="fpv">{{ t.label }}</text>
        </view>
      </view>
    </view>

    <!-- ② 信息行 -->
    <view class="ir">
      <text>共 <text class="irn">{{ list.length }}</text> 份报告</text>
      <text class="ir-hint">{{ activeType || '全部类型' }}</text>
    </view>

    <!-- ③ 骨架屏：首次加载 -->
    <view v-if="loading && list.length === 0" class="skl">
      <view v-for="i in 3" :key="'sk' + i" class="skc">
        <view class="sk-row"><view class="sk-tag"></view><view class="sk-l w60"></view></view>
        <view class="sk-bd">
          <view class="sk-l w90"></view>
          <view class="sk-l w40"></view>
        </view>
      </view>
    </view>

    <!-- ④ 空 / 错误 -->
    <view v-else-if="!loading && list.length === 0" class="st">
      <u-empty :description="errorMsg || '暂无报告'">
        <view v-if="errorMsg" class="stb" @tap="fetchList(true)">重新加载</view>
      </u-empty>
    </view>

    <!-- ⑤ 报告列表 -->
    <view v-else class="cl">
      <view
        v-for="item in list"
        :key="item.id"
        class="card"
        hover-class="tap-scale"
        :hover-stay-time="100"
      >
        <view class="cell-title-row">
          <view class="report-icon"><text class="report-icon-text">文</text></view>
          <text class="cell-title">{{ item.title }}</text>
        </view>
        <view class="cell-meta">
          <text class="type-tag" :class="typeTagCls(item.report_type || item.type)">{{ typeLabel(item.report_type || item.type) }}</text>
          <text v-if="item.publish_date" class="meta-text">{{ item.publish_date }}</text>
        </view>
        <view class="cell-foot">
          <text class="cell-org">{{ item.author || item.period || '' }}</text>
          <view class="download-btn" hover-class="press-feedback" :hover-stay-time="100" @tap.stop="downloadReport(item)">
            <text class="download-btn-text">下载</text>
          </view>
        </view>
      </view>

      <!-- 已有数据时加载失败：错误横幅 + 重试（保留旧数据） -->
      <view v-if="errorMsg && list.length > 0" class="error-banner">
        <text>{{ errorMsg }}</text>
        <text class="error-retry" @tap="fetchList(true)">重试</text>
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
      loading: false,
      loadingMore: false,
      errorMsg: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      typeTabs: [
        { label: '全部', value: '' },
        { label: '白皮书', value: '白皮书' },
        { label: '调研报告', value: '调研报告' },
        { label: '年度报告', value: '年度报告' },
      ],
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
        if (this.activeType) params.type = this.activeType
        if (this.searchText) params.q = this.searchText
        params.page = this.page
        params.page_size = this.pageSize

        var res = await request({ url: '/api/v1/industry-reports', data: params })
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
        if (!reset) {
          // 加载更多失败回滚页码，避免跳过一页
          this.page--
          uni.showToast({ title: '加载失败，请重试', icon: 'none' })
        }
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    async loadMore() {
      this.page++
      await this.fetchList(false)
    },
    onSearch() {
      this.fetchList(true)
    },
    clearSearch() {
      this.searchText = ''
      this.fetchList(true)
    },
    onTabChange(t) {
      if (this.activeType === t.value) return
      this.activeType = t.value
      this.fetchList(true)
    },

    downloadReport(item) {
      var fileUrl = item.file_url || item.download_url || item.url
      if (fileUrl) {
        uni.downloadFile({
          url: fileUrl,
          success: function (res) {
            if (res.statusCode === 200) {
              uni.showToast({ title: '下载成功', icon: 'success' })
              uni.openDocument({
                filePath: res.tempFilePath,
                showMenu: true,
              })
            } else {
              uni.showToast({ title: '下载失败', icon: 'none' })
            }
          },
          fail: function () {
            uni.setClipboardData({
              data: fileUrl,
              success: function () {
                uni.showToast({ title: '链接已复制，请在浏览器打开', icon: 'none' })
              },
            })
          },
        })
      } else {
        uni.showToast({ title: '暂无下载链接', icon: 'none' })
      }
    },

    typeLabel(type) {
      var map = {
        '白皮书': '白皮书',
        '调研报告': '调研报告',
        '年度报告': '年度报告',
      }
      return map[type] || type || '其他'
    },

    typeTagCls(type) {
      var map = {
        '白皮书': 'tag--blue',
        '调研报告': 'tag--orange',
        '年度报告': 'tag--green',
      }
      return map[type] || 'tag--gray'
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
  padding-bottom: 40px;
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
.sk-l.w40 { width: 40%; }
.sk-l.w60 { width: 60%; }
.sk-l.w90 { width: 90%; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* ===== 空 / 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 60px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; }

/* ===== 报告卡片 ===== */
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

.cell-title-row { display: flex; align-items: center; gap: 8px; }
.report-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.report-icon-text { font-size: 14px; font-weight: 700; color: #0A66C2; }

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

.cell-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
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

.meta-text { font-size: 12px; color: #667085; }

.cell-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 2px;
  padding-top: 10px;
  border-top: 1px solid #F0F1F3;
}
.cell-org { font-size: 12px; color: #98A2B3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.download-btn {
  padding: 5px 16px;
  border-radius: 6px;
  background: #0A66C2;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.press-feedback { transform: scale(0.95); opacity: 0.85; }

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

/* 加载更多 */
.load-more { text-align: center; padding: 12px 0; }
.loading-inline { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; color: #667085; }
.no-more { color: #98A2B3; font-size: 12px; }

/* ===== 减弱动效（无障碍） ===== */
.page.no-motion .card,
.page.no-motion .ir { animation: none; }
.page.no-motion .sk-tag, .page.no-motion .sk-l { animation: none; }
</style>
