<template>
  <!--
  IMPECCABLE-DIRECTION v1 · seed=user-pinned · 行业报告（蓝皮书刊物）
  THESIS: 打开的是一份正式协会刊物，不是一条数据记录；拒绝卡片流+下划线筛选的通用列表。
  OWN-WORLD: 蓝皮书刊物世界：深蓝刊头版（渐变 #0A3A6B→#074D92）、类型圆章（字符竖排，纯 CSS 无图片）、出版信息条、宋体刊感标题、首字下沉摘要、附件卡、深蓝底栏；品牌令牌 #0A66C2/#1DD4A8/#F97316。
  STORY: 读者按类型寻刊 → 打开刊物 → 读完正文 → 下载 PDF。
  FIRST VIEWPORT: 深蓝刊头版（状态栏+返回+协会刊物+类型章+宋体刊名+报告期）→ 出版信息 → 摘要 → 正文 → 附件/底栏下载。
  FORM: 用户指定方向 B（蓝皮书刊物），真机对比拍板；数据经列表缓存传入（决策②，无公开按 ID 接口）；正文兼容富文本/纯文本/缺失三态。
  FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
  -->
  <view class="page">
    <template v-if="report">
      <!-- ===== 深蓝刊头版 ===== -->
      <view class="plate" :style="{ paddingTop: statusBarHeight + 'px' }">
        <view class="nav">
          <image :src="ICON_BACK" class="nav-back" hover-class="icon-press" mode="aspectFit" aria-role="button" aria-label="返回" @tap="goBack" />
          <text class="nav-title">协会刊物</text>
        </view>
        <view class="seal-row">
          <view class="seal" :class="{ long: typeChars.length >= 4 }">
            <text
              v-for="ch in typeChars"
              :key="ch"
              class="seal-char"
            >{{ ch }}</text>
          </view>
          <text class="seal-side">重庆市无人机产业协会 · 权威发布</text>
        </view>
        <text class="plate-title">{{ report.title || '-' }}</text>
        <view class="plate-meta">
          <text v-if="report.period">报告期 {{ report.period }}</text>
          <text v-if="report.period" class="plate-dot"></text>
          <text>{{ date }} 发布</text>
        </view>
        <view class="plate-rule"></view>
      </view>

      <!-- ===== 正文区 ===== -->
      <view class="body" :class="{ 'body-with-bottom': fileUrl }">
        <!-- 出版信息 -->
        <view class="pub">
          <text class="cap">出 版 信 息</text>
          <view class="pub-row">
            <text class="pub-l">发布机构</text>
            <text class="pub-v">重庆市无人机产业协会</text>
          </view>
          <view class="pub-row">
            <text class="pub-l">报告期</text>
            <text class="pub-v">{{ report.period || '-' }}</text>
          </view>
          <view class="pub-row">
            <text class="pub-l">作者</text>
            <text class="pub-v">{{ report.author || '-' }}</text>
          </view>
          <view class="pub-row">
            <text class="pub-l">发布时间</text>
            <text class="pub-v">{{ date }}</text>
          </view>
        </view>

        <!-- 摘要：首字下沉 -->
        <view v-if="report.summary" class="abs">
          <text class="cap">摘 要</text>
          <text class="abs-body">
            <text class="abs-drop">{{ summaryHead }}</text>{{ summaryRest }}
          </text>
        </view>

        <!-- 正文：富文本 / 纯文本 / 缺失 三态 -->
        <view class="doc">
          <rich-text
            v-if="prepared.kind === 'html'"
            :nodes="prepared.html"
          />
          <block v-else-if="prepared.kind === 'text'">
            <text
              v-for="(p, i) in prepared.paras"
              :key="i"
              class="doc-p"
            >{{ p }}</text>
          </block>
          <view v-else class="doc-empty">
            <text class="doc-empty-title">本报告暂无正文</text>
            <text class="doc-empty-desc">完整内容请下载报告文件查阅</text>
          </view>
        </view>

        <!-- 附件：整卡可点下载；下载中禁用，失败后文案引导重试 -->
        <view v-if="fileUrl" class="attach" :class="{ busy: downloading }" hover-class="attach-press" aria-role="button" aria-label="下载报告文件" @tap="download">
          <view class="attach-ico">
            <image :src="ICON_FILE" class="attach-ico-img" mode="aspectFit" />
          </view>
          <view class="attach-t">
            <text class="attach-name">{{ fileNameText }}</text>
            <text class="attach-sub">{{ downloading ? '正在下载…' : (downloadFail ? '下载失败，点击重试' : '点击下载报告文件') }}</text>
          </view>
          <view class="attach-btn" hover-class="btn-press">{{ downloading ? '下载中' : '下载' }}</view>
        </view>
        <view v-else class="no-file">
          <text>本报告暂无附件文件</text>
        </view>
      </view>

      <!-- ===== 深蓝底栏 ===== -->
      <view v-if="fileUrl" class="bottom">
        <text class="bottom-l">报告全文 · PDF</text>
        <view class="bottom-btn" :class="{ busy: downloading }" hover-class="btn-press" aria-role="button" aria-label="下载全文" @tap="download">
          <image :src="ICON_DOWNLOAD_NAVY" class="bottom-btn-ico" mode="aspectFit" />
          <text>{{ downloading ? '下载中…' : '下载全文' }}</text>
        </view>
      </view>
    </template>

    <!-- ===== 缓存缺失（深链/分享直达）：品牌化"刊物未能送达"版式 ===== -->
    <view v-else class="miss">
      <view class="plate" :style="{ paddingTop: statusBarHeight + 'px' }">
        <view class="nav">
          <image :src="ICON_BACK" class="nav-back" hover-class="icon-press" mode="aspectFit" aria-role="button" aria-label="返回" @tap="goBack" />
          <text class="nav-title">协会刊物</text>
        </view>
        <view class="seal-row">
          <view class="seal">
            <text class="seal-char">刊</text>
          </view>
          <text class="seal-side">重庆市无人机产业协会 · 权威发布</text>
        </view>
        <text class="plate-title">此刊物内容未能送达</text>
        <view class="plate-meta">
          <text>请在列表页重新打开</text>
        </view>
        <view class="plate-rule"></view>
      </view>
      <view class="miss-body">
        <text v-if="shareMeta && shareMeta.title" class="miss-quote">《{{ shareMeta.title }}》</text>
        <text class="miss-desc">{{ recovering ? '正在查找该报告…' : '未找到该报告，请从列表页重新打开' }}</text>
        <view class="miss-btn" hover-class="btn-press" aria-role="button" aria-label="返回列表" @tap="goBack">返回列表</view>
      </view>
    </view>
  </view>
</template>

<script>
// IMPECCABLE-DIRECTION v1 · seed=user-pinned · 行业报告详情（蓝皮书刊物）
// THESIS: 打开的是一份正式协会刊物。OWN-WORLD: 深蓝刊头版/类型圆章/出版信息条/宋体刊感/附件卡/深蓝底栏。
// STORY: 读刊物→读正文→下载 PDF。FORM: 用户指定方向 B。FINISH: unreviewed and undocumented is unfinished.
import { typeOf, formatDate, prepareContent, fileName, getReportCache, ICON_BACK, ICON_DOWNLOAD_NAVY, ICON_FILE } from './report-common.js'
import { request, BASE_URL } from '../../../utils/request'
import { MOCK_REPORTS } from '../../../utils/mockReports'

// 文件 URL 统一为可下载绝对地址：绝对/协议相对直接可用，其余补 BASE_URL（裸相对 downloadFile 必失败）
function resolveFileUrl(raw) {
  if (!raw) return ''
  if (/^(https?:\/\/|\/\/)/i.test(raw)) return raw
  return BASE_URL + raw
}

export default {
  data() {
    return {
      ICON_BACK,
      ICON_DOWNLOAD_NAVY,
      ICON_FILE,
      report: null,
      downloading: false,
      downloadFail: false,
      shareMeta: null,
      statusBarHeight: 24,
      recovering: false,
    }
  },
  computed: {
    typeDef() {
      return this.report ? typeOf(this.report.category) : typeOf('')
    },
    typeChars() {
      return this.typeDef.label.split('')
    },
    date() {
      return this.report ? formatDate(this.report.created_at) : '-'
    },
    prepared() {
      return this.report ? prepareContent(this.report.content) : { kind: 'empty' }
    },
    fileUrl() {
      if (!this.report) return ''
      return this.report.file_url || this.report.download_url || this.report.url || ''
    },
    // computed 命名带 Text 后缀，避免与上方导入的 fileName() 遮蔽
    fileNameText() {
      return fileName(this.report)
    },
    summaryHead() {
      var s = ((this.report && this.report.summary) || '').trim()
      return s.slice(0, 1)
    },
    summaryRest() {
      var s = ((this.report && this.report.summary) || '').trim()
      return s.slice(1)
    },
  },
  onLoad(options) {
    // 真机状态栏高（编译期 --status-bar-height 恒为 25px，高状态栏设备会压住刊头标题）
    this.statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 24
    this.report = getReportCache(options && options.id)
    // 深链/分享直达：缓存缺失时用分享参数重建刊物身份，miss 态仍可见刊名（平台已解码 query 值）
    if (!this.report && options && options.title) {
      this.shareMeta = { title: options.title, cat: options.cat || '' }
    }
    // 缓存缺失（深链/分享/清缓存）：尽力找回刊物，失败才落 miss 态
    if (!this.report) this.recoverReport(options && options.id)
  },
  onShareAppMessage() {
    var r = this.report
    if (!r) {
      return { title: '行业报告 · 协会刊物', path: '/pkg-service/pages/reports/list' }
    }
    var t = typeOf(r.category)
    return {
      title: (r.title ? r.title + ' · ' : '') + t.label,
      path: '/pkg-service/pages/reports/detail?id=' + r.id +
        '&title=' + encodeURIComponent(r.title || '') +
        '&cat=' + (r.category || ''),
    }
  },
  methods: {
    // 深链/分享直达时列表缓存缺失：先查演示数据（仅开发环境），再尽力按 id 从公开列表接口找回
    async recoverReport(id) {
      if (!id) return
      if (process.env.NODE_ENV === 'development' && MOCK_REPORTS && MOCK_REPORTS.length) {
        var m = MOCK_REPORTS.find(function (r) { return String(r.id) === String(id) })
        if (m) { this.report = m; return }
      }
      this.recovering = true
      try {
        var res = await request({ url: '/api/v1/industry-reports', data: { page: 1, page_size: 100 } })
        var items = Array.isArray(res) ? res : []
        var found = items.find(function (r) { return String(r.id) === String(id) })
        if (found) this.report = found
      } catch (e) { /* 查找失败保持 miss 态 */ }
      this.recovering = false
    },
    goBack() {
      // 深链/分享直达时可能无页面栈：兜底回首页，避免 navigateBack 静默失败
      if (getCurrentPages().length > 1) {
        uni.navigateBack()
      } else {
        uni.reLaunch({ url: '/pages/home/index' })
      }
    },
    // 下载 → 打开文档；失败按状态码分类提示并保留重试入口（按钮/整卡可再次点击重试）
    download() {
      if (this.downloading) return
      var raw = this.fileUrl
      if (!raw) {
        uni.showToast({ title: '暂无下载链接', icon: 'none' })
        return
      }
      var url = resolveFileUrl(raw)
      this.downloading = true
      this.downloadFail = false
      uni.downloadFile({
        url: url,
        success: (res) => {
          if (res.statusCode === 200) {
            uni.showToast({ title: '下载成功', icon: 'success' })
            uni.openDocument({
              filePath: res.tempFilePath,
              showMenu: true,
            })
          } else if (res.statusCode === 404 || res.statusCode === 410) {
            this.downloadFail = true
            uni.showToast({ title: '文件已失效', icon: 'none' })
          } else {
            this.downloadFail = true
            uni.showToast({ title: '下载失败，请重试', icon: 'none' })
          }
        },
        fail: () => {
          this.downloadFail = true
          uni.showToast({ title: '网络异常，请重试', icon: 'none' })
        },
        complete: () => {
          this.downloading = false
        },
      })
    },
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--color-bg);
  /* 次级文字对比度校准：全局 --color-text-secondary(#969799≈2.9:1) 不达 AA，
     本页局部覆盖为 #6b6e73(≈4.8:1)；全局令牌统一修复需团队确认后改 App.vue */
  --color-text-secondary: #6b6e73;
}

/* ===== 刊头版 ===== */
.plate {
  background: linear-gradient(160deg, #0a3a6b 0%, #074d92 100%);
  /* 顶部安全区由 JS 按真机状态栏高经 :style paddingTop 注入；CSS 顶部恒 0，勿回写 var(--status-bar-height) */
  /* 侧 40rpx：刊头/底栏专用缩进，比正文 32rpx 轨更宽（编辑层级），保留字面量 */
  padding: 0 40rpx 40rpx;
  position: relative;
  overflow: hidden;
}
.nav {
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}
.nav-back {
  position: absolute;
  left: 0;
  width: 44rpx;
  height: 44rpx;
  /* 触达热区：视觉 44rpx，盒模型含内边距 88rpx */
  padding: 22rpx;
  margin-left: -22rpx;
}
.nav-title {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 34rpx;
  font-weight: 600;
  color: #ffffff;
  letter-spacing: 4rpx;
}
.seal-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  /* 12/4 为 104rpx 圆章光学微调，勿入阶梯 */
  padding: 12rpx 0 4rpx;
}
.seal {
  /* 104rpx 圆：容纳最长 4 字类型竖排（4 × 20rpx × 1.25 = 100rpx），88rpx 会溢出圆外 */
  width: 104rpx;
  height: 104rpx;
  border: 3rpx solid rgba(255, 255, 255, 0.85);
  border-radius: 50%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  /* 开刊：圆章如盖章压入（scale+淡入，一次 0.3s） */
  animation: rpt-seal-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}
/* 内圈：双层圆章 */
.seal::before {
  content: '';
  position: absolute;
  top: 8rpx;
  left: 8rpx;
  right: 8rpx;
  bottom: 8rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.4);
  border-radius: 50%;
}
/* 类型字竖排：逐字换行渲染，兼容所有基础库（writing-mode 兼容性不可靠） */
.seal-char {
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 26rpx;
  line-height: 1.2;
  color: #ffffff;
  letter-spacing: 2rpx;
}
/* 4 字类型（调研报告/行业分析）：缩字+微调行距，竖排完整落在圆内 */
.seal.long .seal-char { font-size: 20rpx; line-height: 1.25; }
.seal-side {
  font-size: 20rpx;
  letter-spacing: 4rpx;
  color: rgba(255, 255, 255, 0.8);
  animation: rpt-fade-up 0.3s cubic-bezier(0.16, 1, 0.3, 1) 0.05s backwards;
}
.plate-title {
  display: block;
  margin-top: var(--space-sm);
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 42rpx;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.55;
  letter-spacing: 2rpx;
  /* 开刊：刊名轻抬浮现，紧随圆章 */
  animation: rpt-fade-up 0.32s cubic-bezier(0.16, 1, 0.3, 1) 0.08s backwards;
}
.plate-meta {
  margin-top: var(--space-sm);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.82);
  letter-spacing: 2rpx;
  animation: rpt-fade-up 0.3s cubic-bezier(0.16, 1, 0.3, 1) 0.14s backwards;
}
.plate-dot {
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  flex-shrink: 0;
}
.plate-rule {
  margin-top: var(--space-md);
  height: 2rpx;
  position: relative;
  /* 开刊：墨线从左画出（scaleX），菱形印随后浮现。用 ::after 承载线，避免印被缩放压扁 */
}
.plate-rule::before {
  content: '';
  position: absolute;
  left: 0;
  top: -7rpx;
  width: 12rpx;
  height: 12rpx;
  border: 2rpx solid #ffffff;
  transform: rotate(45deg);
  animation: rpt-fade-in 0.26s cubic-bezier(0.16, 1, 0.3, 1) 0.22s backwards;
}
.plate-rule::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  right: 0;
  height: 2rpx;
  background: rgba(255, 255, 255, 0.3);
  transform-origin: left center;
  transform: scaleX(0);
  animation: rpt-rule-draw 0.3s cubic-bezier(0.16, 1, 0.3, 1) 0.14s backwards;
}

/* ===== 正文区 ===== */
.body {
  position: relative;
  /* 上提 24rpx 叠入刊头 40rpx 底内边距：plate-rule 距刊头底 40rpx，始终可见（勿增大此值） */
  margin-top: -24rpx;
  border-radius: 32rpx 32rpx 0 0;
  background: var(--color-bg);
  padding: var(--space-lg) var(--space-lg) 0;
  overflow: hidden;
}

/* 出版信息（圆角标尺：卡片 20rpx，行间用负空间分组，无 hairline） */
.pub {
  background: #ffffff;
  border-radius: 20rpx;
  padding: var(--space-xs) var(--space-md);
  box-shadow: var(--shadow-sm);
}
.cap {
  display: block;
  font-size: 21rpx;
  letter-spacing: 4rpx;
  color: var(--color-text-secondary);
  /* 下 4rpx 补偿 letter-spacing 尾空白，勿入阶梯 */
  padding: var(--space-sm) 0 4rpx;
}
.pub-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 26rpx;
  padding: var(--space-sm) 0;
}
.pub-l {
  color: var(--color-text-secondary);
  flex-shrink: 0;
}
.pub-v {
  flex: 1;
  min-width: 0;
  color: var(--color-text);
  font-weight: 500;
  text-align: right;
  /* 标签值槽 40rpx：32/48 均越 4rpx 容差，保留字面量 */
  margin-left: 40rpx;
}

/* 摘要：首字下沉 */
.abs {
  margin-top: var(--space-lg);
  padding: 0 var(--space-md);
}
.abs-body {
  display: block;
  margin-top: var(--space-sm);
  font-size: 28rpx;
  color: #1a1a1a;
  line-height: 1.85;
  text-align: justify;
}
.abs-drop {
  float: left;
  font-family: Georgia, 'Songti SC', 'STSong', SimSun, serif;
  font-size: 76rpx;
  font-weight: 700;
  color: #074d92;
  line-height: 1;
  /* 10/12 为 76rpx 首字下沉的光学浮动微调，保留字面量 */
  padding: 10rpx 12rpx 0 0;
  /* 首字墨入：开刊序列后首字如印章压入（一次 0.3s；float 元素 animation 兼容小程序） */
  animation: rpt-drop-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) 0.42s backwards;
}
@keyframes rpt-drop-in {
  from { opacity: 0; transform: scale(0.85); }
  to { opacity: 1; transform: scale(1); }
}

/* 正文 */
.doc {
  margin-top: var(--space-lg);
  padding: 0 var(--space-md);
}
.doc-p {
  display: block;
  /* 与 rich-text 内联样式 report-common.js styleContentHtml() 保持一致（30rpx / 段距 24 / 墨黑 #1a1a1a） */
  font-size: 30rpx;
  color: #1a1a1a;
  line-height: 1.9;
  text-align: justify;
  margin-bottom: var(--space-md);
}
.doc-empty {
  /* 负值抵消 .doc 的 --space-md 水平内边距：空态卡边与 .pub/.attach 卡边齐平（32rpx 轨；镜像字面量 -24，勿用 var-in-calc） */
  margin: 0 -24rpx;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 56rpx 40rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: var(--shadow-sm);
}
.doc-empty-title {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--color-text);
}
.doc-empty-desc {
  margin-top: var(--space-sm);
  font-size: 24rpx;
  color: var(--color-text-secondary);
}

/* 附件（实线细边：虚线边框是"占位/未完成"信号，与正式下载动作不匹配；
   实线语言与 attach-ico 的 #cfe3f8 边框一致） */
.attach {
  margin: var(--space-lg) 0 0;
  border: 1rpx solid rgba(10, 102, 194, 0.28);
  background: #f5f9fe;
  border-radius: 20rpx;
  padding: var(--space-md) var(--space-md);
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.18s;
}
.attach-ico {
  width: 72rpx;
  height: 72rpx;
  border-radius: 16rpx;
  background: #ffffff;
  border: 1rpx solid #cfe3f8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.attach-ico-img {
  width: 34rpx;
  height: 34rpx;
}
.attach-t {
  flex: 1;
  min-width: 0;
}
.attach-name {
  display: block;
  font-size: 25rpx;
  font-weight: 600;
  color: #074d92;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.attach-sub {
  display: block;
  margin-top: var(--space-xs);
  font-size: 21rpx;
  /* #4a6a94：品牌蓝系深调，浅蓝底上 ≥4.5:1 */
  color: #4a6a94;
}
.attach-btn {
  flex-shrink: 0;
  height: 60rpx;
  line-height: 60rpx;
  padding: 0 var(--space-lg);
  border-radius: 38rpx;
  background: #074d92;
  color: #ffffff;
  font-size: 24rpx;
  font-weight: 600;
  box-shadow: 0 4rpx 12rpx rgba(7, 77, 146, 0.22);
}
.no-file {
  margin: var(--space-lg) 0 0;
  text-align: center;
  font-size: 24rpx;
  color: var(--color-text-secondary);
  padding: var(--space-md) 0;
}

/* 深蓝底栏 */
.bottom {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  /* 高 124rpx：与 .body-with-bottom 160 补偿配对，改前先看 body-with-bottom；侧 40rpx 与 .plate 同源，保留字面量 */
  height: 124rpx;
  background: #074d92;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 40rpx;
  padding-bottom: env(safe-area-inset-bottom);
}
.bottom-l {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.85);
  letter-spacing: 2rpx;
}
.bottom-btn {
  height: 76rpx;
  padding: 0 44rpx;
  border-radius: 38rpx;
  background: #ffffff;
  color: #074d92;
  font-size: 26rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}
.bottom-btn-ico {
  width: 28rpx;
  height: 28rpx;
}
.body-with-bottom {
  /* 160 = 底栏 124 + 余量 36；env() 补齐刘海屏安全区，否则末段内容被底栏遮挡 */
  padding-bottom: calc(160rpx + env(safe-area-inset-bottom));
}

/* 缓存缺失：品牌化"刊物未能送达"版式（复用刊头版视觉，最坏时刻也穿正装） */
.miss {
  min-height: 100vh;
  background: var(--color-bg);
}
.miss-body {
  /* 底 80rpx：末端收口，保留字面量 */
  padding: var(--space-xl) var(--space-xl) 80rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.miss-quote {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.6;
}
.miss-desc {
  margin-top: var(--space-sm);
  font-size: 24rpx;
  color: var(--color-text-secondary);
  line-height: 1.7;
}
.miss-btn {
  margin-top: var(--space-lg);
  height: 64rpx;
  line-height: 64rpx;
  padding: 0 56rpx;
  border-radius: 38rpx;
  background: #074d92;
  color: #ffffff;
  font-size: 26rpx;
  box-shadow: 0 6rpx 16rpx rgba(7, 77, 146, 0.25);
}
/* 按钮圆角标尺（模块规则）：统一 38rpx，介于卡片 20rpx 与全局 pill(50rpx) 之间，适配刊物方正美学；
   按全局约定补齐 box-shadow（深蓝底栏上的白色按钮除外，深色背景自带层级） */
.attach.busy, .bottom-btn.busy { opacity: 0.55; }

/* ===== 按压反馈（M3 静态档：仅 tap 触觉，无自动动效） ===== */
.attach-press { transform: scale(0.985); opacity: 0.92; }
.btn-press { transform: scale(0.96); opacity: 0.92; }
.icon-press { opacity: 0.6; }
.nav-back { transition: opacity 0.15s; }
.attach-btn, .bottom-btn, .miss-btn {
  transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.18s;
}

/* ===== 开刊动效（一次 0.34s 序列，仅页面打开时播放，无循环） ===== */
@keyframes rpt-seal-in {
  from { opacity: 0; transform: scale(0.8); }
  to { opacity: 1; transform: scale(1); }
}
@keyframes rpt-fade-up {
  from { opacity: 0; transform: translateY(10rpx); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes rpt-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes rpt-rule-draw {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}
</style>
