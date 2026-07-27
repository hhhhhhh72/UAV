<!--
  首页 (贴吧式任务大厅版) - 全新设计
  替换路径: miniprogram/pages/home/index.vue
  关键改进：
    - 顶部简化为 城市+搜索+消息 三件套
    - Hero Banner：蓝色大渐变 + 业务标题 + 实时统计
    - 保留 10 宫格快捷入口 + 同城公告 + 会员/合伙人双 CTA
    - 新增 贴吧式任务信息流 (Tieba Section)
      * 社区头部 + 关注按钮
      * 公众号二维码 Banner
      * 6 个分类 Tab (横向滚动)
      * 帖子卡片：头像 + 顶/独家/推荐 徽章 + 作者 + 手机尾号 + 电话按钮 + ...
              标题 + 地区标签 + 正文 + 全文链接 + 图片网格 + 字段标签
              底栏：浏览/时间 + 点赞/评论/分享
    - 浮动分享按钮 (绿色圆形)
    - 5 项底部 Tab (含中央发布 FAB)
-->
<template>
  <view class="home-new">
    <view class="page-container">
      <!-- ============== 顶部导航 ============== -->
      <view class="top-nav">
        <view class="city-btn" @tap="onCityTap">全国 <text class="city-arrow">▼</text></view>
        <view class="search-box" @tap="onSearch">
          <text class="s-ico">🔍</text>
          <text class="s-ph">{{ currentKeyword }}</text>
        </view>
        <view class="msg-btn">
          <text>💬</text>
          <view v-if="unreadCount" class="msg-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</view>
        </view>
      </view>

      <!-- ============== Hero Banner ============== -->
      <view class="hero-banner" @tap="onBannerClick">
        <text class="drone drone-l">🚁</text>
        <text class="drone drone-r">🚁</text>
        <text class="drone drone-l2">🚁</text>
        <text class="chart chart-r">📊</text>
        <text class="chart chart-r2">💰</text>
        <text class="hero-title">买无人机免费承接项目</text>
        <view class="hero-stats">
          <view class="hs-item">👁 浏览 <text class="hs-num">{{ stats.views }}</text></view>
          <view class="hs-item">📝 发帖 <text class="hs-num">{{ stats.posts }}</text></view>
          <view class="hs-item">🏢 商家 <text class="hs-num">{{ stats.merchants }}</text></view>
        </view>
        <view class="hero-cta">进入社区 ›</view>
      </view>

      <!-- ============== 10宫格 ============== -->
      <view class="icon-grid">
        <view v-for="item in quickEntries" :key="item.id" class="ig-item" @tap="onQuickTap(item)">
          <view class="ig-icon" :style="{ background: item.color }">{{ item.emoji }}</view>
          <text class="ig-label">{{ item.label }}</text>
        </view>
      </view>

      <!-- ============== 同城公告 ============== -->
      <view class="notice-strip" @tap="onNoticeTap">
        <text class="notice-tag">📢</text>
        <text class="notice-text"><text class="notice-kw">同城公告</text>&nbsp;&nbsp;{{ noticeText }}</text>
        <text class="notice-arrow">›</text>
      </view>

      <!-- ============== 双CTA ============== -->
      <view class="dual-cta">
        <view class="dual-card cta-member" @tap="onMemberTap">
          <view class="cta-l"><text class="cta-t1">加入会员</text><text class="cta-t2">加入会员更优惠</text></view>
          <view class="cta-ico">🎖️</view>
        </view>
        <view class="dual-card cta-partner" @tap="onPartnerTap">
          <view class="cta-l"><text class="cta-t1">同城合伙人</text><text class="cta-t2">加入同城合伙人</text></view>
          <view class="cta-ico">🤝</view>
        </view>
      </view>

      <!-- ============== 本地商家 ============== -->
      <view class="local-biz">
        <view class="local-head">
          <view>📍 本地商家</view>
          <text class="local-more" @tap="onLocalMore">全部 ›</text>
        </view>
        <scroll-view class="local-scroll" scroll-x>
          <view v-for="biz in localBusinesses" :key="biz.id" class="biz-card" @tap="onBizTap(biz)">
            <view class="biz-logo" :style="{ background: biz.color }">{{ biz.initial }}</view>
            <text class="biz-name">{{ biz.name }}</text>
          </view>
        </scroll-view>
      </view>

      <!-- ============== 贴吧式任务信息流 ============== -->
      <view class="tieba-section">
        <!-- 社区蓝色头部 -->
        <view class="tieba-header">
          <text class="tieba-name">小飞虾无人机圈子社区</text>
          <text class="tieba-sub">无人机产业链综合社区 · 共建共享</text>
          <view class="tieba-follow" @tap="onFollowTap">+ 关注</view>
        </view>

        <!-- 公众号QR Banner -->
        <view class="qr-banner">
          <view class="qr-l">
            <text class="qr-h">关注"小飞虾"公众号</text>
            <text class="qr-p">实时接收【私信】\n或【评论】消息提醒</text>
          </view>
          <view class="qr-img">
            <text class="qr-icon">📱</text>
            <text class="qr-text">长按识别</text>
          </view>
        </view>

        <!-- 贴吧Tab -->
        <scroll-view class="tieba-tabs" scroll-x>
          <view v-for="tab in tiebaTabs" :key="tab.key" :class="['tieba-tab', { active: activeTiebaTab === tab.key }]" @tap="switchTiebaTab(tab.key)">
            {{ tab.label }}
          </view>
        </scroll-view>

        <!-- 帖子列表 -->
        <view class="post-list">
          <view v-for="post in tiebaPosts" :key="post.id" class="post-card" @tap="onPostTap(post)">
            <!-- 顶部: 头像 + 作者 + 徽章 + 电话 -->
            <view class="post-top">
              <view class="post-avatar" :style="{ background: post.avatarBg }">{{ post.avatarInitial }}</view>
              <view class="post-author-info">
                <view class="post-author-row">
                  <view v-for="badge in post.badges" :key="badge.text" :class="['pb', badge.cls]">{{ badge.text }}</view>
                  <text class="post-author">{{ post.author }}</text>
                  <text class="post-phone">{{ post.phone }}</text>
                </view>
              </view>
              <view class="post-action-row">
                <view class="btn-tel" @tap.stop="onCallTap(post)">📞 电话</view>
                <text class="btn-more">⋯</text>
              </view>
            </view>

            <!-- 标题 + 地区 -->
            <text class="post-title">{{ post.title }}</text>
            <view v-if="post.location" class="post-loc">
              <text class="post-loc-k">📍 地区:</text>
              <text class="post-loc-v">{{ post.location }}</text>
            </view>

            <!-- 正文 -->
            <view v-if="post.body">
              <text class="post-body">{{ post.expanded ? post.body : (post.body.slice(0, 60) + (post.body.length > 60 ? '...' : '')) }}</text>
              <text v-if="post.body.length > 60" class="post-expand" @tap.stop="post.expanded = !post.expanded">
                {{ post.expanded ? '收起 ▴' : '全文 ▾' }}
              </text>
            </view>

            <!-- 图片网格 -->
            <view v-if="post.images && post.images.length" class="post-images">
              <view v-for="(img, idx) in post.images" :key="idx" class="post-img" :style="{ background: img.bg }">
                <text style="font-weight:700">{{ img.label }}</text>
              </view>
            </view>

            <!-- 字段标签 -->
            <view v-if="post.tags && post.tags.length" class="post-tags">
              <view v-for="tag in post.tags" :key="tag.label" class="post-tag">
                <text class="tag-lbl">{{ tag.label }}</text>
                <text class="tag-val">{{ tag.value }}</text>
              </view>
            </view>

            <!-- 底栏：浏览/时间 + 互动 -->
            <view class="post-bottom">
              <view class="post-stats">
                <text>🔥 <text class="ps-num">{{ post.views }}</text> 浏览 · 📅 {{ post.time }}</text>
              </view>
              <view class="post-actions">
                <view class="post-act" @tap.stop="onLikeTap(post)">
                  <text class="ico">{{ post.liked ? '❤️' : '👍' }}</text>
                  <text>{{ post.likes }}</text>
                </view>
                <view class="post-act" @tap.stop="onCommentTap(post)">
                  <text class="ico">💬</text>
                  <text>{{ post.comments }}</text>
                </view>
                <view class="post-act" @tap.stop="onShareTap(post)">
                  <text class="ico">↗️</text>
                </view>
              </view>
            </view>
          </view>

          <view v-if="tiebaPosts.length === 0 && !loading" class="post-empty">
            <text>暂无帖子</text>
          </view>

          <view v-if="loading" class="post-loading">加载中…</view>
          <view v-if="!hasMore && tiebaPosts.length > 0" class="post-no-more">— 已加载全部 —</view>
        </view>
      </view>

      <view style="height: 120rpx;"></view>
    </view>

    <!-- 浮动分享按钮 -->
    <view class="float-share-tieba" @tap="onFloatingShare">
      <text class="fs-ico">↗️</text>
      <text class="fs-lbl">分享</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import { safeNavigateTo, safeSwitchTab } from '../../utils/nav'

// ===== 顶部导航 =====
const unreadCount = ref(3)
const searchKeywords = ref(['吊运项目', '无人机租赁', '植保服务', 'CAAC考证'])
const currentKeyword = computed(() => searchKeywords.value[0])

const onCityTap = () => {
  uni.showActionSheet({
    itemList: ['全国', '重庆市', '北京市', '成都市', '其他'],
    success: () => {},
  })
}
const onSearch = () => { safeNavigateTo('/pages/search/index?keyword=' + encodeURIComponent(currentKeyword.value)) }

// ===== Hero Banner =====
const stats = ref({ views: '669万', posts: '848', merchants: '105' })
const onBannerClick = () => { safeNavigateTo('/pages/community/index') }

// ===== 10 宫格快捷入口 =====
const quickEntries = ref([
  { id: 'lifting', label: '吊运独享', emoji: '🚛', color: '#dbeafe', link: '/pages/community/posts?cat=lifting' },
  { id: 'trade', label: '买卖租赁', emoji: '🔄', color: '#fee2e2', link: '/pages/community/posts?cat=trade' },
  { id: 'training', label: '考证培训', emoji: '📚', color: '#fef3c7', link: '/pages/training/courses' },
  { id: 'plant', label: '植保运输', emoji: '🌾', color: '#d1fae5', link: '/pages/community/posts?cat=plant' },
  { id: 'clean', label: '清洗打药', emoji: '🧽', color: '#dbeafe', link: '/pages/community/posts?cat=clean' },
  { id: 'second', label: '二手交易', emoji: '🛒', color: '#fef3c7', link: '/pages/products/index' },
  { id: 'help', label: '免费互助', emoji: '🤝', color: '#fce7f3', link: '/pages/community/posts?cat=help' },
  { id: 'enter', label: '商家入驻', emoji: '🏢', color: '#fde68a', link: '/pages/enterprise/register' },
  { id: 'chat', label: '社区交流', emoji: '💬', color: '#ddd6fe', link: '/pages/community/index' },
  { id: 'jobs', label: '求职招聘', emoji: '👥', color: '#fee2e2', link: '/pages/jobs/list' },
])
const onQuickTap = (item) => { if (item.link) safeNavigateTo(item.link); else uni.showToast({ title: '功能开发中', icon: 'none' }) }

// ===== 同城公告 =====
const noticeText = ref('小飞虾独家发布了吊运独家信息')
const onNoticeTap = () => { safeNavigateTo('/pages/community/notices') }

// ===== 双 CTA =====
const onMemberTap = () => { safeNavigateTo('/pages/mine/membership') }
const onPartnerTap = () => { safeNavigateTo('/pages/mine/partner') }

// ===== 本地商家 =====
const localBusinesses = ref([
  { id: 'b1', name: '重庆无人机协会', initial: '协', color: 'linear-gradient(135deg,#1e3a8a,#1e40af)' },
  { id: 'b2', name: '小飞虾无人机', initial: '虾', color: 'linear-gradient(135deg,#fbbf24,#f59e0b)' },
  { id: 'b3', name: '鹰眼航空', initial: '鹰', color: 'linear-gradient(135deg,#1da1f2,#0ea5e9)' },
  { id: 'b4', name: '飞翔植保', initial: '飞', color: 'linear-gradient(135deg,#d1fae5,#a7f3d0)' },
  { id: 'b5', name: '海伯植保', initial: '海', color: 'linear-gradient(135deg,#8b5cf6,#7c3aed)' },
])
const onBizTap = (biz) => { safeNavigateTo(`/pages/services/detail?id=${biz.id}`) }
const onLocalMore = () => { safeSwitchTab('/pages/cases/index') }

// ===== 贴吧式信息流 =====
const tiebaTabs = ref([
  { key: 'newest', label: '最新信息' },
  { key: 'lifting', label: '吊运独享' },
  { key: 'trade', label: '买卖租赁' },
  { key: 'training', label: '考证培训' },
  { key: 'plant', label: '植保运输' },
  { key: 'jobs', label: '求职招聘' },
])
const activeTiebaTab = ref('newest')
const switchTiebaTab = (key) => { activeTiebaTab.value = key; fetchPosts(true) }

// Mock 数据（实际从后端拉取）
const MOCK_POSTS = [
  {
    id: 'p1',
    avatarBg: 'linear-gradient(135deg,#fbbf24,#f59e0b)',
    avatarInitial: '王',
    badges: [{ text: '顶', cls: 'pb-top' }, { text: '互帮互助', cls: 'pb-help' }],
    author: '王新刚', phone: '***3000',
    title: '山东省青岛市FC100, 3台。T1007台。万 物 可用。开无人机吊运租赁发票。',
    location: '山东省 青岛市',
    body: '专业承接各类型吊运业务，FC100/T100机型齐全，长年合作优惠价。山东周边可随时调度。已有5年吊运经验，熟悉各种复杂地形作业。诚信经营，价格透明。',
    images: [{ label: '吊运1', bg: 'linear-gradient(135deg,#1da1f2,#0ea5e9)' }, { label: '吊运2', bg: 'linear-gradient(135deg,#fbbf24,#f59e0b)' }],
    tags: [],
    views: '15689', time: '07-09 12:22', likes: 1, comments: 0,
    liked: false, expanded: false,
  },
  {
    id: 'p2',
    avatarBg: 'linear-gradient(135deg,#e64340,#dc2626)',
    avatarInitial: '独',
    badges: [{ text: '独家', cls: 'pb-exclusive' }],
    author: '吊运业务', phone: '小飞虾独家项目',
    title: '独家吊运业务 小飞虾独家项目',
    location: '全国（按需调度）',
    body: '',
    images: [{ label: '🌲', bg: 'linear-gradient(135deg,#22c55e,#16a34a)' }, { label: '🚁', bg: 'linear-gradient(135deg,#fbbf24,#f59e0b)' }],
    tags: [
      { label: '货物类型', value: '树/木头' },
      { label: '项目总量', value: '300吨' },
      { label: '单价', value: '面议' },
    ],
    views: '8924', time: '07-08 10:15', likes: 23, comments: 7,
    liked: false, expanded: false,
  },
  {
    id: 'p3',
    avatarBg: 'linear-gradient(135deg,#8b5cf6,#7c3aed)',
    avatarInitial: '培',
    badges: [{ text: '荐', cls: 'pb-recommend' }, { text: '考证培训', cls: 'pb-training' }],
    author: '飞翔学院', phone: '***8888',
    title: 'CAAC 无人机执照培训 · 7月开班',
    location: '重庆市 渝北区',
    body: 'CAAC 民航局无人机驾驶执照 (视距内/超视距/教员三级)，理论+实操，包过包就业。已培训3000+学员，就业率95%+。',
    images: [],
    tags: [
      { label: '课程', value: 'CAAC执照' },
      { label: '费用', value: '6800元' },
      { label: '开班', value: '7月15日' },
    ],
    views: '3412', time: '07-07 14:08', likes: 12, comments: 4,
    liked: false, expanded: false,
  },
]

const tiebaPosts = ref([])
const loading = ref(false)
const hasMore = ref(true)
const postPage = ref(1)

const fetchPosts = async (reset) => {
  if (reset) {
    postPage.value = 1
    hasMore.value = true
    tiebaPosts.value = []
  }
  loading.value = true
  try {
    // 实际接口: GET /api/v1/community/posts?tab=xxx&page=N
    const res = await request({
      url: '/api/v1/community/posts',
      data: { tab: activeTiebaTab.value, page: postPage.value, page_size: 10 },
    })
    const data = Array.isArray(res) ? res : (res?.data || res || {})
    const items = Array.isArray(data) ? data : (data.items || [])
    const total = (data.total != null) ? data.total : items.length

    // 后端无数据时使用 mock 让原型可演示
    const newItems = items.length > 0 ? items.map(adaptPost) : (postPage.value === 1 ? MOCK_POSTS.map(p => ({ ...p })) : [])

    if (reset) tiebaPosts.value = newItems
    else tiebaPosts.value = tiebaPosts.value.concat(newItems)

    hasMore.value = tiebaPosts.value.length < total || (items.length === 10 && items.length > 0)
  } catch (e) {
    // 失败时使用 mock
    if (reset) tiebaPosts.value = MOCK_POSTS.map(p => ({ ...p }))
  } finally {
    loading.value = false
  }
}
const adaptPost = (raw) => ({
  id: raw.id, avatarBg: raw.avatar_bg || 'linear-gradient(135deg,#fbbf24,#f59e0b)',
  avatarInitial: raw.avatar_initial || (raw.author?.charAt(0) || '用'),
  badges: raw.badges || [], author: raw.author || '匿名', phone: raw.phone || '',
  title: raw.title || '', location: raw.location || '', body: raw.body || raw.description || '',
  images: raw.images || [], tags: raw.tags || [],
  views: raw.views || 0, time: raw.time || raw.created_at || '', likes: raw.likes || 0, comments: raw.comments || 0,
  liked: false, expanded: false,
})

onShow(() => { if (tiebaPosts.value.length === 0) fetchPosts(true) })
onMounted(() => { fetchPosts(true) })
onReachBottom(() => { if (!loading.value && hasMore.value) { postPage.value++; fetchPosts(false) } })

// 贴吧交互
const onFollowTap = () => uni.showToast({ title: '关注成功', icon: 'success' })
const onCallTap = (post) => uni.makePhoneCall({ phoneNumber: post.phone.replace(/\*+/g, '1') || '400-000-0000' })
const onLikeTap = (post) => { post.liked = !post.liked; post.likes += post.liked ? 1 : -1 }
const onCommentTap = (post) => safeNavigateTo(`/pages/community/post-detail?id=${post.id}`)
const onShareTap = (post) => uni.showActionSheet({ itemList: ['分享到微信', '复制链接', '生成分享海报'] })
const onPostTap = (post) => safeNavigateTo(`/pages/community/post-detail?id=${post.id}`)
const onFloatingShare = () => uni.showActionSheet({ itemList: ['分享首页', '生成分享海报', '复制链接'] })
</script>

<style scoped>
.home-new{min-height:100vh;background:#f5f6f8;position:relative}

/* ===== 顶部导航 ===== */
.top-nav{
  display:flex;align-items:center;gap:10rpx;padding:18rpx 28rpx;
  background:#fff;border-bottom:1px solid #f0f1f3
}
.city-btn{font-size:28rpx;font-weight:600;color:#1a1a1a;display:flex;align-items:center;gap:2rpx}
.city-arrow{font-size:18rpx;color:#969799;margin-left:4rpx}
.search-box{
  flex:1;display:flex;align-items:center;gap:8rpx;
  background:#f5f6f8;border-radius:999px;padding:12rpx 24rpx
}
.s-ico{font-size:24rpx;opacity:.6}
.s-ph{flex:1;font-size:24rpx;color:#b4b6b8;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.msg-btn{
  width:60rpx;height:60rpx;display:flex;align-items:center;justify-content:center;
  font-size:34rpx;color:#969799;position:relative
}
.msg-badge{
  position:absolute;top:4rpx;right:0;background:#e64340;color:#fff;
  font-size:18rpx;font-weight:700;min-width:28rpx;height:28rpx;border-radius:14rpx;
  display:flex;align-items:center;justify-content:center;padding:0 6rpx
}

/* ===== Hero Banner ===== */
.hero-banner{
  position:relative;margin:20rpx 24rpx 0;border-radius:28rpx;overflow:hidden;
  height:280rpx;
  background:linear-gradient(135deg,#1e3a8a 0%,#1da1f2 50%,#0ea5e9 100%);
  color:#fff
}
.drone{position:absolute;font-size:100rpx;opacity:.4}
.drone-l{left:16rpx;top:60rpx;transform:rotate(-15deg)}
.drone-r{right:36rpx;top:16rpx;transform:rotate(8deg)}
.drone-l2{left:120rpx;bottom:28rpx;transform:rotate(5deg)}
.chart{position:absolute;font-size:44rpx;opacity:.7}
.chart-r{right:160rpx;top:36rpx}
.chart-r2{right:36rpx;bottom:36rpx}
.hero-title{
  position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);
  font-size:42rpx;font-weight:800;letter-spacing:2rpx;
  text-shadow:0 4rpx 16rpx rgba(0,0,0,.3);white-space:nowrap
}
.hero-stats{
  position:absolute;left:28rpx;bottom:24rpx;display:flex;gap:20rpx;
  font-size:20rpx;align-items:center
}
.hs-item{display:flex;align-items:center;gap:6rpx;color:rgba(255,255,255,.9);font-weight:600}
.hs-num{color:#fff;font-weight:800;font-size:26rpx}
.hero-cta{
  position:absolute;right:28rpx;bottom:24rpx;
  background:#fff;color:#1da1f2;font-size:22rpx;font-weight:700;
  padding:10rpx 28rpx;border-radius:999px
}

/* ===== 10宫格 ===== */
.icon-grid{
  display:grid;grid-template-columns:repeat(5,1fr);gap:12rpx;
  margin:20rpx 24rpx 0;padding:28rpx 16rpx;background:#fff;border-radius:28rpx
}
.ig-item{display:flex;flex-direction:column;align-items:center;padding:12rpx 8rpx}
.ig-icon{
  width:88rpx;height:88rpx;border-radius:24rpx;display:flex;
  align-items:center;justify-content:center;font-size:44rpx;margin-bottom:8rpx
}
.ig-label{font-size:22rpx;color:#1a1a1a;font-weight:500;text-align:center}

/* ===== 同城公告 ===== */
.notice-strip{
  display:flex;align-items:center;gap:16rpx;margin:20rpx 24rpx 0;
  padding:18rpx 28rpx;background:#fff;border-radius:24rpx;font-size:26rpx;color:#646566
}
.notice-tag{
  background:#ff6b35;color:#fff;font-size:22rpx;font-weight:700;
  padding:6rpx 16rpx;border-radius:12rpx;flex-shrink:0
}
.notice-text{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.notice-kw{color:#ff6b35;font-weight:700}
.notice-arrow{color:#969799;font-size:28rpx;flex-shrink:0}

/* ===== 双CTA ===== */
.dual-cta{display:grid;grid-template-columns:1fr 1fr;gap:20rpx;margin:20rpx 24rpx 0}
.dual-card{
  padding:28rpx 32rpx;border-radius:28rpx;display:flex;align-items:center;justify-content:space-between;
  position:relative;overflow:hidden
}
.cta-member{background:linear-gradient(135deg,#f97316,#ea580c);color:#fff}
.cta-partner{background:linear-gradient(135deg,#fbbf24,#f59e0b);color:#fff}
.cta-l .cta-t1{font-size:30rpx;font-weight:700;display:block}
.cta-l .cta-t2{font-size:22rpx;opacity:.85;margin-top:6rpx;display:block}
.cta-ico{font-size:56rpx;opacity:.95}

/* ===== 本地商家 ===== */
.local-biz{margin-top:20rpx}
.local-head{
  display:flex;justify-content:space-between;align-items:center;
  padding:32rpx 28rpx 16rpx;font-size:30rpx;font-weight:700
}
.local-more{font-weight:normal;font-size:24rpx;color:#969799}
.local-scroll{display:flex;gap:16rpx;padding:0 24rpx 20rpx;white-space:nowrap}
.biz-card{
  flex-shrink:0;width:260rpx;background:#fff;border-radius:24rpx;
  padding:16rpx 20rpx;display:flex;align-items:center;gap:16rpx
}
.biz-logo{
  width:68rpx;height:68rpx;border-radius:16rpx;display:flex;
  align-items:center;justify-content:center;font-weight:700;color:#fff;font-size:28rpx
}
.biz-name{
  flex:1;font-size:24rpx;font-weight:600;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:160rpx
}

/* ===== 贴吧式任务信息流 ===== */
.tieba-section{background:#fff;margin-top:28rpx;border-radius:28rpx 28rpx 0 0;overflow:hidden}
.tieba-header{
  position:relative;height:180rpx;
  background:linear-gradient(135deg,#0ea5e9 0%,#1da1f2 50%,#38bdf8 100%);
  color:#fff;padding:28rpx 32rpx
}
.tieba-header::before{
  content:"";position:absolute;left:-80rpx;top:-40rpx;width:280rpx;height:280rpx;
  background:radial-gradient(circle,rgba(255,255,255,.18),transparent);border-radius:50%
}
.tieba-name{font-size:32rpx;font-weight:800;position:relative}
.tieba-sub{font-size:22rpx;opacity:.85;margin-top:8rpx;display:block;position:relative}
.tieba-follow{
  position:absolute;right:28rpx;top:28rpx;
  background:#fff;color:#1da1f2;font-size:22rpx;font-weight:700;
  padding:10rpx 28rpx;border-radius:999px
}

.qr-banner{
  margin:20rpx 24rpx;background:linear-gradient(135deg,#38bdf8,#1da1f2);border-radius:24rpx;
  padding:24rpx 28rpx;color:#fff;display:flex;align-items:center;justify-content:space-between
}
.qr-h{font-size:26rpx;font-weight:700;line-height:1.4;display:block}
.qr-p{font-size:20rpx;opacity:.85;margin-top:4rpx;line-height:1.4;display:block}
.qr-img{
  width:108rpx;height:108rpx;background:#fff;border-radius:16rpx;
  display:flex;align-items:center;justify-content:center;flex-direction:column;
  font-size:18rpx;color:#666;padding:8rpx;text-align:center
}
.qr-icon{font-size:48rpx}.qr-text{font-size:18rpx;margin-top:4rpx}

.tieba-tabs{
  display:flex;background:#fff;padding:0 16rpx;border-bottom:1px solid #ebedf0;white-space:nowrap
}
.tieba-tab{
  flex-shrink:0;padding:24rpx 28rpx;font-size:26rpx;color:#646566;
  font-weight:500;position:relative
}
.tieba-tab.active{color:#1da1f2;font-weight:700}
.tieba-tab.active::after{
  content:"";position:absolute;bottom:0;left:50%;transform:translateX(-50%);
  width:36rpx;height:6rpx;border-radius:4rpx;background:#1da1f2
}

/* ===== 帖子卡片 ===== */
.post-list{padding:16rpx 0;background:#fafafa}
.post-card{
  margin:16rpx 20rpx;background:#fff;border-radius:24rpx;
  padding:28rpx 28rpx 24rpx;box-shadow:0 2rpx 8rpx rgba(0,0,0,.04)
}
.post-top{display:flex;align-items:center;gap:16rpx;margin-bottom:16rpx}
.post-avatar{
  width:64rpx;height:64rpx;border-radius:50%;
  display:flex;align-items:center;justify-content:center;
  color:#fff;font-size:26rpx;font-weight:700;flex-shrink:0
}
.post-author-info{flex:1;min-width:0}
.post-author-row{display:flex;align-items:center;gap:10rpx;flex-wrap:wrap}
.pb{font-size:18rpx;padding:2rpx 12rpx;border-radius:8rpx;font-weight:700}
.pb-top{background:#1da1f2;color:#fff}
.pb-help{background:#f97316;color:#fff}
.pb-exclusive{background:linear-gradient(90deg,#dc2626,#f97316);color:#fff}
.pb-recommend{background:#8b5cf6;color:#fff}
.pb-training{background:#22c55e;color:#fff}
.post-author{font-size:26rpx;font-weight:700;color:#1a1a1a}
.post-phone{font-size:22rpx;color:#969799;font-weight:500}

.post-action-row{display:flex;align-items:center;gap:12rpx}
.btn-tel{
  background:linear-gradient(135deg,#1da1f2,#0ea5e9);color:#fff;
  font-size:22rpx;font-weight:700;padding:8rpx 24rpx;border-radius:999px
}
.btn-more{color:#969799;font-size:36rpx;line-height:1}

.post-title{
  font-size:28rpx;font-weight:700;color:#1a1a1a;
  margin-bottom:12rpx;line-height:1.4;display:block
}
.post-loc{font-size:24rpx;color:#646566;margin-bottom:12rpx;display:block}
.post-loc-k{font-weight:700;color:#1a1a1a}
.post-loc-v{color:#646566;margin-left:6rpx}
.post-body{font-size:24rpx;color:#646566;line-height:1.6;margin-bottom:8rpx;display:block}
.post-expand{color:#1da1f2;font-size:24rpx;font-weight:600;margin-left:8rpx;display:inline-block}

.post-images{display:grid;grid-template-columns:1fr 1fr;gap:8rpx;margin:16rpx 0}
.post-img{
  height:180rpx;border-radius:16rpx;
  display:flex;align-items:center;justify-content:center;font-size:60rpx;color:#fff
}

.post-tags{display:flex;flex-wrap:wrap;gap:12rpx;margin:12rpx 0}
.post-tag{
  font-size:22rpx;padding:6rpx 20rpx;border-radius:999px;font-weight:600;
  background:#f5f6f8;color:#646566
}
.tag-lbl{color:#969799;font-weight:500;margin-right:6rpx}
.tag-val{color:#1a1a1a;font-weight:700}

.post-bottom{
  display:flex;justify-content:space-between;align-items:center;
  margin-top:20rpx;padding-top:16rpx;border-top:1px dashed #ebedf0
}
.post-stats{font-size:22rpx;color:#969799}
.ps-num{color:#1a1a1a;font-weight:700}
.post-actions{display:flex;gap:28rpx}
.post-act{
  font-size:22rpx;color:#646566;display:flex;align-items:center;gap:8rpx
}
.post-act .ico{font-size:28rpx}

.post-empty{text-align:center;padding:80rpx 0;color:#969799;font-size:26rpx}
.post-loading{text-align:center;padding:32rpx 0;color:#969799;font-size:24rpx}
.post-no-more{text-align:center;padding:24rpx 0;color:#c8c9cc;font-size:22rpx}

/* ===== 浮动分享 ===== */
.float-share-tieba{
  position:fixed;bottom:160rpx;right:50%;margin-right:-340rpx;
  width:96rpx;height:96rpx;border-radius:50%;
  background:linear-gradient(135deg,#22c55e,#16a34a);
  color:#fff;
  display:flex;align-items:center;justify-content:center;flex-direction:column;
  box-shadow:0 8rpx 24rpx rgba(34,197,94,.35);z-index:5;line-height:1
}
.fs-ico{font-size:32rpx}
.fs-lbl{font-size:18rpx;margin-top:2rpx;font-weight:700}
</style>