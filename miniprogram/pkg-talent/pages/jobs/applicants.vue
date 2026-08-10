<template>
  <view class="page-container">
    <u-nav-bar title="投递管理" show-back @back="goBack" />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text class="loading-text">加载中...</text>
    </view>

    <!-- Empty -->
    <view v-else-if="!list.length" class="empty-state">
      <u-empty description="暂无投递" />
    </view>

    <!-- 投递列表 -->
    <view v-else class="apl-list">
      <view v-for="item in list" :key="item.application.id" class="apl-card">
        <view class="apl-row1" @tap="toggle(item.application.id)">
          <view class="apl-main">
            <text class="apl-name">{{ (item.resume && (item.resume.name || item.resume.title)) || '求职者' }}</text>
            <text v-if="item.resume && item.resume.phone" class="apl-phone">{{ item.resume.phone }}</text>
          </view>
          <u-tag :type="statusTag(item.application.status)" size="mini">{{ statusLabel[item.application.status] || item.application.status }}</u-tag>
        </view>
        <view class="apl-meta">
          <text>{{ item.resume && item.resume.education ? item.resume.education + ' · ' : '' }}{{ formatDate(item.application.created_at) }}</text>
        </view>

        <!-- 简历详情（展开） -->
        <view v-if="expanded === item.application.id && item.resume" class="resume-box">
          <view v-if="item.resume.email" class="resume-line"><text class="resume-k">邮箱</text><text>{{ item.resume.email }}</text></view>
          <view v-if="item.resume.work_experience" class="resume-line"><text class="resume-k">工作经历</text><text>{{ item.resume.work_experience }}</text></view>
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
          <view v-if="item.resume.content" class="resume-line"><text class="resume-k">说明</text><text>{{ item.resume.content }}</text></view>
        </view>

        <!-- 操作按钮 -->
        <view class="apl-actions">
          <u-button v-if="item.application.status === 'submitted'" size="mini" type="primary" plain round @tap="updateStatus(item, 'viewed')">已查看</u-button>
          <u-button v-if="item.application.status === 'submitted' || item.application.status === 'viewed'" size="mini" type="warning" plain round @tap="updateStatus(item, 'interviewing')">约面试</u-button>
          <u-button v-if="item.application.status !== 'offered' && item.application.status !== 'rejected'" size="mini" type="success" plain round @tap="updateStatus(item, 'offered')">录用</u-button>
          <u-button v-if="item.application.status !== 'rejected' && item.application.status !== 'withdrawn'" size="mini" type="danger" plain round @tap="updateStatus(item, 'rejected')">婉拒</u-button>
        </view>
      </view>
    </view>
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
const statusTag = (s) => ({ submitted: 'warning', viewed: 'primary', interviewing: 'warning', offered: 'success', rejected: 'danger', withdrawn: 'info' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return ''
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
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
.page-container { min-height: 100vh; background: var(--color-bg); padding-bottom: 20px; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 60px 0; }
.loading-text { font-size: 24rpx; color: var(--color-text-secondary); }
.empty-state { padding-top: 60px; }
.apl-list { padding: 8px 12px; }
.apl-card { background: var(--color-bg-card); border-radius: 12rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 1px 3px rgba(0,0,0,.05); }
.apl-row1 { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; }
.apl-main { display: flex; align-items: center; gap: 16rpx; min-width: 0; }
.apl-name { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.apl-phone { font-size: 24rpx; color: var(--color-primary); }
.apl-meta { font-size: 22rpx; color: var(--color-text-placeholder); margin-top: 8rpx; }
.resume-box { margin-top: 16rpx; padding: 20rpx; background: var(--color-bg); border-radius: 8rpx; display: flex; flex-direction: column; gap: 12rpx; }
.resume-line { display: flex; gap: 16rpx; font-size: 24rpx; color: var(--color-text); }
.resume-k { color: var(--color-text-secondary); flex-shrink: 0; width: 96rpx; }
.skill-tags { display: flex; flex-wrap: wrap; gap: 8rpx; }
.skill-tag { font-size: 20rpx; padding: 2rpx 12rpx; background: var(--color-primary-light); color: var(--color-primary); border-radius: 6rpx; }
.cert-img { width: 100%; height: 320rpx; background: #fff; border-radius: 8rpx; }
.apl-actions { display: flex; gap: 16rpx; margin-top: 16rpx; flex-wrap: wrap; }
</style>
