<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布预览</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="goBack">编辑</view>
    </view>

    <!-- 预览卡片 -->
    <view class="pub-preview-card">
      <view class="pub-preview-type">{{ typeConfig.short }} · 待审核</view>
      <view class="pub-preview-title">{{ title }}</view>
      <view class="pub-preview-meta">
        <text v-for="(m, i) in metaList" :key="i">{{ m }}</text>
      </view>
      <view class="pub-preview-copy">{{ copyText }}</view>
    </view>

    <!-- 审核说明 -->
    <view class="pub-review-note">
      <text class="pub-review-note-b">审核说明</text>
      <text>提交后平台将核验内容合规性；联系方式仅向登录并发起对接的用户展示。</text>
    </view>

    <!-- 发布前检查 -->
    <view class="pub-section">
      <view class="pub-section-title">发布前检查</view>
      <view class="pub-form-card">
        <view class="pub-check-row"><text class="pub-check-mark">✓</text><text>已填写必填项目</text></view>
        <view class="pub-check-row"><text class="pub-check-mark">✓</text><text>已确认发布内容真实性</text></view>
        <view class="pub-check-row">
          <text class="pub-check-mark">✓</text>
          <text>{{ photoCount ? '已添加 ' + photoCount + ' 张图片' : '可补充图片提升展示效果' }}</text>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区 -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="goBack">返回修改</view>
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="openConfirm">提交审核</view>
    </view>

    <!-- 确认弹窗 -->
    <view v-if="showConfirm" class="pub-modal" @tap="showConfirm = false">
      <view class="pub-dialog" @tap.stop>
        <view class="pub-dialog-title">确认提交审核？</view>
        <view class="pub-dialog-text">{{ typeConfig.name }}提交后将进入审核。审核前可在「我的发布」中继续修改草稿。</view>
        <view class="pub-dialog-actions">
          <view class="pub-dialog-btn" @tap="showConfirm = false">再检查一下</view>
          <view class="pub-dialog-btn" @tap="submitPublish">确认提交</view>
        </view>
      </view>
    </view>

    <!-- 成功弹窗 -->
    <view v-if="showSuccess" class="pub-modal">
      <view class="pub-dialog pub-success">
        <view class="pub-success-mark">✓</view>
        <view class="pub-dialog-title">已提交审核</view>
        <view class="pub-dialog-text">{{ typeConfig.name }}已保存。审核通过后会公开展示，并可在「我的发布」中查看状态。</view>
        <view class="pub-dialog-actions">
          <view class="pub-dialog-btn" @tap="goHome">返回发布首页</view>
          <view class="pub-dialog-btn" @tap="goMyPosts">查看我的发布</view>
        </view>
      </view>
    </view>

    <!-- 底部黑色 toast -->
    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { TYPES, computePreviewMeta, makePost, upsertPost, loadFormState, clearFormState } from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const type = ref('')
const values = ref({})
const photoCount = ref(0)
const resumeId = ref('')
const showConfirm = ref(false)
const showSuccess = ref(false)
const toast = ref('')
const toastTimer = ref(null)

const typeConfig = computed(() => TYPES[type.value] || null)

const title = computed(() => values.value.title || '待命名发布内容')
const metaList = computed(() => computePreviewMeta(type.value, values.value))
const copyText = computed(() => values.value.description || '暂未填写补充说明。发布前仍可返回编辑，审核通过后将在相应的大厅或列表页展示。')

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

function goBack() {
  uni.navigateBack()
}

function openConfirm() {
  showConfirm.value = true
}

function submitPublish() {
  const t = typeConfig.value
  if (!t) return
  const post = makePost({
    id: resumeId.value || '',
    type: type.value,
    values: Object.assign({}, values.value),
    photoCount: photoCount.value,
    statusKey: 'pending',
    status: '审核中',
    date: '今日 刚刚提交',
    note: '平台正在核验发布内容与联系人信息。',
  })
  upsertPost(post)
  showConfirm.value = false
  showSuccess.value = true
}

function goHome() {
  clearFormState()
  uni.switchTab({ url: '/pages/publish/index' })
}

function goMyPosts() {
  clearFormState()
  uni.navigateTo({ url: '/pages/publish/my-posts?tab=pending' })
}

onLoad((options) => {
  initSafeTop()
  // 从表单页带入的临时状态
  const state = options && options.state
  if (state) {
    try {
      const parsed = JSON.parse(decodeURIComponent(state))
      type.value = parsed.type || ''
      values.value = parsed.values || {}
      photoCount.value = parsed.photoCount || 0
      resumeId.value = parsed.resumeId || ''
      return
    } catch (e) { /* fallthrough */ }
  }
  // 兜底：读取表单页写入的 storage
  const st = loadFormState()
  if (st) {
    type.value = st.type || ''
    values.value = st.values || {}
    photoCount.value = st.photoCount || 0
    resumeId.value = st.resumeId || ''
  }
})
</script>

<style scoped>
@import './pub-style.css';
.pub-fade { opacity: 0.6; }
.pub-review-note-b { color: #0A66C2; font-weight: 700; }
</style>
