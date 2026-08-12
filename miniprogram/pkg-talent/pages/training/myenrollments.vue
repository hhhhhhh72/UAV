<template>
  <view class="me-page">
    <u-nav-bar title="我的报名" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && enrollments.length === 0"
      empty-text="还没有报名记录"
      :show-create="!loading && !errorMsg"
      create-text="去逛逛培训课程"
      @retry="fetchList"
      @create="goCourses"
    >
      <view class="list-body">
        <view v-for="e in enrollments" :key="e.id" class="me-card">
          <!-- 顶部：课程名 + 状态 -->
          <view class="me-top">
            <text class="me-course">{{ e.course_title || '培训课程' }}</text>
            <text class="me-status" :class="'status--' + e.status">{{ statusLabel(e.status) }}</text>
          </view>

          <!-- 报名时间 -->
          <view class="me-time">报名时间 {{ dateText(e.created_at) }}</view>

          <!-- 报名人信息 -->
          <view class="me-meta">
            <text class="me-meta-item">报名人 {{ e.name || '—' }}</text>
            <text class="me-meta-item" v-if="e.phone">{{ e.phone }}</text>
          </view>
        </view>
      </view>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../../utils/request'
import { safeBack } from '../../../utils/nav'
import StateView from '../../../components/StateView.vue'

// 与后端 validEnrollmentStatus / 管理后台状态语义对齐（用户视角文案）
const STATUS_MAP = {
  pending: '待审核',
  approved: '已通过',
  paid: '已缴费',
  enrolled: '已报名',
  rejected: '已驳回',
}

const loading = ref(false)
const errorMsg = ref('')
const enrollments = ref([])

function statusLabel(s) { return STATUS_MAP[s] || s || '未知' }
function dateText(iso) { return iso ? String(iso).slice(0, 10) : '—' }

async function fetchList() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/enrollments/mine' })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    enrollments.value = Array.isArray(list) ? list : []
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goCourses() {
  uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
}

function goBack() {
  safeBack()
}

// onShow 而非 onLoad：报名提交返回后立即看到最新记录
onShow(() => fetchList())
onPullDownRefresh(() => {
  fetchList().finally(() => uni.stopPullDownRefresh())
})
</script>

<style scoped>
.me-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

.list-body {
  padding: 12px;
}

.me-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.me-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.me-course {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.me-status {
  flex-shrink: 0;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status--pending { color: #E96012; background: #FFF0E6; }
.status--approved { color: #168A55; background: #E9F7F0; }
.status--paid { color: #0A66C2; background: #EAF3FB; }
.status--enrolled { color: #168A55; background: #E9F7F0; }
.status--rejected { color: #D92D20; background: #FEF0EF; }

.me-time {
  margin-top: 8px;
  font-size: 11px;
  color: #98A2B3;
}

.me-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
}

.me-meta-item {
  font-size: 12px;
  color: #667085;
}
</style>
