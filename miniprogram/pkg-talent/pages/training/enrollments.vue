<template>
  <view class="er-page">
    <!-- 课程选择 -->
    <view v-if="myCourses.length" class="er-picker" @tap="showPicker = true">
      <text class="er-picker-label">课程</text>
      <text class="er-picker-value">{{ currentCourse ? currentCourse.title : '选择课程' }}</text>
      <text class="er-arrow">›</text>
    </view>
    <u-popup :show="showPicker" position="bottom" round @close="showPicker = false">
      <view class="er-sheet">
        <scroll-view scroll-y class="er-sheet-list">
          <view v-for="c in myCourses" :key="c.id" class="er-sheet-item" :class="{ on: c.id === courseId }" @tap="pickCourse(c)">
            <text class="er-sheet-name">{{ c.title }}</text>
            <text class="er-sheet-meta">{{ c.org_name || '' }} · 状态 {{ statusText[c.status] || c.status }}</text>
          </view>
        </scroll-view>
      </view>
    </u-popup>

    <view v-if="loading" class="er-state">加载中...</view>

    <view v-else-if="!myCourses.length" class="er-empty">
      <view class="er-empty-mark">!</view>
      <text class="er-empty-title">暂无机构课程</text>
      <text class="er-empty-desc">先在「发布-发布培训课程」创建课程，通过审核后可管理报名</text>
    </view>

    <view v-else-if="!courseId" class="er-empty">
      <view class="er-empty-mark">›</view>
      <text class="er-empty-title">请选择课程</text>
      <text class="er-empty-desc">选择一门课程查看和管理报名</text>
    </view>

    <view v-else-if="!list.length" class="er-empty">
      <view class="er-empty-mark">!</view>
      <text class="er-empty-title">暂无报名</text>
      <text class="er-empty-desc">学员报名后展示在这里，可审核通过/拒绝</text>
    </view>

    <view v-else class="er-list">
      <view v-for="it in list" :key="it.id" class="er-card">
        <view class="er-head">
          <text class="er-name">{{ it.name }}</text>
          <text class="er-status" :class="'er-status--' + it.status">{{ statusText[it.status] || it.status }}</text>
        </view>
        <view class="er-meta">
          <text class="er-meta-item">电话 {{ it.phone || '—' }}</text>
          <text v-if="it.id_card" class="er-meta-item">身份证 {{ it.id_card }}</text>
        </view>
        <view v-if="it.photo_url || it.id_card_image || it.id_card_back" class="er-photos">
          <image v-if="it.photo_url" class="er-photo" :src="BASE_URL + it.photo_url" mode="aspectFill" @tap="preview(it)" />
          <image v-if="it.id_card_image" class="er-photo" :src="BASE_URL + it.id_card_image" mode="aspectFill" @tap="preview(it)" />
          <image v-if="it.id_card_back" class="er-photo" :src="BASE_URL + it.id_card_back" mode="aspectFill" @tap="preview(it)" />
        </view>
        <text v-if="it.review_note" class="er-note">审核备注：{{ it.review_note }}</text>
        <view v-if="it.status === 'enrolled' || it.status === 'paid'" class="er-actions">
          <view class="er-btn er-btn--ok" @tap="review(it, 'approve')">通过</view>
          <view class="er-btn er-btn--no" @tap="review(it, 'reject')">拒绝</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request, getStoredUser, getErrorMessage, requireLogin, BASE_URL } from '../../../utils/request'

const statusText = { enrolled: '已报名', paid: '已缴费', approved: '已通过', rejected: '已拒绝', completed: '已结业' }
const myCourses = ref([])
const courseId = ref('')
const list = ref([])
const loading = ref(true)
const showPicker = ref(false)

const user = getStoredUser()
const myId = user && (user.id || user.user_id)
const currentCourse = computed(() => myCourses.value.find((c) => c.id === courseId.value) || null)

const loadCourses = async () => {
  try {
    const res = await request({ url: '/api/v1/training-courses' })
    const all = Array.isArray(res) ? res : ((res && res.data) || [])
    myCourses.value = all.filter((c) => c.org_id === myId)
  } catch (e) {
    myCourses.value = []
  } finally {
    // 必须在这里收掉 loading：此前只有 loadList() 置 false，而 loadList() 首行是
    // if (!courseId) return —— 没选课程时页面会永久停在"加载中..."（BUG-003）。
    loading.value = false
  }
}

const loadList = async () => {
  if (!courseId.value) return
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/training-courses/' + encodeURIComponent(courseId.value) + '/enrollments' })
    list.value = Array.isArray(res) ? res : ((res && res.data) || [])
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const preview = (it) => {
  const urls = ['photo_url', 'id_card_image', 'id_card_back'].map((k) => it[k]).filter(Boolean).map((p) => BASE_URL + p)
  if (urls.length) uni.previewImage({ urls })
}

const pickCourse = (c) => {
  courseId.value = c.id
  showPicker.value = false
  loadList()
}

const review = (it, action) => {
  const doReview = (reason) =>
    request({ url: '/api/v1/enrollments/' + encodeURIComponent(it.id) + '/review', method: 'POST', data: { action, reason: reason || '' } })
      .then(() => {
        uni.showToast({ title: action === 'approve' ? '已通过' : '已拒绝', icon: 'success' })
        loadList()
      })
      .catch((e) => uni.showToast({ title: getErrorMessage(e) || '操作失败，请重试', icon: 'none' }))
  if (action === 'reject') {
    uni.showModal({
      title: '拒绝报名',
      editable: true,
      placeholderText: '请填写拒绝原因（必填）',
      success: (r) => { if (r.confirm) doReview((r.content || '').trim()) }
    })
  } else {
    uni.showModal({ title: '通过报名', content: '确认通过该学员报名？', success: (r) => { if (r.confirm) doReview('') } })
  }
}

onShow(() => {
  if (requireLogin('请先登录后管理课程报名', '/pages/home/index')) loadCourses()
})
</script>

<style>
page { background: var(--color-bg); }
.er-page { min-height: 100vh; padding: 20rpx; box-sizing: border-box; }
.er-picker { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 10px; padding: 26rpx 24rpx; display: flex; align-items: center; gap: 16rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.er-picker-label { font-size: 26rpx; color: #667085; flex-shrink: 0; }
.er-picker-value { flex: 1; font-size: 28rpx; color: #17212B; font-weight: 600; }
.er-arrow { color: #98A2B3; }
.er-sheet { padding: 20rpx 24rpx calc(24rpx + env(safe-area-inset-bottom)); max-height: 60vh; }
.er-sheet-list { max-height: 52vh; }
.er-sheet-item { padding: 22rpx 8rpx; border-bottom: 1rpx solid #F0F2F5; }
.er-sheet-item.on .er-sheet-name { color: #0A66C2; }
.er-sheet-name { display: block; font-size: 28rpx; color: #17212B; }
.er-sheet-meta { display: block; font-size: 22rpx; color: #98A2B3; margin-top: 6rpx; }
.er-state { text-align: center; padding: 120rpx 0; color: #98A2B3; font-size: 26rpx; }
.er-empty { text-align: center; padding: 100rpx 40rpx 0; }
.er-empty-mark { width: 96rpx; height: 96rpx; border-radius: 50%; background: #F4F6F8; color: #98A2B3; font-size: 44rpx; display: flex; align-items: center; justify-content: center; margin: 0 auto 24rpx; }
.er-empty-title { display: block; font-size: 30rpx; font-weight: 600; color: #17212B; }
.er-empty-desc { display: block; font-size: 24rpx; color: #98A2B3; margin-top: 10rpx; line-height: 1.6; }
.er-list { margin-top: 20rpx; }
.er-card { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 10px; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.er-head { display: flex; justify-content: space-between; align-items: center; }
.er-name { font-size: 30rpx; font-weight: 600; color: #17212B; }
.er-status { font-size: 22rpx; font-weight: 600; padding: 6rpx 16rpx; border-radius: 6rpx; }
.er-status--enrolled { background: #EAF3FB; color: #0A66C2; }
.er-status--paid { background: #FFF4E5; color: #B54708; }
.er-status--approved { background: #E9F7F0; color: #0B6B41; }
.er-status--rejected { background: #FEF3F2; color: #D92D20; }
.er-status--completed { background: #F1F3F5; color: #667085; }
.er-meta { margin-top: 12rpx; }
.er-meta-item { font-size: 24rpx; color: #475467; margin-right: 24rpx; }
.er-note { display: block; margin-top: 10rpx; font-size: 24rpx; color: #D92D20; }
.er-photos { display: flex; gap: 14rpx; margin-top: 14rpx; }
.er-photo { width: 104rpx; height: 104rpx; border-radius: 8rpx; background: #F4F6F8; }
.er-actions { display: flex; gap: 16rpx; margin-top: 18rpx; }
.er-btn { flex: 1; height: 76rpx; border-radius: 38rpx; display: flex; align-items: center; justify-content: center; font-size: 26rpx; font-weight: 600; }
.er-btn--ok { background: #0A66C2; color: #fff; }
.er-btn--no { background: #FEF3F2; color: #D92D20; }
</style>
