<template>
  <view class="page-container">
    <u-nav-bar title="投递管理" show-back @back="goBack" />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 空态 -->
    <view v-else-if="!list.length" class="state-panel">
      <view class="state-mark">投</view>
      <text class="state-title">暂无投递</text>
      <text class="state-desc">职位发布后，求职者的投递会出现在这里</text>
    </view>

    <template v-else>
      <view class="list-head">
        <text class="list-title">候选人</text>
        <text class="list-count">共 {{ list.length }} 人</text>
      </view>

      <!-- 投递列表 -->
      <view class="apl-list">
        <view v-for="item in list" :key="item.application.id" class="apl-card">
          <view class="apl-row1" hover-class="row-press" :hover-stay-time="100" @tap="toggle(item.application.id)">
            <view class="apl-main">
              <text class="apl-name">{{ (item.resume && (item.resume.name || item.resume.title)) || '求职者' }}</text>
              <text v-if="item.resume && item.resume.phone" class="apl-phone">{{ item.resume.phone }}</text>
            </view>
            <text class="tag" :class="statusTagClass(item.application.status)">{{ statusLabel[item.application.status] || item.application.status }}</text>
          </view>
          <view class="apl-meta">
            <text>{{ item.resume && item.resume.education ? item.resume.education + ' · ' : '' }}{{ formatDate(item.application.created_at) }} 投递</text>
          </view>

          <!-- 简历详情（展开） -->
          <view v-if="expanded === item.application.id && item.resume" class="resume-box">
            <view v-if="item.resume.email" class="resume-line"><text class="resume-k">邮箱</text><text class="resume-v">{{ item.resume.email }}</text></view>
            <view v-if="item.resume.work_experience" class="resume-line"><text class="resume-k">工作经历</text><text class="resume-v">{{ item.resume.work_experience }}</text></view>
            <view v-if="item.resume.skills && item.resume.skills.length" class="resume-line">
              <text class="resume-k">技能</text>
              <view class="skill-tags"><text v-for="s in item.resume.skills" :key="s" class="skill-tag">{{ s }}</text></view>
            </view>
            <image
              v-if="item.resume.certificate_url"
              :src="fullUrl(item.resume.certificate_url)"
              mode="aspectFit"
              class="cert-img"
              @tap="previewCert(item.resume.certificate_url)"
            />
            <view v-if="item.resume.content" class="resume-line"><text class="resume-k">说明</text><text class="resume-v">{{ item.resume.content }}</text></view>
          </view>

          <!-- 操作（对齐后端状态机：submitted→viewed/rejected；viewed→interviewing/rejected；interviewing→offered/rejected；offered 终态） -->
          <view class="apl-actions">
            <view v-if="item.resume" class="apl-btn" hover-class="btn-press" :hover-stay-time="100" @tap="toggle(item.application.id)">
              {{ expanded === item.application.id ? '收起简历' : '查看简历' }}
            </view>
            <view v-if="item.application.status === 'submitted'" class="apl-btn apl-btn--primary" hover-class="btn-press" :hover-stay-time="100" @tap="updateStatus(item, 'viewed')">标记已查看</view>
            <view v-if="item.application.status === 'viewed'" class="apl-btn apl-btn--primary" hover-class="btn-press" :hover-stay-time="100" @tap="updateStatus(item, 'interviewing')">约面试</view>
            <view v-if="item.application.status === 'interviewing'" class="apl-btn apl-btn--primary" hover-class="btn-press" :hover-stay-time="100" @tap="updateStatus(item, 'offered')">录用</view>
            <view v-if="canReject(item.application.status)" class="apl-btn apl-btn--danger" hover-class="btn-press" :hover-stay-time="100" @tap="updateStatus(item, 'rejected')">婉拒</view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL } from '../../../utils/request'

const goBack = () => uni.navigateBack()
const list = ref([])
const loading = ref(false)
const expanded = ref('')

const statusLabel = { submitted: '待处理', viewed: '已查看', interviewing: '面试中', offered: '已录用', rejected: '未通过', withdrawn: '已撤回' }
// 标签配色与全站四色板同源：橙=待处理 蓝=进行中 紫=面试 绿=已录用 红=未通过 灰=已撤回
const STATUS_CLASS = { submitted: 'tag-orange', viewed: 'tag-blue', interviewing: 'tag-purple', offered: 'tag-green', rejected: 'tag-red', withdrawn: 'tag-gray' }
const statusTagClass = (s) => STATUS_CLASS[s] || 'tag-gray'
const canReject = (s) => s === 'submitted' || s === 'viewed' || s === 'interviewing'

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate()) + ' ' + p(dt.getHours()) + ':' + p(dt.getMinutes())
}
const fullUrl = (u) => (u && u.startsWith('http') ? u : BASE_URL + (u || ''))
const previewCert = (u) => uni.previewImage({ urls: [fullUrl(u)] })
const toggle = (id) => { expanded.value = expanded.value === id ? '' : id }

const load = async (jobId) => {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/applications', data: { job_id: jobId } })
    list.value = Array.isArray(res) ? res : ((res && res.data) || [])
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const updateStatus = async (item, status) => {
  try {
    await request({
      url: '/api/v1/applications/' + encodeURIComponent(item.application.id) + '/status',
      method: 'PATCH',
      data: { status },
    })
    uni.showToast({ title: '已更新', icon: 'success' })
    item.application.status = status
  } catch (e) {
    uni.showToast({ title: (e && e.message) || '操作失败', icon: 'none' })
  }
}

onLoad((opts) => { if (opts.job_id) load(opts.job_id) })
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

/* ===== 头部：标题 + 计数 ===== */
.list-head { display: flex; align-items: baseline; gap: 16rpx; padding: 32rpx 32rpx 16rpx; }
.list-title { font-size: 36rpx; font-weight: 700; color: #17212B; }
.list-count { font-size: 24rpx; color: #667085; }

/* ===== 候选人卡片（对齐「我的发布」卡片规范） ===== */
.apl-list { display: flex; flex-direction: column; gap: 20rpx; padding: 0 32rpx; }
.apl-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 26rpx;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.045);
}
.apl-row1 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  min-height: 56rpx;
}
.row-press { opacity: 0.85; }
.apl-main { display: flex; align-items: center; gap: 16rpx; min-width: 0; }
.apl-name { font-size: 30rpx; font-weight: 700; color: #17212B; }
.apl-phone { font-size: 24rpx; color: #0A66C2; }

.tag { border-radius: 8rpx; padding: 6rpx 12rpx; font-size: 20rpx; line-height: 1; flex: none; }
.tag-blue { color: #0A66C2; background: #EAF3FB; }
.tag-green { color: #168A55; background: #E9F7F0; }
.tag-orange { color: #DB5F0D; background: #FFF0E6; }
.tag-purple { color: #7B61D1; background: #F0EDFF; }
.tag-red { color: #D92D20; background: #FEF3F2; }
.tag-gray { color: #667085; background: #F1F3F5; }

.apl-meta { font-size: 22rpx; color: #667085; margin-top: 12rpx; }

.resume-box {
  margin-top: 20rpx;
  padding: 22rpx;
  background: #F7F8FA;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.resume-line { display: flex; gap: 16rpx; font-size: 24rpx; color: #17212B; }
.resume-k { color: #667085; flex: none; width: 112rpx; }
.resume-v { flex: 1; min-width: 0; line-height: 1.6; }
.skill-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.skill-tag { font-size: 20rpx; padding: 6rpx 12rpx; background: #EAF3FB; color: #0A66C2; border-radius: 8rpx; line-height: 1; }
.cert-img { width: 100%; height: 320rpx; background: #fff; border-radius: 12rpx; }

/* 操作按钮：语义状态流转，实心=下一步，浅红=婉拒 */
.apl-actions { display: flex; flex-wrap: wrap; gap: 16rpx; margin-top: 22rpx; padding-top: 20rpx; border-top: 1px solid #EEF1F4; }
.apl-btn {
  min-height: 72rpx;
  padding: 0 28rpx;
  border-radius: 999rpx;
  border: 1px solid #E4E7EC;
  background: #fff;
  color: #17212B;
  font-size: 24rpx;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.apl-btn--primary { background: #0A66C2; border-color: #0A66C2; color: #fff; }
.apl-btn--danger { background: #FEF3F2; border-color: #FECDCA; color: #D92D20; }
.btn-press { transform: scale(0.96); opacity: 0.9; }

/* ===== 状态面板（空态，全局同款） ===== */
.state-panel {
  min-height: 620rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
  text-align: center;
}
.state-mark {
  width: 124rpx;
  height: 124rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 54rpx;
  font-weight: 700;
}
.state-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.state-desc { margin-top: 14rpx; font-size: 24rpx; color: #667085; line-height: 1.6; }

/* ===== 加载中 ===== */
.loading-state { display: flex; align-items: center; justify-content: center; gap: 16rpx; padding: 120rpx 0; }
.loading-text { font-size: 24rpx; color: #667085; }
</style>
