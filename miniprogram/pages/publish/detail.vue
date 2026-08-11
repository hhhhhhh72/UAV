<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">发布详情</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="editPost">编辑</view>
    </view>

    <!-- 状态 + 信息卡 -->
    <view class="pub-detail-status" :class="statusCls">
      <text class="pub-status-dot"></text>
      <text>{{ post.status }}</text>
    </view>

    <view class="pub-detail-card">
      <view class="pub-post-type" style="display:inline-block">{{ post.label }}</view>
      <view class="pub-detail-title">{{ post.title }}</view>
      <view class="pub-detail-meta">
        <text v-for="(m, i) in post.meta" :key="i">{{ m }}</text>
      </view>
      <view class="pub-detail-note">{{ post.note }}</view>
    </view>

    <!-- 处理进度 -->
    <view class="pub-section-title">处理进度</view>
    <view class="pub-timeline">
      <view class="pub-timeline-row">
        <view class="pub-timeline-mark"></view>
        <view class="pub-timeline-body">
          <text class="pub-timeline-b">已保存发布信息</text>
          <text class="pub-timeline-date">{{ post.date }}</text>
        </view>
      </view>
      <view class="pub-timeline-row pub-timeline-row--current">
        <view class="pub-timeline-mark"></view>
        <view class="pub-timeline-body">
          <text class="pub-timeline-b">{{ currentStepTitle }}</text>
          <text class="pub-timeline-date">{{ currentStepDesc }}</text>
        </view>
      </view>
    </view>

    <!-- 动作按钮 -->
    <view class="pub-detail-actions">
      <view
        class="pub-btn"
        :class="secondaryCls"
        hover-class="pub-btn--active"
        @tap="openAction(secondaryAction)"
      >{{ secondaryLabel }}</view>
      <view
        class="pub-btn pub-btn--primary"
        hover-class="pub-btn--active"
        @tap="editPost"
      >{{ primaryLabel }}</view>
    </view>

    <!-- 二次确认弹窗 -->
    <view v-if="actionModal" class="pub-modal" @tap="actionModal = null">
      <view class="pub-dialog" @tap.stop>
        <view class="pub-dialog-title">确认{{ actionModal }}？</view>
        <view class="pub-dialog-text">{{ actionText }}</view>
        <view class="pub-dialog-actions">
          <view class="pub-dialog-btn" @tap="actionModal = null">取消</view>
          <view class="pub-dialog-btn" :class="{ 'pub-dialog-btn--danger': isDeleteAction }" @tap="applyAction">
            确认{{ actionModal }}
          </view>
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
import { getPost, removePost, upsertPost } from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const id = ref('')
const backTab = ref('all')
const backKind = ref('all')
const post = ref(null)
const actionModal = ref('')
const toast = ref('')
const toastTimer = ref(null)

const statusCls = computed(() => (post.value ? 'status-' + post.value.statusKey : 'status-draft'))

const statusTextMap = {
  pending: '平台正在审核',
  live: '内容已公开展示',
  rejected: '等待补充后重新提交',
  draft: '草稿待完善',
}
const statusDescMap = {
  pending: '通常在 1 个工作日内完成',
  live: '可继续查看咨询与对接意向',
  rejected: '补充机构或资质证明后可再次提交',
  draft: '完善必填信息后提交审核',
}
const currentStepTitle = computed(() => (post.value ? statusTextMap[post.value.statusKey] : ''))
const currentStepDesc = computed(() => (post.value ? statusDescMap[post.value.statusKey] : ''))

const actionMap = {
  pending: { primary: '修改内容', secondary: '撤回申请' },
  live: { primary: '编辑发布', secondary: '下架发布' },
  rejected: { primary: '重新编辑', secondary: '删除发布' },
  draft: { primary: '继续编辑', secondary: '删除草稿' },
}
const primaryLabel = computed(() => (post.value ? actionMap[post.value.statusKey].primary : ''))
const secondaryLabel = computed(() => (post.value ? actionMap[post.value.statusKey].secondary : ''))
const secondaryAction = computed(() => (post.value ? actionMap[post.value.statusKey].secondary : ''))

const secondaryCls = computed(() => {
  if (!post.value) return 'pub-btn--secondary'
  const k = post.value.statusKey
  return k === 'draft' || k === 'rejected' ? 'pub-btn--danger' : 'pub-btn--secondary'
})

const isDeleteAction = computed(() => actionModal.value && actionModal.value.startsWith('删除'))

const actionText = computed(() => {
  const a = actionModal.value
  if (!a) return ''
  if (a.startsWith('删除')) return '删除后无法恢复该发布内容。'
  if (a === '撤回申请') return '撤回后内容将转为草稿，您可以修改后再次提交。'
  return '下架后将不再公开展示，但可在草稿中继续编辑。'
})

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => { toast.value = '' }, 2200)
}

function openAction(label) {
  actionModal.value = label
}

function applyAction() {
  const label = actionModal.value
  const p = post.value
  actionModal.value = ''
  if (label.startsWith('删除')) {
    removePost(p.id)
    showToast('发布内容已删除')
    setTimeout(goBack, 600)
    return
  }
  // 撤回 / 下架：转为草稿
  p.status = '草稿'
  p.statusKey = 'draft'
  p.date = '刚刚保存'
  p.note = '内容已保存为草稿，可继续编辑后再次提交。'
  upsertPost(p)
  showToast(label + '成功，内容已移入草稿')
  setTimeout(goBack, 600)
}

function editPost() {
  uni.navigateTo({
    url: '/pages/publish/form?type=' + post.value.type + '&id=' + post.value.id,
  })
}

function goBack() {
  uni.navigateBack({
    success: () => {},
  })
}

onLoad((options) => {
  initSafeTop()
  id.value = (options && options.id) || ''
  backTab.value = (options && options.tab) || 'all'
  backKind.value = (options && options.kind) || 'all'
  const p = getPost(id.value)
  if (!p) {
    // 已被删除
    uni.navigateBack()
    return
  }
  post.value = p
})
</script>

<style scoped>
@import './pub-style.css';
.pub-fade { opacity: 0.6; }
.pub-timeline-body { flex: 1; }
.pub-timeline-b { display: block; font-size: 12px; color: #17212B; }
.pub-timeline-date { font-size: 11px; color: #98A2B3; }
</style>
