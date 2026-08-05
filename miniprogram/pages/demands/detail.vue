<template>
  <view class="dd-page">
    <view class="dd-header">
      <text class="back-btn" @tap="goBack">‹</text>
      <text class="title">需求详情</text>
      <view class="more-btn" @tap="showMoreActions">⋯</view>
    </view>

    <view v-if="post" class="dd-content">
      <!-- 驳回/关闭原因（发布者可见） -->
      <view v-if="post.status === 'rejected' && post.reject_reason" class="reject-banner">
        <text class="reject-tag">已驳回</text>
        <text class="reject-text">{{ post.reject_reason }}</text>
      </view>
      <view v-else-if="post.status === 'cancelled' && post.reject_reason" class="reject-banner cancelled">
        <text class="reject-tag">已关闭</text>
        <text class="reject-text">{{ post.reject_reason }}</text>
      </view>

      <!-- 决策摘要 -->
      <view class="summary-card">
        <view class="summary-tags">
          <text class="tag-blue">{{ bizTypeLabel(post.tag) }}</text>
          <text v-if="post.status" class="tag-status">{{ statusLabel(post.status) }}</text>
        </view>
        <text class="summary-title">{{ post.title }}</text>
        <view class="summary-meta">
          <view class="meta-block">
            <text class="meta-label">预算</text>
            <text class="meta-value budget">{{ formatBudget(post.budget_fen) }}</text>
          </view>
          <view class="meta-block">
            <text class="meta-label">地区</text>
            <text class="meta-value">{{ post.location || '重庆' }}</text>
          </view>
          <view class="meta-block">
            <text class="meta-label">发布时间</text>
            <text class="meta-value">{{ post.time }}</text>
          </view>
        </view>
      </view>

      <!-- 需求描述 -->
      <view class="section-block">
        <text class="section-title">需求描述</text>
        <text class="desc-text">{{ post.desc }}</text>
      </view>

      <!-- 图片证据 -->
      <view v-if="post.photos && post.photos.length" class="section-block">
        <text class="section-title">相关图片</text>
        <view class="photo-grid">
          <image
            v-for="(p, i) in post.photos"
            :key="i"
            :src="p"
            mode="aspectFill"
            class="photo"
            @tap="preview(i)"
          />
        </view>
      </view>

      <!-- 发布方 -->
      <view class="publisher-row">
        <view class="pub-avatar">{{ avatarText }}</view>
        <view class="pub-info">
          <text class="pub-name">{{ post.userName }}</text>
          <text class="pub-time">发布于 {{ post.time }}</text>
        </view>
      </view>

      <!-- 对接说明 + 风险提示 -->
      <view class="notice-block">
        <text class="notice-title">对接说明</text>
        <text class="notice-line">· 点击「联系对接」登记意向，发布方将看到您的联系方式</text>
        <text class="notice-line">· 双方在线下联系洽谈，平台不参与资金流转</text>
        <text class="notice-line warn">· 为保障安全，请见面交易，切勿提前支付任何费用</text>
      </view>
    </view>

    <!-- 底部操作栏：联系对接（主）+ 拨打电话（次） -->
    <view class="sticky-call" v-if="post">
      <view class="sc-btn sc-secondary" @tap="callPhone">
        <text>拨打电话</text>
      </view>
      <view class="sc-btn sc-primary" @tap="showIntentSheet = true">
        <text>联系对接</text>
      </view>
    </view>

    <!-- 联系对接意向登记弹层 -->
    <u-popup :show="showIntentSheet" position="bottom" round @close="showIntentSheet = false">
      <view class="intent-sheet">
        <text class="intent-title">登记对接意向</text>
        <text class="intent-desc">发布方将看到您的联系方式和备注，线下沟通成交</text>
        <view class="intent-field">
          <text class="intent-label">姓名</text>
          <input v-model="intentForm.name" class="intent-input" placeholder="您的称呼" />
        </view>
        <view class="intent-field">
          <text class="intent-label">联系电话</text>
          <input v-model="intentForm.contact" class="intent-input" type="number" placeholder="必填，方便发布方联系您" />
        </view>
        <view class="intent-field">
          <text class="intent-label">备注</text>
          <textarea v-model="intentForm.remark" class="intent-textarea" placeholder="可简要说明您的作业能力/意向（选填）" />
        </view>
        <view class="intent-submit" @tap="submitIntent">
          <text>提交对接意向</text>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../utils/request'
import { bizTypeLabel } from '../../utils/enums'

const post = ref(null)
const showIntentSheet = ref(false)
const intentForm = ref({ name: '', contact: '', remark: '' })
const submittingIntent = ref(false)

// 页面 id 从 onLoad 的 options 取（setup 顶层执行时页面尚未注册进
// getCurrentPages() 栈，直接读取会拿到空值导致数据不加载）
let postId = ''

const avatarText = computed(() => {
  const name = post.value?.userName || '用'
  return name.slice(0, 1)
})

const goBack = () => uni.navigateBack()
const preview = (i) => uni.previewImage({ urls: post.value.photos, current: i })
const callPhone = () => post.value?.phone && uni.makePhoneCall({ phoneNumber: post.value.phone })

const showMoreActions = () => {
  uni.showActionSheet({
    itemList: ['分享', '举报'],
    success: (res) => {
      if (res.tapIndex === 0) uni.showShareMenu()
      if (res.tapIndex === 1) uni.showToast({ title: '举报已提交', icon: 'none' })
    },
  })
}

const statusLabel = (s) => ({ pending: '待审核', published: '已发布', completed: '已完成', cancelled: '已取消', rejected: '已驳回' }[s] || '')
const formatBudget = (fen) => {
  if (fen == null || fen === 0) return '面议'
  var yuan = (fen / 100).toFixed(2)
  return '¥' + yuan.replace(/\.00$/, '')
}

// ---- 数据加载 ----
const loadDetail = async () => {
  if (!postId) return
  try {
    const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(postId) })
    const d = (res && res.data) || res
    if (d && d.id) {
      post.value = {
        tag: d.biz_type || '需求',
        userName: d.publisher_name || '平台用户',
        time: (d.created_at || '').slice(0, 16).replace('T', ' '),
        title: d.title || '',
        location: d.district || '',
        desc: d.description || '暂无详细描述',
        photos: parseImgs(d.images),
        phone: d.contact || '',
        status: d.status || '',
        budget_fen: d.budget_fen,
        reject_reason: (d.biz_fields && d.biz_fields.reject_reason) || ''
      }
    }
  } catch (e) { /* 保持空态 */ }
}
const parseImgs = (imgs) => {
  try {
    const arr = typeof imgs === 'string' ? JSON.parse(imgs) : (imgs || [])
    return arr.map(u => (u && u.startsWith('http') ? u : BASE_URL + u))
  } catch (e) { return [] }
}

// ---- 联系对接意向登记 ----
const submitIntent = async () => {
  if (submittingIntent.value) return
  if (!intentForm.value.contact) {
    uni.showToast({ title: '请填写联系电话', icon: 'none' })
    return
  }
  submittingIntent.value = true
  try {
    await request({
      url: '/api/v1/demands/' + encodeURIComponent(postId) + '/intents',
      method: 'POST',
      data: {
        intentor_name: intentForm.value.name,
        contact: intentForm.value.contact,
        remark: intentForm.value.remark,
      },
    })
    uni.showToast({ title: '对接意向已登记', icon: 'success' })
    showIntentSheet.value = false
    intentForm.value = { name: '', contact: '', remark: '' }
  } catch (e) {
    uni.showToast({ title: '提交失败，请稍后重试', icon: 'none' })
  } finally {
    submittingIntent.value = false
  }
}

onLoad((options) => {
  postId = (options && options.id) || ''
  loadDetail()
})
</script>

<style scoped>
.dd-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 96px;
}

/* Header */
.dd-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-btn { font-size: 26px; color: #17212B; line-height: 1; width: 40px; }
.title { flex: 1; font-size: 17px; font-weight: 600; color: #17212B; }
.more-btn { font-size: 20px; color: #667085; padding: 0 6px; }

.dd-content {
  padding: 12px;
}

/* 驳回 banner */
.reject-banner {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  background: #FEF3F2;
  margin-bottom: 8px;
}

.reject-banner.cancelled { background: #FFF7F1; }

.reject-tag {
  font-size: 12px;
  font-weight: 700;
  color: #D92D20;
  flex-shrink: 0;
}

.reject-banner.cancelled .reject-tag { color: #B54708; }

.reject-text {
  font-size: 13px;
  color: #344054;
  line-height: 1.5;
}

/* 决策摘要 */
.summary-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 8px;
}

.summary-tags {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.tag-blue {
  font-size: 11px;
  color: #0A66C2;
  background: #EAF3FB;
  padding: 3px 8px;
  border-radius: 4px;
}

.tag-status {
  font-size: 11px;
  color: #168A55;
  background: #E9F7F0;
  padding: 3px 8px;
  border-radius: 4px;
}

.summary-title {
  font-size: 19px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.35;
  display: block;
}

.summary-meta {
  display: flex;
  gap: 24px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #EEF1F4;
}

.meta-block {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.meta-label {
  font-size: 11px;
  color: #98A2B3;
}

.meta-value {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
}

.meta-value.budget {
  color: #E96012;
  font-size: 17px;
}

/* 分组区块 */
.section-block {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 8px;
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  display: block;
  margin-bottom: 10px;
}

.desc-text {
  font-size: 14px;
  color: #344054;
  line-height: 1.7;
  white-space: pre-wrap;
}

/* 图片 */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
}

.photo {
  width: 100%;
  height: 104px;
  border-radius: 8px;
  background: #F4F6F8;
}

/* 发布方 */
.publisher-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
}

.pub-avatar {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pub-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pub-name {
  font-size: 14px;
  font-weight: 600;
  color: #17212B;
}

.pub-time {
  font-size: 11px;
  color: #98A2B3;
}

/* 对接说明 */
.notice-block {
  background: #F4F8FC;
  border-radius: 8px;
  padding: 12px 16px;
}

.notice-title {
  font-size: 13px;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 6px;
}

.notice-line {
  display: block;
  font-size: 12px;
  color: #344054;
  line-height: 1.7;
}

.notice-line.warn {
  color: #B54708;
}

/* 底部操作栏 */
.sticky-call {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px;
  background: #fff;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  display: flex;
  gap: 10px;
}

.sc-btn {
  flex: 1;
  height: 46px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}

.sc-primary {
  background: #0A66C2;
  color: #fff;
}

.sc-secondary {
  background: #fff;
  color: #0A66C2;
  border: 1px solid #0A66C2;
}

/* 意向弹层 */
.intent-sheet {
  padding: 24px 16px calc(24px + env(safe-area-inset-bottom));
}

.intent-title {
  display: block;
  font-size: 18px;
  font-weight: 700;
  color: #17212B;
}

.intent-desc {
  display: block;
  font-size: 12px;
  color: #667085;
  margin: 4px 0 16px;
}

.intent-field {
  margin-bottom: 14px;
}

.intent-label {
  display: block;
  font-size: 13px;
  color: #344054;
  margin-bottom: 6px;
  font-weight: 500;
}

.intent-input {
  height: 44px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  padding: 0 12px;
  font-size: 14px;
  background: #fff;
}

.intent-textarea {
  width: 100%;
  box-sizing: border-box;
  height: 88px;
  border: 1px solid #E4E7EC;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 14px;
  background: #fff;
}

.intent-submit {
  margin-top: 8px;
  height: 46px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}
</style>
