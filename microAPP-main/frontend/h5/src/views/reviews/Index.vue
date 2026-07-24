<template>
  <div class="reviews-page">
    <van-nav-bar title="服务评价" left-arrow @click-left="$router.back()" fixed placeholder />
    <HomeFloatButton />

    <van-tabs v-model:active="activeTab" sticky offset-top="46" @change="onTabChange">
      <van-tab v-for="tab in sections" :key="tab.key" :title="tab.label">
        <!-- 板块评分概要 -->
        <div class="section-summary" v-if="getSectionReviews(tab.key).length > 0">
          <div class="summary-score">{{ getAvgRating(tab.key) }}</div>
          <van-rate
            v-model="avgRatings[tab.key]"
            readonly
            allow-half
            size="14"
            color="#ffd21e"
            void-color="#e8e8e8"
          />
          <div class="summary-count">{{ getSectionReviews(tab.key).length }} 条评价</div>
        </div>

        <!-- 评价列表 -->
        <div class="review-list" v-if="getSectionReviews(tab.key).length > 0">
          <div class="review-card" v-for="item in getSectionReviews(tab.key)" :key="item.id">
            <div class="review-header">
              <van-image
                v-if="item.userAvatar && !item.isAnonymous"
                round
                width="32"
                height="32"
                :src="item.userAvatar"
                fit="cover"
              />
              <div v-else class="default-avatar-small">
                <van-icon name="user-circle-o" size="18" :color="item.isAnonymous ? '#969799' : '#bdc3c7'" />
              </div>
              <div class="review-user-info">
                <span class="review-user-name">
                  {{ item.isAnonymous ? '匿名用户' : item.userName }}
                  <van-tag v-if="item.isAnonymous" type="default" size="mini" class="anon-tag">匿名</van-tag>
                </span>
                <van-rate
                  v-model="item.rating"
                  readonly
                  size="10"
                  color="#ffd21e"
                  void-color="#e8e8e8"
                />
              </div>
              <span class="review-time">{{ formatTime(item.createTime) }}</span>
            </div>

            <!-- 课程标签 -->
            <div class="review-course-tag" v-if="item.courseName">
              <van-icon name="bookmark-o" size="12" color="#1989fa" />
              <span>{{ item.courseName }}</span>
            </div>

            <div class="review-content">{{ item.content }}</div>

            <!-- 图片展示 -->
            <div class="review-images" v-if="item.images && item.images.length > 0">
              <van-image
                v-for="(img, idx) in item.images"
                :key="idx"
                width="80"
                height="80"
                :src="img"
                fit="cover"
                radius="6"
                class="review-img-item"
                @click="previewImages(item.images, idx)"
              />
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <van-empty
          v-else
          image="search"
          description="暂无评价，快来发表第一条吧"
        />
      </van-tab>
    </van-tabs>

    <!-- 底部悬浮按钮 -->
    <div class="submit-btn-wrapper">
      <van-button type="primary" round block @click="openSubmitForm">我要评价</van-button>
    </div>

    <!-- 评价提交弹窗 -->
    <van-popup
      v-model:show="showForm"
      position="bottom"
      round
      :style="{ minHeight: '50%', maxHeight: '85%' }"
      closeable
    >
      <div class="form-popup">
        <div class="form-title">发表评价</div>
        <div class="form-section-label">
          当前板块：<van-tag type="primary">{{ currentSectionLabel }}</van-tag>
        </div>

        <van-form @submit="onSubmit" @failed="onFailed" ref="formRef">
          <!-- 课程选择 -->
          <van-field
            v-if="courseList.length > 0"
            :model-value="formData.courseName"
            is-link
            readonly
            label="选择课程"
            placeholder="请选择评价的课程"
            @click="showCoursePicker = true"
          />

          <div class="rating-field">
            <span class="rating-label">评分</span>
            <van-rate
              v-model="formData.rating"
              size="24"
              color="#ffd21e"
              void-color="#e8e8e8"
            />
            <span class="rating-text" v-if="formData.rating">{{ ratingTexts[formData.rating - 1] }}</span>
          </div>

          <van-field
            v-model="formData.content"
            type="textarea"
            placeholder="请输入您的评价内容..."
            rows="4"
            maxlength="500"
            show-word-limit
            :rules="[{ required: true, message: '请输入评价内容' }]"
          />

          <!-- 图片上传 -->
          <div class="upload-section">
            <div class="upload-label">上传图片（最多9张）</div>
            <van-uploader
              v-model="fileList"
              :max-count="9"
              :max-size="5 * 1024 * 1024"
              :after-read="afterReadFile"
              @oversize="onOversize"
              accept="image/*"
              multiple
            />
          </div>

          <div class="anonymous-switch">
            <span class="switch-label">匿名评价</span>
            <van-switch
              v-model="formData.isAnonymous"
              size="20"
              active-color="#07c160"
              inactive-color="#dcdee0"
            />
            <span class="switch-hint" v-if="!formData.isAnonymous">将显示您的用户名和头像</span>
            <span class="switch-hint" v-else>将以"匿名用户"身份发表</span>
          </div>

          <div class="form-actions">
            <van-button round block type="primary" native-type="submit" :loading="submitting">
              提交评价
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <!-- 课程选择器 -->
    <van-popup v-model:show="showCoursePicker" position="bottom" round>
      <van-picker
        :columns="courseColumns"
        @confirm="onCoursePicked"
        @cancel="showCoursePicker = false"
        title="选择课程"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showDialog, showImagePreview } from 'vant'
import axios from '@/utils/http'
import { authStorage } from '@/utils/http'
import HomeFloatButton from '@/components/HomeFloatButton.vue'

const router = useRouter()

const sections = [
  { key: 'yanxue', label: '研学' },
  { key: 'sale', label: '无人机销售' },
  { key: 'park', label: '乐园' }
]

const ratingTexts = ['很差', '较差', '一般', '较好', '非常好']

const activeTab = ref(0)
const allReviews = ref([])
const loading = ref(false)
const showForm = ref(false)
const submitting = ref(false)
const formRef = ref(null)

// 课程相关
const courseList = ref([])
const showCoursePicker = ref(false)
const courseColumns = computed(() => courseList.value.map(c => ({ text: c.name, value: c.id })))

// 图片上传
const fileList = ref([])

const formData = ref({
  rating: 5,
  content: '',
  isAnonymous: false,
  courseName: ''
})

const currentSectionKey = computed(() => sections[activeTab.value].key)
const currentSectionLabel = computed(() => sections[activeTab.value].label)

const avgRatings = ref({
  yanxue: 0,
  sale: 0,
  park: 0
})

const getSectionReviews = (sectionKey) => {
  return allReviews.value.filter(r => r.section === sectionKey)
}

const getAvgRating = (sectionKey) => {
  const reviews = getSectionReviews(sectionKey)
  if (reviews.length === 0) return '0.0'
  const sum = reviews.reduce((acc, r) => acc + r.rating, 0)
  const avg = sum / reviews.length
  avgRatings.value[sectionKey] = Math.round(avg * 10) / 10
  return avg.toFixed(1)
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${month}-${day}`
}

const previewImages = (images, startIdx) => {
  showImagePreview({
    images,
    startPosition: startIdx
  })
}

const fetchReviews = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/reviews')
    allReviews.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch reviews:', err)
  } finally {
    loading.value = false
  }
}

const fetchCourses = async (sectionKey) => {
  try {
    const res = await axios.get('/api/reviews/courses', { params: { section: sectionKey } })
    courseList.value = res.data?.data || []
  } catch {
    courseList.value = []
  }
}

const onTabChange = () => {
  // 切换tab时预加载课程列表
  fetchCourses(currentSectionKey.value)
}

const openSubmitForm = () => {
  const token = authStorage.getAccessToken()
  const userStr = localStorage.getItem('user')
  if (!token && !userStr) {
    showDialog({
      title: '提示',
      message: '请先登录后再发表评价',
      confirmButtonText: '去登录',
      showCancelButton: true,
      cancelButtonText: '取消'
    }).then(() => {
      router.push('/login')
    }).catch(() => {})
    return
  }
  formData.value = { rating: 5, content: '', isAnonymous: false, courseName: '' }
  fileList.value = []
  fetchCourses(currentSectionKey.value)
  showForm.value = true
}

const onCoursePicked = ({ selectedOptions }) => {
  if (selectedOptions && selectedOptions.length > 0) {
    formData.value.courseName = selectedOptions[0].text
  }
  showCoursePicker.value = false
}

// 图片上传处理
const afterReadFile = async (file) => {
  const files = Array.isArray(file) ? file : [file]
  for (const f of files) {
    f.status = 'uploading'
    f.message = '上传中...'
    try {
      const fd = new FormData()
      fd.append('file', f.file)
      const res = await axios.post('/api/upload', fd, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      if (res.data?.url) {
        f.status = 'done'
        f.message = ''
        f.url = res.data.url
      } else {
        f.status = 'failed'
        f.message = '上传失败'
      }
    } catch {
      f.status = 'failed'
      f.message = '上传失败'
    }
  }
}

const onOversize = () => {
  showToast('图片大小不能超过5MB')
}

const onSubmit = async () => {
  if (formData.value.rating < 1) {
    showToast('请选择评分')
    return
  }

  // 检查是否有正在上传的图片
  const uploading = fileList.value.some(f => f.status === 'uploading')
  if (uploading) {
    showToast('图片正在上传，请稍候')
    return
  }

  // 收集已上传成功的图片URL
  const uploadedImages = fileList.value
    .filter(f => f.status === 'done' && f.url)
    .map(f => f.url)

  submitting.value = true
  try {
    const res = await axios.post('/api/reviews', {
      section: currentSectionKey.value,
      rating: formData.value.rating,
      content: formData.value.content,
      isAnonymous: formData.value.isAnonymous,
      courseName: formData.value.courseName,
      images: uploadedImages
    })
    if (res.data?.success) {
      showToast('评价提交成功，等待审核')
      showForm.value = false
      fetchReviews()
    } else {
      showToast(res.data?.message || '提交失败')
    }
  } catch (err) {
    showToast(err.response?.data?.message || '提交失败，请重试')
  } finally {
    submitting.value = false
  }
}

const onFailed = () => {
  showToast('请填写评价内容')
}

onMounted(() => {
  fetchReviews()
  fetchCourses(currentSectionKey.value)
})
</script>

<style scoped>
.reviews-page {
  background: #f5f5f7;
  min-height: 100vh;
  width: 100%;
  max-width: var(--page-max-width);
  margin: 0 auto;
  padding-bottom: 72px;
}

/* 固定导航栏居中约束 */
.reviews-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

/* Sticky tabs 固定时居中约束 */
.reviews-page :deep(.van-sticky--fixed) {
  left: 50% !important;
  transform: translateX(-50%);
  width: 100% !important;
  max-width: var(--page-max-width);
}

.section-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #fff;
  margin: 10px 12px;
  border-radius: 10px;
}

.summary-score {
  font-size: 28px;
  font-weight: 700;
  color: #1d1d1f;
  line-height: 1;
}

.summary-count {
  font-size: 12px;
  color: #86868b;
  margin-left: auto;
  white-space: nowrap;
}

.review-list {
  padding: 0 12px;
}

.review-card {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  margin-bottom: 8px;
}

.review-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.default-avatar-small {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f2f2f7;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.review-user-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.review-user-name {
  font-size: 13px;
  font-weight: 500;
  color: #1d1d1f;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}

.anon-tag {
  flex-shrink: 0;
  margin-left: 4px;
}

.review-time {
  font-size: 11px;
  color: #86868b;
  flex-shrink: 0;
}

.review-course-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: #e8f4ff;
  border-radius: 4px;
  font-size: 12px;
  color: #1989fa;
  margin-bottom: 6px;
}

.review-content {
  font-size: 13px;
  color: #333;
  line-height: 1.6;
  word-break: break-all;
}

.review-images {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.review-img-item {
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
}

/* 底部悬浮提交按钮 */
.submit-btn-wrapper {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);
  padding: 10px 16px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -1px 8px rgba(0, 0, 0, 0.06);
  z-index: 10;
}

/* 评价弹窗 */
.form-popup {
  padding: 16px;
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
  max-height: 80vh;
  overflow-y: auto;
}

.form-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d1d1f;
  text-align: center;
  margin-bottom: 12px;
}

.form-section-label {
  font-size: 13px;
  color: #666;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.rating-field {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  margin-bottom: 4px;
}

.rating-label {
  font-size: 14px;
  color: #646566;
  flex-shrink: 0;
}

.rating-text {
  font-size: 12px;
  color: #ffd21e;
  flex-shrink: 0;
}

.upload-section {
  padding: 10px 16px;
}

.upload-label {
  font-size: 13px;
  color: #646566;
  margin-bottom: 8px;
}

.anonymous-switch {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin-top: 8px;
  background: #f7f8fa;
  border-radius: 8px;
}

.switch-label {
  font-size: 14px;
  color: #323233;
  flex-shrink: 0;
}

.switch-hint {
  font-size: 12px;
  color: #969799;
  margin-left: auto;
}

.form-actions {
  padding: 12px 0 0;
}
</style>
