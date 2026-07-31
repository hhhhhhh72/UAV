<template>
  <view class="detail-page">
    <!-- Navbar -->
    <van-nav-bar
      :title="detail ? detail.title : '成果详情'"
      left-arrow
      @click-left="goBack"
    />

    <!-- ===== LOADING SKELETON ===== -->
    <view v-if="loading" class="skeleton-wrap">
      <view class="skel-hero"></view>
      <view class="skel-title">
        <view class="skel-line w80"></view>
        <view class="skel-line w40"></view>
      </view>
      <view class="skel-card">
        <view class="skel-line w60"></view>
        <view class="skel-line w100"></view>
        <view class="skel-line w100"></view>
        <view class="skel-line w40"></view>
      </view>
      <view class="skel-card">
        <view class="skel-line w40"></view>
        <view class="skel-line w100"></view>
        <view class="skel-line w80"></view>
      </view>
    </view>

    <!-- ===== ERROR ===== -->
    <view v-else-if="errorMsg" class="state-view">
      <view class="state-icon">&#9888;</view>
      <text class="state-text">加载失败，请检查网络连接</text>
      <text class="state-hint">请确认网络后重试</text>
      <view class="state-btn" @tap="fetchDetail">重新加载</view>
    </view>

    <!-- ===== EMPTY ===== -->
    <view v-else-if="!detail" class="state-view">
      <view class="state-icon">&#128269;</view>
      <text class="state-text">该成果已下架或不存在</text>
      <text class="state-hint">请返回列表浏览其他成果</text>
      <view class="state-btn" @tap="goBack">返回列表</view>
    </view>

    <!-- ===== NORMAL ===== -->
    <template v-else>
      <!-- Hero Cover -->
      <view class="hero" :style="{ background: heroBg }">
        <view class="hero-glow"></view>
        <text class="hero-icon">{{ heroIcon }}</text>
        <text class="hero-tag">{{ detail.field || detail.category || '科技成果' }}</text>
        <view class="hero-wave">
          <view class="hero-wave-inner"></view>
        </view>
      </view>

      <!-- Title + Badges (sticky) -->
      <view class="title-bar" :class="{ stuck: isStuck }">
        <text class="main-title">{{ detail.title }}</text>
        <view class="badges">
          <text v-if="detail.status === 'hot'" class="badge badge-hot">热门</text>
          <text v-if="isTransformed" class="badge badge-trans">已转化</text>
          <text v-if="stageLabel" class="badge badge-stage">{{ stageLabel }}</text>
        </view>
      </view>

      <!-- Stats Card -->
      <view class="card stats-card" :class="{ visible: animReady }">
        <view class="stat-item">
          <text class="stat-val">{{ displayViews }}</text>
          <text class="stat-label">浏览</text>
        </view>
        <view class="stat-item">
          <text class="stat-val">{{ displayFavs }}</text>
          <text class="stat-label">收藏</text>
        </view>
        <view class="stat-item">
          <text class="stat-val">{{ formatDate(detail.created_at || detail.date) }}</text>
          <text class="stat-label">发布日期</text>
        </view>
      </view>

      <!-- Description Card -->
      <view class="card" :class="{ visible: animReady }">
        <view class="card-head">
          <view class="card-dot"></view>
          <text class="card-title">成果描述</text>
        </view>
        <text class="card-body">{{ detail.description || detail.desc || '暂无描述' }}</text>
      </view>

      <!-- Info Card -->
      <view class="card" :class="{ visible: animReady }">
        <view class="card-head">
          <view class="card-dot"></view>
          <text class="card-title">基本信息</text>
        </view>
        <view class="info-table">
          <view v-if="detail.inventors" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">发明人</text>
            <text class="info-val">{{ detail.inventors }}</text>
          </view>
          <view v-if="detail.patent_number || detail.patent_no" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">专利号</text>
            <text class="info-val">{{ detail.patent_number || detail.patent_no }}</text>
          </view>
          <view v-if="detail.application_area || detail.area" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">应用领域</text>
            <text class="info-val">{{ detail.application_area || detail.area }}</text>
          </view>
          <view v-if="detail.org_name || detail.org" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">所属机构</text>
            <text class="info-val">{{ detail.org_name || detail.org }}</text>
          </view>
          <view v-if="detail.org_dept" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">实验室</text>
            <text class="info-val">{{ detail.org_dept }}</text>
          </view>
          <view v-if="detail.publish_date || detail.date || detail.created_at" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">发布日期</text>
            <text class="info-val">{{ formatDate(detail.publish_date || detail.date || detail.created_at) }}</text>
          </view>
          <view v-if="detail.stage || detail.status" class="info-row" :style="{ transitionDelay: (0.06 * ri++) + 's' }">
            <text class="info-key">成果阶段</text>
            <text class="info-val" :class="stageClass">{{ stageLabel || detail.stage || detail.status }}</text>
          </view>
        </view>
      </view>

      <!-- Attachments Card -->
      <view v-if="detail.attachments && detail.attachments.length" class="card" :class="{ visible: animReady }">
        <view class="card-head">
          <view class="card-dot"></view>
          <text class="card-title">附件资料</text>
        </view>
        <view class="att-list">
          <view
            v-for="(att, ai) in detail.attachments"
            :key="ai"
            class="att-item"
            :style="{ transitionDelay: (0.08 * ai) + 's' }"
          >
            <text class="att-icon">{{ attIcon(att.name || att.title) }}</text>
            <view class="att-info">
              <text class="att-name">{{ att.name || att.title || '附件 ' + (ai + 1) }}</text>
              <text v-if="att.size" class="att-size">{{ att.size }}</text>
            </view>
            <text class="att-dl">下载</text>
          </view>
        </view>
      </view>

      <!-- Spacer for bottom bar -->
      <view style="height:80px"></view>
    </template>

    <!-- ===== BOTTOM ACTION BAR ===== -->
    <view v-if="detail && !loading" class="bottom-bar" :class="{ hidden: barHidden }">
      <view class="btn-icon" :class="{ faved: isFav }" @tap="toggleFav">
        <van-icon :name="isFav ? 'like' : 'like-o'" size="18px" />
      </view>
      <view class="btn-primary" @tap="onContact">联系对接</view>
      <view class="btn-outline" @tap="toggleFav">{{ isFav ? '已收藏' : '收藏' }}</view>
      <view class="btn-icon share-btn" @tap="onShare">
        <van-icon name="share" size="18px" color="#666" />
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

/* ===== 字段映射 ===== */
var FIELD_ICONS = {
  '飞控系统': '\u2708', '遥感测绘': '\uD83C\uDF0D', '动力系统': '\u2699',
  'AI算法': '\uD83E\uDDE0', '载荷设备': '\uD83D\uDCF7', '集群协同': '\uD83D\uDCE1',
  '通信链路': '\uD83D\uDCE6', '标准规范': '\uD83D\uDCCB', '地面站': '\uD83D\uDCBB',
  '无人机': '\uD83D\uDEE9', '飞控': '\u2708', '载荷': '\uD83D\uDCF7',
  '软件': '\uD83D\uDCBB', '材料': '\u2699',
}

var FIELD_BG = {
  '飞控系统': 'linear-gradient(160deg,#0d47a1,#1565c0 30%,#1976d2 60%,#0d47a1)',
  '遥感测绘': 'linear-gradient(160deg,#1b5e20,#2e7d32 30%,#388e3c 60%,#1b5e20)',
  '动力系统': 'linear-gradient(160deg,#e65100,#ef6c00 50%,#f57c00)',
  'AI算法': 'linear-gradient(160deg,#4a148c,#6a1b9a 50%,#7b1fa2)',
  '载荷设备': 'linear-gradient(160deg,#b71c1c,#c62828 50%,#d32f2f)',
  '集群协同': 'linear-gradient(160deg,#004d40,#00695c 50%,#00796b)',
  '通信链路': 'linear-gradient(160deg,#1a237e,#283593 50%,#303f9f)',
  '标准规范': 'linear-gradient(160deg,#37474f,#455a64 50%,#546e7a)',
  '地面站': 'linear-gradient(160deg,#bf360c,#d84315 50%,#e64a19)',
  '无人机': 'linear-gradient(160deg,#0d47a1,#1976d2)',
  '飞控': 'linear-gradient(160deg,#1b5e20,#2e7d32)',
  '载荷': 'linear-gradient(160deg,#b71c1c,#c62828)',
  '软件': 'linear-gradient(160deg,#4a148c,#7b1fa2)',
  '材料': 'linear-gradient(160deg,#e65100,#ef6c00)',
}

/* ===== DEMO DATA (后端不可用时兜底) ===== */
var DEMO_DATA = [
  { id: '1', field: '飞控系统', title: '无人机智能自适应飞控系统 V3.0', org_name: '北航无人机研究所', org_dept: '飞行控制实验室', date: '2026-07-15', views: 2380, favs: 186, status: 'hot', stage: 'industrialization', description: '本项目针对复杂气象环境下无人机自主飞行控制的核心难题，提出了一种基于深度强化学习的自适应飞控架构。\n\n该系统集成了多传感器融合、在线参数优化与故障容错三大核心技术，可在强风、低能见度等极端条件下保持飞行稳定性。\n\n经实测验证，在6级阵风条件下姿态控制精度提升42%，系统响应延迟降低至8ms以内。目前已进入产业化中试阶段，与多家工业无人机企业达成合作意向。', inventors: '张明远、李晓峰、王建国、陈思雨', patent_number: 'CN202610012345.6', application_area: '工业巡检 | 应急救援 | 农业植保 | 物流配送', attachments: [{ name: '飞控系统技术白皮书 V3.0.pdf', size: '12.4 MB' }, { name: '风洞测试报告_2026Q2.pdf', size: '8.7 MB' }, { name: '自适应飞控算法源码仓库链接.txt', size: '0.1 MB' }] },
  { id: '2', field: '遥感测绘', title: '高精度无人机航测三维建模技术研究', org_name: '中科院遥感所', date: '2026-06-28', views: 1820, favs: 142, status: '', stage: 'pilot', description: '利用无人机搭载高精度激光雷达进行航测，实现厘米级三维建模。' },
  { id: '3', field: '动力系统', title: '工业级氢燃料电池动力系统', org_name: '清华大学', date: '2026-07-02', views: 3100, favs: 256, status: 'transformed', stage: 'industrialization', description: '自主研发的氢燃料电池系统，可为工业无人机提供超长续航。' },
  { id: '4', field: 'AI算法', title: '基于视觉Transformer的自主避障算法', org_name: '浙江大学', date: '2026-05-18', views: 1650, favs: 98, status: '', stage: 'laboratory' },
  { id: '5', field: '载荷设备', title: '轻量化多光谱成像载荷装置', org_name: '武汉大学', date: '2026-07-10', views: 1420, favs: 112, status: '', stage: 'pilot' },
  { id: '6', field: '集群协同', title: '无人机群协同搜索与救援调度系统', org_name: '国防科技大学', date: '2026-04-22', views: 2750, favs: 203, status: '', stage: 'laboratory' },
  { id: '7', field: '通信链路', title: '无人机超视距5G通信中继系统', org_name: '华为', date: '2026-07-20', views: 4200, favs: 315, status: 'hot', stage: 'industrialization' },
  { id: '8', field: '标准规范', title: '民用无人机适航审定技术标准', org_name: '民航科研院', date: '2026-03-15', views: 980, favs: 67, status: '', stage: 'listed' },
  { id: '9', field: '地面站', title: '便携式无人机地面控制站GCS-200', org_name: '成都纵横', date: '2026-06-10', views: 1560, favs: 89, status: '', stage: 'pilot' },
  { id: '10', field: '飞控系统', title: '基于MPC的倾转旋翼过渡段控制', org_name: '南京航空航天大学', date: '2026-07-08', views: 1230, favs: 74, status: '', stage: 'laboratory' },
  { id: '11', field: '载荷设备', title: '无人机载SAR雷达小型化成像系统', org_name: '中国电子科技集团', date: '2026-05-30', views: 2100, favs: 178, status: 'transformed', stage: 'listed' },
  { id: '12', field: 'AI算法', title: '无人机图像实时目标检测与跟踪平台', org_name: '大疆创新科技', date: '2026-07-01', views: 3800, favs: 290, status: 'hot', stage: 'industrialization' },
]

var STAGE_MAP = {
  laboratory: '实验室阶段', pilot: '中试阶段',
  industrialization: '已产业化', listed: '已上市',
}

export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      detail: null,
      isFav: false,
      isStuck: false,
      barHidden: false,
      animReady: false,
      displayViews: '0',
      displayFavs: '0',
      lastScrollTop: 0,
    }
  },

  computed: {
    heroIcon() {
      return FIELD_ICONS[this.detail.field || this.detail.category] || '\uD83D\uDE80'
    },
    heroBg() {
      return FIELD_BG[this.detail.field || this.detail.category] || FIELD_BG['飞控系统']
    },
    isTransformed() {
      var s = this.detail.stage || this.detail.status || ''
      return s === 'transformed' || s === 'industrialization' || s === '产业化' || s === '已转化' || s === 'listed'
    },
    stageLabel() {
      var s = this.detail.stage || this.detail.status || ''
      return STAGE_MAP[s] || (s && s !== 'hot' ? s : '')
    },
    stageClass() {
      var s = this.detail.stage || this.detail.status || ''
      var map = { laboratory: 'cl-stage', pilot: 'cl-warn', industrialization: 'cl-ok', hot: 'cl-hot' }
      return map[s] || ''
    },
  },

  onLoad(options) {
    this.id = options.id || ''
    this.fetchDetail()
  },

  onPageScroll(e) {
    var st = e.scrollTop
    this.isStuck = st > 80
    if (st > this.lastScrollTop && st > 200) {
      this.barHidden = true
    } else if (st < this.lastScrollTop || st < 60) {
      this.barHidden = false
    }
    this.lastScrollTop = st
  },

  methods: {
    async fetchDetail() {
      if (!this.id) { this.errorMsg = '缺少参数'; return }
      this.loading = true
      this.errorMsg = ''

      try {
        var res = await request({
          url: '/api/v1/achievements/' + encodeURIComponent(this.id),
        })
        var data = (res && res.data) || res || null
        this.detail = data
        if (data && data.title) {
          // #ifdef MP-WEIXIN
          uni.setNavigationBarTitle({ title: data.title })
          // #endif
        }
        this.$nextTick(function () {
          setTimeout(function () { this.animReady = true }.bind(this), 100)
          this.startCountUp()
        }.bind(this))
      } catch (e) {
        /* 后端不可用时使用演示数据 */
        var demo = DEMO_DATA.filter(function (x) { return x.id === this.id }.bind(this))[0]
        if (demo) {
          this.detail = demo
          this.$nextTick(function () {
            setTimeout(function () { this.animReady = true }.bind(this), 100)
            this.startCountUp()
          }.bind(this))
        } else {
          this.errorMsg = '网络异常，请稍后重试'
        }
      } finally {
        this.loading = false
      }
    },

    startCountUp() {
      var views = this.detail.views || this.detail.view_count || 0
      var favs = this.detail.favs || this.detail.fav_count || 0
      var dur = 800
      var self = this
      var frame = 0
      var totalFrames = Math.ceil(dur / 16)
      var vStep = Math.ceil(views / totalFrames)
      var fStep = Math.ceil(favs / totalFrames)
      var vCur = 0, fCur = 0
      var timer = setInterval(function () {
        frame++
        vCur = Math.min(vCur + vStep, views)
        fCur = Math.min(fCur + fStep, favs)
        self.displayViews = self.fmtNum(vCur)
        self.displayFavs = self.fmtNum(fCur)
        if (frame >= totalFrames) clearInterval(timer)
      }, 16)
    },

    /* ===== ACTIONS ===== */
    toggleFav() {
      this.isFav = !this.isFav
      if (this.isFav) {
        uni.showToast({ title: '已收藏', icon: 'success', duration: 1200 })
      }
    },

    onContact() {
      uni.showToast({ title: '联系对接 (功能待开放)', icon: 'none', duration: 1500 })
    },

    onShare() {
      // #ifdef MP-WEIXIN
      uni.showToast({ title: '点击右上角分享', icon: 'none', duration: 1500 })
      // #endif
    },

    goBack() {
      uni.navigateBack()
    },

    /* ===== UTILS ===== */
    attIcon(name) {
      if (!name) return '\uD83D\uDCC4'
      var ext = name.split('.').pop().toLowerCase()
      return ext === 'pdf' ? '\uD83D\uDCC4' : ext === 'txt' ? '\uD83D\uDCCB' : '\uD83D\uDCC1'
    },

    fmtNum(n) {
      if (!n) return '0'
      if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
      if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
      return String(n)
    },

    formatDate(iso) {
      if (!iso) return ''
      var d = new Date(iso)
      if (isNaN(d.getTime())) return String(iso).slice(0, 10)
      var m = d.getMonth() + 1
      var day = d.getDate()
      return d.getFullYear() + '-' + (m < 10 ? '0' : '') + m + '-' + (day < 10 ? '0' : '') + day
    },
  },
}
</script>

<style scoped>
/* ===== BASE ===== */
page { background: #f5f6f8; }
.detail-page { min-height: 100vh; background: #f5f6f8; padding-bottom: env(safe-area-inset-bottom); }

/* ===== SKELETON ===== */
.skeleton-wrap { padding: 0; }
.skel-hero { width: 100%; aspect-ratio: 16/9; background: #f0f1f3; animation: shimmer 1.5s infinite; }
.skel-title { padding: 20px 16px; background: #fff; }
.skel-card { margin: 0 16px 12px; padding: 16px; background: #fff; border-radius: 14px; }
.skel-line { height: 14px; background: #f0f1f3; border-radius: 4px; margin-bottom: 8px; animation: shimmer 1.5s infinite; }
.skel-line.w100 { width: 100%; } .skel-line.w80 { width: 80%; } .skel-line.w60 { width: 60%; } .skel-line.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== HERO ===== */
.hero {
  position: relative; width: 100%; aspect-ratio: 16/9;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  overflow: hidden;
}
.hero-glow {
  position: absolute; top: -20%; right: -15%;
  width: 200px; height: 200px; border-radius: 50%;
  background: radial-gradient(circle, rgba(255,255,255,.1) 0%, transparent 70%);
}
.hero-icon {
  font-size: 52px; position: relative; z-index: 1;
  animation: heroFloat 3s ease-in-out infinite;
}
@keyframes heroFloat { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-6px); } }
.hero-tag {
  position: relative; z-index: 1; margin-top: 10px; font-size: 11px;
  color: rgba(255,255,255,.85); background: rgba(255,255,255,.15);
  padding: 4px 14px; border-radius: 12px;
}
.hero-wave {
  position: absolute; bottom: -1px; left: 0; right: 0; z-index: 2;
  height: 24px; overflow: hidden;
}
.hero-wave-inner {
  width: 200%; height: 24px; margin-left: -50%;
  background: #fff; border-radius: 50% 50% 0 0;
}

/* ===== TITLE BAR (sticky) ===== */
.title-bar {
  padding: 20px 16px 12px; background: #fff;
  position: sticky; top: 0; z-index: 15;
  transition: box-shadow .25s;
}
.title-bar.stuck { box-shadow: 0 4px 16px rgba(0,0,0,.08); }
.main-title { font-size: 18px; font-weight: 700; color: #1a1a1a; line-height: 1.35; display: block; margin-bottom: 10px; }
.badges { display: flex; gap: 8px; flex-wrap: wrap; }
.badge { font-size: 11px; padding: 3px 10px; border-radius: 10px; font-weight: 500; color: #fff; }
.badge-hot { background: #ff3b30; }
.badge-trans { background: #34c759; }
.badge-stage { background: #e8f0fe; color: #1967d2; }

/* ===== CARDS ===== */
.card {
  margin: 0 16px 12px; padding: 16px;
  background: #fff; border-radius: 14px;
  box-shadow: 0 2px 12px rgba(0,0,0,.03);
  opacity: 0; transform: translateY(20px);
  transition: opacity .45s ease, transform .45s ease;
}
.card.visible { opacity: 1; transform: translateY(0); }

.card-head { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.card-dot { width: 4px; height: 18px; background: #1989fa; border-radius: 2px; flex-shrink: 0; }
.card-title { font-size: 15px; font-weight: 700; color: #1a1a1a; }
.card-body { font-size: 14px; color: #666; line-height: 1.75; display: block; white-space: pre-wrap; }

/* Stats card */
.stats-card { display: flex; padding: 16px 0; }
.stat-item { flex: 1; text-align: center; position: relative; }
.stat-item + .stat-item::before { content: ''; position: absolute; left: 0; top: 8px; bottom: 8px; width: .5px; background: #f0f0f0; }
.stat-val { font-size: 17px; font-weight: 700; color: #1a1a1a; display: block; }
.stat-label { font-size: 11px; color: #999; margin-top: 2px; display: block; }

/* Info table */
.info-table { display: flex; flex-direction: column; }
.info-row { display: flex; padding: 12px 0; border-bottom: .5px solid #f5f5f5; }
.info-row:last-child { border-bottom: none; }
.info-key { width: 70px; flex-shrink: 0; font-size: 13px; color: #999; }
.info-val { flex: 1; font-size: 14px; color: #333; word-break: break-all; }
.info-val.cl-stage { color: #1967d2; font-weight: 600; }
.info-val.cl-warn { color: #ff9f0a; font-weight: 600; }
.info-val.cl-ok { color: #34c759; font-weight: 600; }
.info-val.cl-hot { color: #ff3b30; font-weight: 600; }

/* Attachments */
.att-list { display: flex; flex-direction: column; gap: 8px; }
.att-item {
  display: flex; align-items: center; gap: 10px; padding: 12px;
  background: #f5f6f8; border-radius: 10px;
}
.att-item:active { background: #eef1f5; }
.att-icon { font-size: 18px; flex-shrink: 0; }
.att-info { flex: 1; min-width: 0; }
.att-name { font-size: 13px; color: #1a1a1a; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }
.att-size { font-size: 11px; color: #999; margin-top: 2px; display: block; }
.att-dl { font-size: 12px; color: #1989fa; flex-shrink: 0; font-weight: 500; }

/* ===== STATE VIEWS ===== */
.state-view { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 100px 20px; min-height: 400px; }
.state-icon { font-size: 48px; margin-bottom: 12px; opacity: .5; display: block; }
.state-text { font-size: 14px; color: #999; display: block; margin-bottom: 4px; }
.state-hint { font-size: 12px; color: #bbb; display: block; margin-bottom: 16px; }
.state-btn { padding: 8px 24px; border-radius: 22px; background: #1989fa; color: #fff; font-size: 13px; font-weight: 500; display: inline-block; }
.state-btn:active { opacity: .8; }

/* ===== BOTTOM BAR ===== */
.bottom-bar {
  position: fixed; bottom: 0; left: 0; right: 0;
  background: #fff; border-top: .5px solid #f0f0f0;
  display: flex; align-items: center; padding: 10px 16px; gap: 10px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  box-shadow: 0 -2px 12px rgba(0,0,0,.04);
  transition: transform .3s cubic-bezier(.4,0,.2,1);
  z-index: 50;
}
.bottom-bar.hidden { transform: translateY(100%); }
.btn-icon {
  width: 40px; height: 40px; border-radius: 50%;
  background: #f5f6f8; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.btn-icon:active { transform: scale(.88); }
.btn-icon.faved { color: #ff3b30; }
.btn-primary {
  flex: 1; height: 42px; border-radius: 21px; border: none;
  background: linear-gradient(135deg,#1565c0,#1976d2);
  color: #fff; font-size: 14px; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 14px rgba(25,118,210,.35);
}
.btn-primary:active { transform: scale(.97); }
.btn-outline {
  height: 42px; border-radius: 21px; border: 1.5px solid #1989fa;
  background: #fff; color: #1989fa; font-size: 14px; font-weight: 600;
  padding: 0 16px; display: flex; align-items: center; flex-shrink: 0;
}
.btn-outline:active { background: #e8f0fe; }
</style>
