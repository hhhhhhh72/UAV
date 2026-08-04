<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="院校不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ① 白色顶栏 -->
        <view class="top-bar">
          <view class="status-placeholder" :style="{ height: statusBarHeight + 'px' }" />
          <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
          <text class="top-title">院校详情</text>
        </view>

        <!-- ② 封面 + 悬浮头像 -->
        <view class="cover-wrap">
          <image v-if="detail.cover || detail.cover_image || detail.image" :src="detail.cover || detail.cover_image || detail.image" class="cover-img" mode="aspectFill" />
          <view v-else class="cover-placeholder"><text class="cover-emoji">校</text></view>
        </view>

        <view class="main-card">
          <view class="header-row">
            <view class="avatar">{{ initShort(detail) }}</view>
            <view class="header-info">
              <text class="college-name">{{ detail.name || detail.title || '未知院校' }}</text>
              <text class="college-location">{{ [detail.city || '', detail.levelTags || (detail.tags || []).join(' · ')].filter(Boolean).join(' · ') }}</text>
            </view>
          </view>

          <!-- ③ 标签 + 数据条 -->
          <view class="tag-row">
            <text v-if="coopTypeLabel(detail.coop_type)" class="tag-item coop-tag">{{ coopTypeLabel(detail.coop_type) }}</text>
            <text v-for="t in compTags(detail)" :key="t" class="tag-item" :style="tagStyle(t)">{{ t }}</text>
          </view>

          <view class="stats-row">
            <view class="stat" v-for="s in statsData(detail)" :key="s.label">
              <text class="stat-num">{{ s.value }}</text>
              <text class="stat-label">{{ s.label }}</text>
            </view>
          </view>

          <!-- ④ 院校简介 -->
          <view class="section-title">院校简介</view>
          <view class="intro-text">{{ detail.intro || detail.description || '暂无简介' }}</view>

          <!-- ⑤ 无人机相关专业 -->
          <view v-if="majorsList(detail).length > 0" class="section-block">
            <view class="section-title">无人机相关专业</view>
            <view class="major-list">
              <view v-for="m in majorsList(detail)" :key="m.name" class="major-item">
                <view>
                  <text class="major-name">{{ m.name }}</text>
                  <text class="major-meta">{{ m.degree || '本科' }} · {{ m.duration || 4 }}年制{{ m.key ? ' · ' + m.key : '' }}</text>
                </view>
                <text v-if="m.flagship" class="flagship-tag">王牌专业</text>
              </view>
            </view>
          </view>

          <!-- ⑥ 合作企业 -->
          <view v-if="partnerList(detail).length > 0" class="section-block">
            <view class="section-title">合作企业</view>
            <scroll-view class="partner-scroll" scroll-x :show-scrollbar="false">
              <view v-for="p in partnerList(detail)" :key="p.name" class="partner-card">
                <text class="partner-emoji">{{ p.icon || '企' }}</text>
                <text class="partner-name">{{ p.name }}</text>
                <text class="partner-type">{{ p.type || '合作单位' }}</text>
              </view>
            </scroll-view>
          </view>

          <!-- ⑦ 校园环境 -->
          <view class="section-block">
            <view class="section-title">校园环境</view>
            <view class="gallery-row">
              <view v-if="detail.photos && detail.photos.length > 0">
                <image v-for="(img, i) in detail.photos" :key="i" :src="img" class="gallery-img" mode="aspectFill" @click="previewPhotos(i)" />
              </view>
              <view v-else class="gallery-placeholder"><text class="placeholder-icon">景</text></view>
              <view v-if="!detail.photos || detail.photos.length === 0" class="gallery-placeholder"><text class="placeholder-icon">学</text></view>
            </view>
          </view>

          <!-- ⑧ 底部按钮 -->
          <view class="bottom-bar">
            <view class="btn-outline" @click="callPhone">联系电话</view>
            <view class="btn-primary" @click="openWebsite">访问官网</view>
          </view>
          <view class="bottom-spacer" />
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

const statusBarHeight = ref(44)
const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

function tagStyle(tag) {
  if (['博士点', '硕士点', '双一流', '985', '211'].indexOf(tag) >= 0) return { background: '#fff4e6', color: 'var(--color-warning)' }
  return { background: 'var(--color-primary-light)', color: 'var(--color-primary)' }
}

// 院校分域（功能方案修订版 三·五 分域：科研合作/人才培养/综合）
function coopTypeLabel(t) {
  return { research: '科研合作', talent: '人才培养', both: '综合' }[t] || ''
}
function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (Array.isArray(item.specialties)) return item.specialties
  return []
}

function initShort(item) {
  if (item.shortName) return item.shortName
  var name = item.name || ''
  return name.charAt(0) || '院'
}

function statsData(item) {
  return [
    { label: '无人机专业', value: item.majorCount || item.major_count || 0 },
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

function goBack() { uni.navigateBack({ delta: 1 }) }

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
  try { statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 44 } catch (e) {}
  loadDetail()
})
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ① 顶栏 */
.top-bar { background: #ffffff; padding: 0 32rpx 160rpx; }

.status-placeholder { width: 100%; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(10,102,194,0.08);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  margin-bottom: 24rpx;
}

.back-icon { color: var(--color-primary); font-size: 40rpx; font-weight: 300; }
.top-title { color: var(--color-primary); font-size: 28rpx; font-weight: 500; }

/* ② 封面 */
.cover-wrap {
  margin: -112rpx 24rpx 0; height: 440rpx; border-radius: 32rpx;
  overflow: hidden; position: relative; z-index: 1;
  box-shadow: 0 16rpx 48rpx rgba(10,102,194,0.3);
}

.cover-img { width: 100%; height: 100%; }
.cover-placeholder {
  width: 100%; height: 100%; background: linear-gradient(135deg, var(--color-primary), #1976d2);
  display: flex; align-items: center; justify-content: center;
}
.cover-emoji { font-size: 160rpx; opacity: 0.1; }

.main-card {
  background: #ffffff; border-radius: 48rpx 48rpx 0 0;
  margin-top: -48rpx; padding: 48rpx 32rpx 32rpx; position: relative; z-index: 1;
}

.header-row {
  display: flex; align-items: center; gap: 20rpx; margin-bottom: 16rpx;
  position: relative; z-index: 3;
}

.avatar {
  width: 120rpx; height: 120rpx; background: var(--color-primary); border-radius: 24rpx;
  display: flex; align-items: center; justify-content: center;
  color: #ffffff; font-size: 48rpx; font-weight: 600; flex-shrink: 0;
  box-shadow: 0 8rpx 32rpx rgba(10,102,194,0.3); margin-top: -80rpx;
  position: relative; z-index: 3;
}

.header-info { flex: 1; }
.college-name { font-size: 44rpx; font-weight: 700; color: var(--color-text); display: block; line-height: 1.3; }
.college-location { font-size: 26rpx; color: var(--color-primary); font-weight: 500; display: block; margin-top: 6rpx; }

/* ③ 标签 + 数据 */
.tag-row { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 28rpx; }
.tag-item { padding: 6rpx 18rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; }
.coop-tag { background: var(--color-primary-light); color: var(--color-primary); font-weight: 600; }

.stats-row { display: flex; gap: 12rpx; margin-bottom: 36rpx; }
.stat { flex: 1; padding: 24rpx 8rpx; background: #f8fafc; border-radius: 18rpx; text-align: center; }
.stat-num { font-size: 44rpx; font-weight: 700; color: var(--color-primary); display: block; }
.stat-label { font-size: 22rpx; color: var(--color-text-secondary); display: block; margin-top: 6rpx; }

/* Section */
.section-block { margin-top: 36rpx; }
.section-title {
  font-size: 30rpx; font-weight: 700; color: var(--color-text);
  padding-left: 20rpx; border-left: 6rpx solid var(--color-primary);
  line-height: 1.3; margin-bottom: 20rpx;
}

/* ④ 简介 */
.intro-text { font-size: 28rpx; color: #4a4a4a; line-height: 1.8; white-space: pre-line; margin-bottom: 36rpx; }

/* ⑤ 专业 */
.major-list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 36rpx; }
.major-item {
  padding: 24rpx; background: #f8fafc; border-radius: 18rpx;
  display: flex; justify-content: space-between; align-items: center;
}
.major-name { font-size: 28rpx; font-weight: 500; color: var(--color-text); display: block; }
.major-meta { font-size: 24rpx; color: var(--color-text-secondary); display: block; margin-top: 6rpx; }
.flagship-tag {
  font-size: 22rpx; color: var(--color-primary); background: var(--color-primary-light);
  padding: 6rpx 16rpx; border-radius: 10rpx; font-weight: 500; flex-shrink: 0;
}

/* ⑥ 合作企业 */
.partner-scroll { display: flex; gap: 16rpx; margin-bottom: 36rpx; white-space: nowrap; padding-bottom: 4rpx; }
.partner-card {
  padding: 24rpx 28rpx; background: #f8fafc; border-radius: 18rpx;
  text-align: center; flex-shrink: 0; min-width: 160rpx; display: inline-block;
}
.partner-emoji { font-size: 48rpx; display: block; margin-bottom: 8rpx; }
.partner-name { font-size: 26rpx; font-weight: 500; color: var(--color-text); display: block; }
.partner-type { font-size: 22rpx; color: var(--color-text-secondary); display: block; margin-top: 4rpx; }

/* ⑦ 校园环境 */
.gallery-row { display: flex; gap: 16rpx; margin-bottom: 40rpx; }
.gallery-img { flex: 1; height: 280rpx; border-radius: 18rpx; background: var(--color-bg); }
.gallery-placeholder { flex: 1; height: 280rpx; border-radius: 18rpx; background: linear-gradient(135deg, var(--color-primary), #1976d2); display: flex; align-items: center; justify-content: center; }
.placeholder-icon { font-size: 80rpx; opacity: 0.12; }

/* ⑧ 底部 */
.bottom-bar { display: flex; gap: 24rpx; border-top: 1rpx solid var(--color-divider); padding-top: 24rpx; }
.btn-outline { flex: 1; height: 88rpx; border-radius: 48rpx; border: 2rpx solid var(--color-primary); color: var(--color-primary); display: flex; align-items: center; justify-content: center; font-size: 28rpx; font-weight: 500; }
.btn-primary { flex: 1; height: 88rpx; border-radius: 48rpx; background: var(--color-primary); color: #ffffff; display: flex; align-items: center; justify-content: center; font-size: 28rpx; font-weight: 600; }
.bottom-spacer { height: calc(40rpx + env(safe-area-inset-bottom)); }
</style>
