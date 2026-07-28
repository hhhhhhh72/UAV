<template>
  <view class="dd-page">
    <view class="dd-header" :style="{ paddingTop: (statusBarH || 24) + 'px' }">
      <text class="back-btn" @tap="goBack">‹</text>
      <text class="title">【{{ post.tag || '互助' }}】{{ (post.title || '山东省需求').slice(0, 16) }}</text>
      <text class="more-btn">⋯</text>
    </view>

    <view v-if="post" class="dd-content">
      <!-- 用户信息 -->
      <view class="user-row">
        <image :src="post.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="user-ava" />
        <view class="user-meta">
          <view class="user-line1">
            <text class="user-tag">{{ post.tag || '互助' }}</text>
            <text class="user-name">{{ post.userName }}</text>
          </view>
          <view class="user-line2">
            <text class="user-time">{{ post.time }} </text>
            <text class="user-from">来自 {{ post.from || '小飞虾无人机圈子社区平台' }}</text>
          </view>
        </view>
      </view>

      <!-- 浏览/点赞统计 -->
      <view class="stat-row">
        <text class="stat-num">{{ post.views || 0 }}</text>
        <text> 浏览、</text>
        <text class="stat-num">{{ post.likes || 0 }}</text>
        <text> 人点赞</text>
      </view>

      <!-- 地区 -->
      <view class="loc-row">
        <text class="loc-label">地区：</text>
        <text class="loc-text">{{ post.location }}</text>
      </view>

      <!-- 描述 -->
      <view class="desc-block">
        <text class="desc-text" :class="{ collapse: !descExpanded }">{{ post.desc }}</text>
        <text class="desc-toggle" @tap="descExpanded = !descExpanded">{{ descExpanded ? '收起' : '全文' }}</text>
      </view>

      <!-- 大图 -->
      <view class="photo-block" v-if="post.photos && post.photos.length">
        <image v-for="(p, i) in post.photos" :key="i" :src="p" mode="aspectFill" class="photo" @tap="preview(i)" />
      </view>

      <view v-else class="photo-block photo-empty">
        <text class="photo-text">（暂无图片）</text>
      </view>

      <!-- 悬浮右侧操作按钮 -->
      <view class="float-actions">
        <view class="fa-item" @tap="generatePoster">
          <text class="fa-icon">⊞</text>
          <text class="fa-label">生成海报</text>
        </view>
        <view class="fa-item fa-share" @tap="share">
          <text class="fa-icon">➤</text>
          <text class="fa-label">分享</text>
        </view>
      </view>

      <!-- 底部广告条 -->
      <view v-if="showAd" class="ad-bar">
        <text class="ad-text">📞 点击拨号，已被人浏览过 26 人</text>
        <text class="ad-close" @tap="showAd = false">✕</text>
      </view>

      <!-- 点赞区 -->
      <view class="like-row">
        <text class="like-btn">👍 {{ post.likes || 0 }}</text>
      </view>

      <!-- 联系人 -->
      <view class="contact-row">
        <text class="cl-label">联系人：</text>
        <text class="cl-name">{{ post.userName }}</text>
      </view>
      <view class="contact-row">
        <text class="cl-tip">联系电话请告知我需要的无人机型号，飞手要求，地点等具体要求详细说明。</text>
      </view>

      <!-- 风险提示 -->
      <view class="warn-row">
        <text class="warn-text">如遇无效、虚假、诈骗信息，<text class="warn-link">请立即举报</text></text>
        <text class="warn-sub">为了您的资金安全，请见面交易，切勿提前支付任何费用</text>
      </view>

      <!-- 全部评论 -->
      <view class="comment-block">
        <view class="comment-head">
          <text class="ch-title">全部评论</text>
        </view>
        <view v-if="comments.length" class="comment-list">
          <view v-for="(c, i) in comments" :key="i" class="comment-item">
            <image :src="c.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="ci-ava" />
            <view class="ci-meta"><text class="ci-name">{{ c.name }}</text><text class="ci-text">{{ c.text }}</text></view>
          </view>
        </view>
        <view v-else class="comment-empty">还没有评论...</view>
      </view>

      <!-- 帖子发布者其他发布 -->
      <view class="other-block">
        <view class="ob-head">
          <image :src="post.avatar || '/static/home-bg.jpg'" mode="aspectFill" class="ob-ava" />
          <text class="ob-name">{{ post.userName }}</text>
        </view>
        <view class="ob-stat">
          <text>发布 <text class="ob-num">{{ postStats.posts || 0 }}</text> 条</text>
        </view>
      </view>
    </view>

    <!-- 右下浮动按钮 -->
    <view class="fab">
      <view class="fab-item" @tap="report"><text class="fab-ico">⚐</text><text class="lab">举报</text></view>
      <view class="fab-item fab-on" @tap="share"><text class="fab-ico">➤</text><text class="lab">分享</text></view>
      <view class="fab-item" @tap="goComment"><text class="fab-ico">💬</text><text class="lab">评论</text></view>
    </view>

    <!-- 底部拨打电话按钮 -->
    <view class="sticky-call">
      <view class="sc-btn" @tap="callPhone">
        <text class="sc-ico">📞</text>
        <text>拨打电话</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { safeNavigateTo } from '../../utils/nav'

const statusBarH = ref(24)
const post = ref(null)
const comments = ref([])
const showAd = ref(true)
const descExpanded = ref(false)
const postStats = ref({ posts: 1 })

const postId = computed(() => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1]
  return page?.options?.id || ''
})

const goBack = () => uni.navigateBack()
const expandDesc = () => { descExpanded.value = true }
const preview = (i) => uni.previewImage({ urls: post.value.photos, current: i })
const callPhone = () => post.value?.phone && uni.makePhoneCall({ phoneNumber: post.value.phone })
const share = () => uni.showShareMenu()
const report = () => uni.showToast({ title: '举报已提交', icon: 'none' })
const goComment = () => uni.showToast({ title: '评论功能开发中', icon: 'none' })
const generatePoster = () => uni.showToast({ title: '海报生成中', icon: 'none' })

onMounted(() => {
  statusBarH.value = (uni.getSystemInfoSync().statusBarHeight || 24) + 6
  // 加载数据（实际应从接口拉）
  post.value = {
    tag: '互助',
    userName: '王新刚***3000',
    time: '07-09 12:22',
    from: '小飞虾无人机圈子社区平台',
    title: '山东省青岛市吊运项目',
    views: 15999,
    likes: 1,
    location: '山东省青岛市',
    desc: '山东省青岛市FC100，3台，T1007台。万物可吊。能开无人机吊运租赁发票。客户地址在哪里，我们就去哪里作业。我们所有的设备技术人员都配备了专业操作平台，万物可吊。',
    photos: ['/static/home-bg.jpg', '/static/home-bg.jpg'],
    phone: '4001234567',
    avatar: '/static/home-bg.jpg'
  }
  // 兜底评论
  if (!comments.value.length) {
    comments.value = []
  }
})
</script>

<style scoped>
.dd-page { min-height: 100vh; background: #fff; padding-bottom: 80px; }

/* Header */
.dd-header {
  display: flex; align-items: center; gap: 12px; padding: 0 14px 12px;
  background: #fff; border-bottom: 1px solid #f0f0f0;
}
.back-btn { font-size: 28px; color: #333; line-height: 1; }
.title { flex: 1; font-size: 15px; font-weight: 500; color: #333; }
.more-btn { font-size: 22px; color: #666; padding: 0 6px; }

/* User */
.user-row { display: flex; align-items: center; gap: 12px; padding: 14px; background: #fff; }
.user-ava { width: 48px; height: 48px; border-radius: 50%; background: #e8f2fc; }
.user-meta { flex: 1; }
.user-line1 { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.user-tag { background: #1989fa; color: #fff; font-size: 11px; padding: 2px 8px; border-radius: 10px; }
.user-name { font-size: 15px; font-weight: 600; color: #5a3ed1; }
.user-line2 { font-size: 11px; color: #999; }
.user-time { color: #666; }
.user-from { color: #1989fa; }

/* Stats */
.stat-row { padding: 8px 14px 14px; font-size: 13px; color: #666; background: #fff; }
.stat-num { font-weight: 600; color: #ff6b35; }

/* Location */
.loc-row { display: flex; padding: 12px 14px; background: #fff; border-top: 1px solid #f0f0f0; border-bottom: 1px solid #f0f0f0; font-size: 14px; }
.loc-label { color: #999; min-width: 60px; }
.loc-text { color: #333; flex: 1; }

/* Description */
.desc-block { padding: 14px; background: #fff; font-size: 14px; line-height: 1.7; color: #333; }
.desc-text { display: inline; }
.desc-text.collapse { display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
.desc-toggle { color: #1989fa; font-size: 13px; margin-left: 6px; }

/* Photo */
.photo-block { position: relative; padding: 0 14px 14px; background: #fff; }
.photo { width: 100%; height: 220px; border-radius: 8px; display: block; margin-bottom: 8px; background: #f0f0f0; }
.photo-empty { height: 80px; display: flex; align-items: center; justify-content: center; padding: 14px; }
.photo-text { color: #999; font-size: 13px; }

/* Float actions right side */
.float-actions { position: absolute; right: 14px; top: 60px; }
.fa-item { background: rgba(0,0,0,0.55); color: #fff; width: 56px; padding: 8px 0; border-radius: 12px; text-align: center; margin-bottom: 10px; }
.fa-share { background: #07c160; }
.fa-icon { display: block; font-size: 18px; line-height: 1; margin-bottom: 2px; }
.fa-label { font-size: 10px; }

/* Ad bar */
.ad-bar { display: flex; justify-content: space-between; align-items: center; background: #fff8f0; padding: 10px 14px; margin: 0 14px; border-radius: 8px; font-size: 12px; }
.ad-text { color: #ff6b35; flex: 1; }
.ad-close { color: #999; font-size: 14px; padding: 0 6px; }

/* Like */
.like-row { padding: 20px 14px 12px; background: #fff; }
.like-btn { font-size: 24px; color: #1989fa; }

/* Contact */
.contact-row { padding: 4px 14px; background: #fff; font-size: 14px; }
.cl-label { color: #999; min-width: 60px; display: inline-block; }
.cl-name { color: #1989fa; font-weight: 500; }
.cl-tip { font-size: 12px; color: #999; line-height: 1.6; padding: 4px 0; }

/* Warn */
.warn-row { margin: 14px; padding: 14px; background: #fff5f5; border-radius: 8px; }
.warn-text { font-size: 13px; color: #333; font-weight: 500; display: block; margin-bottom: 6px; }
.warn-link { color: #1989fa; text-decoration: underline; }
.warn-sub { font-size: 11px; color: #999; }

/* Comment */
.comment-block { margin: 14px; background: #fff; border-radius: 12px; padding: 14px; }
.comment-head { border-bottom: 1px solid #f0f0f0; padding-bottom: 10px; margin-bottom: 12px; }
.ch-title { font-size: 15px; font-weight: 600; }
.comment-list { }
.comment-item { display: flex; gap: 10px; padding: 10px 0; }
.ci-ava { width: 36px; height: 36px; border-radius: 50%; background: #e8f2fc; }
.ci-meta { flex: 1; }
.ci-name { font-size: 13px; color: #666; }
.ci-text { display: block; font-size: 13px; color: #333; margin-top: 2px; }
.comment-empty { text-align: center; color: #999; font-size: 14px; padding: 20px; }

/* Other */
.other-block { margin: 14px; padding: 14px; background: #fff; border-radius: 12px; }
.ob-head { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.ob-ava { width: 40px; height: 40px; border-radius: 50%; background: #e8f2fc; }
.ob-name { font-size: 14px; font-weight: 500; }
.ob-stat { font-size: 12px; color: #666; }
.ob-num { color: #1989fa; font-weight: 600; }

/* FAB */
.fab { position: fixed; right: 14px; bottom: 80px; display: flex; flex-direction: column; gap: 10px; }
.fab-item { background: rgba(0,0,0,0.55); color: #fff; padding: 8px 6px; border-radius: 16px; text-align: center; min-width: 50px; }
.fab-on { background: #07c160; }
.fab-ico { font-size: 16px; display: block; }
.lab { font-size: 10px; }

/* Sticky Call */
.sticky-call { position: fixed; left: 0; right: 0; bottom: 0; padding: 12px; background: #fff; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.sc-btn { background: #1989fa; color: #fff; height: 48px; border-radius: 24px; display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 17px; font-weight: 600; }
.sc-ico { font-size: 20px; }
</style>
