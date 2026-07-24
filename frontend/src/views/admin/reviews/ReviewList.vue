<template>
  <div class="review-manage">
    <!-- 筛选区 -->
    <div class="filter-bar">
      <van-tabs v-model:active="statusTab" @change="onStatusChange">
        <van-tab title="全部" name="all" />
        <van-tab title="待审核" name="pending" />
        <van-tab title="已通过" name="approved" />
        <van-tab title="已拒绝" name="rejected" />
      </van-tabs>
      <div class="section-filter">
        <van-tag
          v-for="sec in sectionOptions"
          :key="sec.key"
          :type="sectionFilter === sec.key ? 'primary' : 'default'"
          size="medium"
          round
          @click="toggleSectionFilter(sec.key)"
        >
          {{ sec.label }}
        </van-tag>
      </div>
    </div>

    <!-- 评价列表 -->
    <div class="review-list" v-if="filteredReviews.length > 0">
      <div class="review-card" v-for="item in filteredReviews" :key="item.id">
        <div class="review-card-header">
          <div class="review-user">
            <van-image
              v-if="item.userAvatar"
              round
              width="28"
              height="28"
              :src="item.userAvatar"
              fit="cover"
            />
            <div v-else class="avatar-placeholder">
              <van-icon name="contact" size="16" color="#bdc3c7" />
            </div>
            <span class="user-name">{{ item.userName }}</span>
          </div>
          <div class="review-meta">
            <van-tag :type="sectionTagType(item.section)" size="small">{{ sectionLabel(item.section) }}</van-tag>
            <van-tag :type="statusTagType(item.status)" size="small" plain>{{ statusLabel(item.status) }}</van-tag>
          </div>
        </div>

        <div class="review-card-body">
          <van-rate v-model="item.rating" readonly size="14" color="#ffd21e" void-color="#e8e8e8" />
          <!-- 课程标签 -->
          <div class="admin-course-tag" v-if="item.courseName">
            <van-icon name="bookmark-o" size="12" />
            {{ item.courseName }}
          </div>
          <div class="review-text">{{ item.content }}</div>
          <!-- 评价图片 -->
          <div class="admin-review-images" v-if="item.images && item.images.length > 0">
            <van-image
              v-for="(img, imgIdx) in item.images"
              :key="imgIdx"
              width="60"
              height="60"
              radius="4"
              :src="img"
              fit="cover"
              @click="previewImage(item.images, imgIdx)"
            />
          </div>
          <div class="review-time">{{ formatTime(item.createTime) }}</div>
        </div>

        <div class="review-card-actions" v-if="item.status === 'pending'">
          <van-button size="small" type="success" plain @click="handleApprove(item)">通过</van-button>
          <van-button size="small" type="warning" plain @click="handleReject(item)">拒绝</van-button>
          <van-button size="small" type="danger" plain @click="handleDelete(item)">删除</van-button>
        </div>
        <div class="review-card-actions" v-else>
          <van-button size="small" type="danger" plain @click="handleDelete(item)">删除</van-button>
        </div>
      </div>
    </div>

    <van-empty v-else description="暂无评价数据" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showToast, showConfirmDialog, showImagePreview } from 'vant'
import axios from '@/utils/http'

const sectionOptions = [
  { key: 'all', label: '全部板块' },
  { key: 'yanxue', label: '研学' },
  { key: 'sale', label: '无人机销售' },
  { key: 'park', label: '乐园' }
]

const statusTab = ref('all')
const sectionFilter = ref('all')
const reviews = ref([])
const loading = ref(false)

const sectionLabel = (key) => {
  const map = { yanxue: '研学', sale: '无人机销售', park: '乐园' }
  return map[key] || key
}

const statusLabel = (status) => {
  const map = { pending: '待审核', approved: '已通过', rejected: '已拒绝' }
  return map[status] || status
}

const sectionTagType = (section) => {
  const map = { yanxue: 'primary', sale: 'success', park: 'warning' }
  return map[section] || 'default'
}

const statusTagType = (status) => {
  const map = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'default'
}

const filteredReviews = computed(() => {
  let list = reviews.value
  if (statusTab.value !== 'all') {
    list = list.filter(r => r.status === statusTab.value)
  }
  if (sectionFilter.value !== 'all') {
    list = list.filter(r => r.section === sectionFilter.value)
  }
  return list
})

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const previewImage = (images, startIdx) => {
  showImagePreview({ images, startPosition: startIdx })
}

const fetchReviews = async () => {
  loading.value = true
  try {
    const params = { limit: 200 }
    const res = await axios.get('/api/admin/reviews', { params })
    reviews.value = res.data?.data || []
  } catch (err) {
    showToast('获取评价数据失败')
  } finally {
    loading.value = false
  }
}

const onStatusChange = () => {
  // 前端筛选，无需重新请求
}

const toggleSectionFilter = (key) => {
  sectionFilter.value = sectionFilter.value === key ? 'all' : key
}

const updateReviewStatus = async (item, status) => {
  try {
    const res = await axios.post(`/api/admin/reviews/${item.id}`, { status })
    if (res.data?.success) {
      item.status = status
      item.reviewTime = new Date().toISOString()
      showToast(status === 'approved' ? '已通过' : '已拒绝')
    }
  } catch (err) {
    showToast('操作失败')
  }
}

const handleApprove = (item) => {
  updateReviewStatus(item, 'approved')
}

const handleReject = (item) => {
  updateReviewStatus(item, 'rejected')
}

const handleDelete = (item) => {
  showConfirmDialog({
    title: '确认删除',
    message: '删除后不可恢复，确定要删除这条评价吗？'
  }).then(async () => {
    try {
      const res = await axios.delete(`/api/admin/reviews/${item.id}`)
      if (res.data?.success) {
        reviews.value = reviews.value.filter(r => r.id !== item.id)
        showToast('已删除')
      }
    } catch (err) {
      showToast('删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchReviews()
})
</script>

<style scoped>
.review-manage {
  max-width: 800px;
  margin: 0 auto;
}

.filter-bar {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
  overflow: hidden;
}

.section-filter {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  flex-wrap: wrap;
}

.review-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.review-card {
  background: #fff;
  border-radius: 12px;
  padding: 12px;
}

.review-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  gap: 8px;
}

.review-user {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.avatar-placeholder {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #f2f2f7;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  font-size: 13px;
  font-weight: 500;
  color: #1d1d1f;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.review-meta {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.review-card-body {
  margin-bottom: 8px;
}

.admin-course-tag {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: #e8f2fc;
  color: #0071e3;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  margin-top: 6px;
}

.review-text {
  font-size: 13px;
  color: #333;
  line-height: 1.6;
  margin-top: 6px;
  word-break: break-all;
}

.admin-review-images {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}

.admin-review-images :deep(.van-image) {
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
}

.review-time {
  font-size: 11px;
  color: #86868b;
  margin-top: 4px;
}

.review-card-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 8px;
  border-top: 1px solid #f5f5f7;
}

@media (max-width: 400px) {
  .review-card-header {
    flex-direction: column;
  }
  .review-meta {
    align-self: flex-start;
  }
}
</style>
