<template>
  <view class="cd-page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="院校不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <view class="cd-content">
          <!-- 决策摘要：名称 / 类型 / 关键数据 -->
          <view class="summary-card">
            <view class="summary-tags">
              <text class="tag-level" :class="'tag-level--' + collegeLevel(detail)">{{ levelLabel(detail) }}</text>
              <text v-if="coopTypeLabel(detail.coop_type)" class="tag-chip tag-chip--blue">{{ coopTypeLabel(detail.coop_type) }}</text>
              <text v-for="t in compTags(detail)" :key="t" class="tag-chip tag-chip--blue">{{ t }}</text>
            </view>
            <text class="summary-name">{{ detail.name || detail.title || '未知院校' }}</text>
            <text class="summary-location">{{ locationText(detail) }}</text>
            <view class="summary-stats">
              <view v-for="s in statsData(detail)" :key="s.label" class="stat-block">
                <text class="stat-num">{{ s.value }}</text>
                <text class="stat-label">{{ s.label }}</text>
              </view>
            </view>
          </view>

          <!-- 院校简介 -->
          <view class="section-block">
            <text class="section-title">院校简介</text>
            <text class="intro-text">{{ detail.intro || detail.description || '暂无简介' }}</text>
          </view>

          <!-- 无人机相关专业 -->
          <view v-if="majorsList(detail).length > 0" class="section-block">
            <text class="section-title">无人机相关专业</text>
            <view class="major-list">
              <view v-for="m in majorsList(detail)" :key="m.name" class="major-item">
                <view class="major-info">
                  <text class="major-name">{{ m.name }}</text>
                  <text class="major-meta">{{ m.degree || '本科' }} · {{ m.duration || 4 }}年制{{ m.key ? ' · ' + m.key : '' }}</text>
                </view>
                <text v-if="m.flagship" class="flagship-tag">王牌</text>
              </view>
            </view>
          </view>

          <!-- 合作企业 -->
          <view v-if="partnerList(detail).length > 0" class="section-block">
            <text class="section-title">合作企业</text>
            <scroll-view scroll-x :show-scrollbar="false" class="partner-scroll">
              <view v-for="p in partnerList(detail)" :key="p.name" class="partner-card">
                <view class="partner-icon"><text class="partner-icon-text">{{ p.icon || '企' }}</text></view>
                <text class="partner-name">{{ p.name }}</text>
                <text class="partner-type">{{ p.type || '合作单位' }}</text>
              </view>
            </scroll-view>
          </view>

          <!-- 校园环境（图片 4:3） -->
          <view class="section-block">
            <text class="section-title">校园环境</text>
            <view v-if="photosList(detail).length > 0" class="photo-grid">
              <image
                v-for="(img, i) in photosList(detail)"
                :key="i"
                :src="img"
                mode="aspectFill"
                class="photo"
                @tap="previewPhotos(i)"
              />
            </view>
            <view v-else class="photo-empty">
              <text class="photo-empty-text">暂无校园环境图片</text>
            </view>
          </view>
        </view>

        <!-- 底部操作栏 -->
        <view class="sticky-bar">
          <view class="sb-btn sb-outline" hover-class="tap-fade" @tap="callPhone">拨打电话</view>
          <view class="sb-btn sb-primary" hover-class="tap-fade" @tap="openWebsite">访问官网</view>
        </view>
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

/* 院校层级：985/211=顶尖、本科、专科 */
function collegeLevel(item) {
  var tags = item.tags || []
  if (tags.indexOf('985') >= 0 || tags.indexOf('211') >= 0) return 'top'
  if (tags.indexOf('专科') >= 0 || tags.indexOf('高职') >= 0) return 'vocational'
  return 'undergraduate'
}

function levelLabel(item) {
  return { top: '985/211', undergraduate: '本科', vocational: '专科' }[collegeLevel(item)] || '本科'
}

/* 院校分域（科研合作/人才培养/综合） */
function coopTypeLabel(t) {
  return { research: '科研合作', talent: '人才培养', both: '综合' }[t] || ''
}

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags.slice(0, 3)
  if (Array.isArray(item.specialties)) return item.specialties.slice(0, 3)
  return []
}

function locationText(item) {
  return [item.city || '', item.levelTags || ''].filter(Boolean).join(' · ') || '暂无位置信息'
}

function statsData(item) {
  return [
    { label: '无人机专业', value: (item.majors && item.majors.length) || item.majorCount || item.major_count || 0 },
    { label: '合作企业', value: item.partnerCount || item.partner_count || 0 },
    { label: '在读学生', value: (item.studentCount || item.student_count || 0) + '+' },
    { label: '硕博导师', value: item.teacherCount || item.teacher_count || 0 },
  ]
}

function majorsList(item) {
  if (Array.isArray(item.majors) && item.majors.length > 0) return item.majors
  return []
}

function partnerList(item) {
  if (Array.isArray(item.partners) && item.partners.length > 0) return item.partners
  return []
}

function photosList(item) {
  if (Array.isArray(item.photos) && item.photos.length > 0) return item.photos
  return []
}

function previewPhotos(idx) {
  if (detail.value && Array.isArray(detail.value.photos)) {
    uni.previewImage({ urls: detail.value.photos, current: idx })
  }
}

function callPhone() {
  var phone = detail.value && detail.value.phone
  if (!phone) { uni.showToast({ title: '电话未录入', icon: 'none' }); return }
  uni.makePhoneCall({ phoneNumber: phone })
}

function openWebsite() {
  if (detail.value && detail.value.website) {
    uni.navigateTo({ url: '/pages/webview/index?url=' + encodeURIComponent(detail.value.website) })
  } else {
    uni.showToast({ title: '官网信息暂未录入', icon: 'none' })
  }
}

async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    var res = await request({ url: '/api/v1/colleges' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    detail.value = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { detail.value = items[i]; break }
    }
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.cd-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(74px + env(safe-area-inset-bottom));
}

.cd-content {
  padding: 12px;
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
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.tag-level,
.tag-chip {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag-level--top { background: #FFF0E6; color: #E96012; }
.tag-level--undergraduate { background: #EAF3FB; color: #0A66C2; }
.tag-level--vocational { background: #E9F7F0; color: #168A55; }

.tag-chip--blue {
  background: #EAF3FB;
  color: #0A66C2;
}

.summary-name {
  font-size: 19px;
  font-weight: 700;
  color: #17212B;
  line-height: 1.35;
  display: block;
}

.summary-location {
  font-size: 12px;
  color: #667085;
  display: block;
  margin-top: 4px;
}

.summary-stats {
  display: flex;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #EEF1F4;
}

.stat-block {
  flex: 1;
  text-align: center;
}

.stat-num {
  font-size: 17px;
  font-weight: 700;
  color: #17212B;
  display: block;
}

.stat-label {
  font-size: 10px;
  color: #98A2B3;
  display: block;
  margin-top: 2px;
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

.intro-text {
  font-size: 13px;
  color: #344054;
  line-height: 1.7;
  white-space: pre-line;
}

/* 专业 */
.major-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.major-item {
  padding: 10px 12px;
  background: #F4F6F8;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.major-info {
  flex: 1;
  min-width: 0;
}

.major-name {
  font-size: 13px;
  font-weight: 600;
  color: #17212B;
  display: block;
}

.major-meta {
  font-size: 11px;
  color: #667085;
  display: block;
  margin-top: 3px;
}

.flagship-tag {
  font-size: 10px;
  color: #E96012;
  background: #FFF0E6;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
  flex-shrink: 0;
}

/* 合作企业 */
.partner-scroll {
  display: flex;
  gap: 8px;
  white-space: nowrap;
  padding-bottom: 4px;
}

.partner-card {
  padding: 12px 14px;
  background: #F4F6F8;
  border-radius: 8px;
  text-align: center;
  flex-shrink: 0;
  min-width: 96px;
  display: inline-block;
}

.partner-icon {
  width: 44px;
  height: 44px;
  margin: 0 auto 6px;
  background: #EAF3FB;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.partner-icon-text {
  font-size: 18px;
  font-weight: 600;
  color: #0A66C2;
}

.partner-name {
  font-size: 12px;
  font-weight: 500;
  color: #17212B;
  display: block;
}

.partner-type {
  font-size: 10px;
  color: #98A2B3;
  display: block;
  margin-top: 2px;
}

/* 校园环境（4:3） */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
}

.photo {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 8px;
  background: #F4F6F8;
}

.photo-empty {
  aspect-ratio: 4 / 3;
  border-radius: 8px;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-empty-text {
  font-size: 12px;
  color: #98A2B3;
}

/* 底部操作栏 */
.sticky-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
  display: flex;
  gap: 10px;
}

.sb-btn {
  flex: 1;
  height: 46px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
}

.sb-primary {
  background: #0A66C2;
  color: #fff;
}

.sb-outline {
  background: #fff;
  color: #0A66C2;
  border: 1px solid #0A66C2;
}

.tap-fade {
  opacity: 0.7;
}
</style>
