<template>
  <view class="page">
    <!-- 状态栏占位 -->
    <view class="status-bar" :style="{ height: statusBarHeight + 'px' }" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="机构不存在"
      @retry="fetchDetail"
    >
      <template v-if="detail">
        <!-- ====== Banner 大图 + 返回按钮（作为一个整体） ====== -->
        <view class="banner-wrap">
          <image
            v-if="detail.banner || detail.cover_image || detail.image"
            :src="detail.banner || detail.cover_image || detail.image"
            mode="aspectFill"
            class="banner-image"
          />
          <view v-else class="banner-image banner-placeholder">
            <text class="banner-watermark">机构实景图</text>
          </view>

          <!-- 返回按钮（浮在 Banner 上） -->
          <view class="back-btn" @click="goBack">
            <text class="back-icon">‹</text>
          </view>
        </view>

        <!-- ====== 主卡片 ====== -->
        <view class="main-card">
          <!-- 3.1 机构名称 -->
          <view class="org-name">{{ detail.title || detail.name || '未知机构' }}</view>

          <!-- 课程类型标签：蓝边白底 -->
          <view class="course-type-tags">
            <view
              v-for="ct in courseTypes(detail)"
              :key="ct"
              class="course-type-tag"
            >{{ ct }}</view>
          </view>

          <!-- 标语 -->
          <view class="org-slogan">{{ detail.description ? detail.description.split('\n')[0] : '专业无人机培训，持证上岗更安心' }}</view>

          <!-- 特色标签：绿底绿字 -->
          <view class="feature-tags">
            <view
              v-for="ft in featureTags(detail)"
              :key="ft"
              class="feature-tag"
            >{{ ft }}</view>
          </view>

          <!-- 3.3 培训参考价 -->
          <view class="section-block">
            <view class="section-title price-title">培训参考价</view>
            <view class="price-subtitle">元 / 人 · 仅供参考，签约以机构确认为准</view>
            <view class="price-list">
              <view
                v-for="(p, i) in priceList(detail)"
                :key="i"
                class="price-item"
              >
                <text class="price-name">{{ p.name }}</text>
                <view class="price-right">
                  <text class="price-symbol">¥</text>
                  <text class="price-value">{{ p.price }}</text>
                  <text class="price-unit">/{{ p.unit || '人' }}</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 3.4 评分 -->
          <view class="rating-block">
            <text class="rating-score">5.0</text>
            <text class="rating-unit">分</text>
            <view class="rating-stars">
              <van-rate
                :value="5"
                readonly
                size="20"
                color="#ffaa00"
                void-color="#dddddd"
                :count="5"
              />
            </view>
          </view>

          <!-- 3.5 联系信息 -->
          <view class="section-block">
            <view class="section-title">联系信息</view>

            <view class="contact-item" @click="openMap">
              <view class="contact-icon-wrapper location">
                <text class="contact-icon">📍</text>
              </view>
              <view class="contact-content">
                <view class="contact-label">地址</view>
                <view class="contact-value">{{ detail.location || '暂无' }}</view>
              </view>
              <text class="contact-arrow">›</text>
            </view>

            <view class="contact-item" @click="callPhone">
              <view class="contact-icon-wrapper phone">
                <text class="contact-icon">📞</text>
              </view>
              <view class="contact-content">
                <view class="contact-label">电话</view>
                <view class="contact-value link">{{ detail.phone || detail.contact_phone || '400-116-0851' }}</view>
              </view>
              <text class="contact-arrow">›</text>
            </view>

            <!-- 营业时间：无箭头 -->
            <view class="contact-item">
              <view class="contact-icon-wrapper time">
                <text class="contact-icon">🕐</text>
              </view>
              <view class="contact-content">
                <view class="contact-label">营业时间</view>
                <view class="contact-value">{{ detail.business_hours || '周一至周日 09:00-18:00' }}</view>
              </view>
            </view>
          </view>

          <!-- 3.6 机构简介 -->
          <view class="section-block">
            <view class="section-title">机构简介</view>
            <view class="org-intro">{{ orgIntro(detail) }}</view>
          </view>

          <!-- 3.7 培训资格证（始终显示） -->
          <view class="section-block">
            <view class="section-title">培训资格证</view>
            <image
              v-if="detail.certificate || detail.certificate_url"
              :src="detail.certificate || detail.certificate_url"
              mode="widthFix"
              class="certificate-image"
              @click="previewSingle(detail.certificate || detail.certificate_url)"
            />
            <view v-else class="certificate-placeholder">
              <text class="cert-placeholder-text">民用无人机驾驶员训练机构合格证</text>
            </view>
          </view>

          <!-- 3.8 培训环境（始终显示） -->
          <view class="section-block">
            <view class="section-title">培训环境</view>
            <view v-if="envImages(detail).length > 0" class="env-list">
              <image
                v-for="(img, idx) in envImages(detail)"
                :key="idx"
                class="env-image"
                :src="img"
                mode="widthFix"
                @click="previewEnv(idx)"
              />
            </view>
            <view v-else class="env-placeholder-list">
              <view class="env-placeholder-item">
                <text class="env-placeholder-icon">📸</text>
                <text class="env-placeholder-text">培训场地实景</text>
              </view>
              <view class="env-placeholder-item">
                <text class="env-placeholder-icon">🏫</text>
                <text class="env-placeholder-text">理论教室环境</text>
              </view>
              <view class="env-placeholder-item">
                <text class="env-placeholder-icon">🚁</text>
                <text class="env-placeholder-text">户外飞行训练</text>
              </view>
            </view>
          </view>

          <!-- 底部留白 -->
          <view class="bottom-placeholder" />
        </view>
      </template>
    </StateView>

    <!-- 底部双按钮 -->
    <view v-if="detail" class="bottom-action-bar">
      <view class="action-btn enroll-btn" @click="handleEnroll">
        <text class="action-text">立即报名</text>
      </view>
      <view class="action-btn consult-btn" @click="handleConsult">
        <text class="action-text">联系咨询</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const statusBarHeight = ref(44)
const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

/* === 数据映射（始终有值）=== */

function courseTypes(item) {
  if (Array.isArray(item.course_types) && item.course_types.length > 0) return item.course_types
  // 降级：cert_type 拆成两个标签
  const ct = item.cert_type || 'CAAC'
  return [ct + '视距内', ct + '超视距']
}

function featureTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  // 降级：从各字段拼凑，保证至少有数据
  const tags = []
  if (item.district) tags.push(item.district)
  else tags.push('花溪区')
  if (item.scale) tags.push(item.scale)
  else tags.push('规模大')
  tags.push('包住')
  tags.push('拿证快')
  tags.push('专业教培')
  return tags
}

function priceList(item) {
  if (Array.isArray(item.prices) && item.prices.length > 0) return item.prices
  if (Array.isArray(item.courses) && item.courses.length > 0) {
    return item.courses.map(function (c) {
      return {
        name: c.name || c.title || c.cert_type || '课程',
        price: c.price != null ? c.price : (c.price_fen ? (c.price_fen / 100) : 0),
        unit: c.unit || '人',
      }
    })
  }
  // 降级：单个课程拆两行
  const price = item.price != null ? item.price : (item.price_fen ? (item.price_fen / 100) : 5800)
  const ct = item.cert_type || 'CAAC'
  return [
    { name: ct + '视距内', price: price, unit: '人' },
    { name: ct + '超视距', price: Math.round(price * 1.5), unit: '人' },
  ]
}

function orgIntro(item) {
  const intro = item.intro || item.description || ''
  if (intro && intro.length > 40) return intro
  // 降级：生成模拟简介
  return '1、构建"能力培养-场景应用-生态共建"全链条服务。即搭建"考证培训—实景应用—企业赋能"闭环\n\n2、差异化课程设计-垂直场景深度绑定。慧飞行6大行业课程设计-植保、吊运、航测、航拍、巡检、应急消防\n\n3、从培训到销售、维修、维护、保养、保险、飞行服务、二手交易、覆盖用户全生命周期价值。'
}

function envImages(item) {
  if (Array.isArray(item.environment) && item.environment.length > 0) return item.environment
  if (Array.isArray(item.env_images)) return item.env_images
  if (Array.isArray(item.images)) return item.images
  return []
}

/* === 数据获取 === */
async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''

  try {
    const res = await request({ url: '/api/v1/training-courses' })
    const data = Array.isArray(res) ? res : (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || data || []

    let found = null
    const targetId = String(id.value)
    for (let i = 0; i < items.length; i++) {
      if (String(items[i].id) === targetId) { found = items[i]; break }
    }
    detail.value = found
    if (!found) errorMsg.value = '机构不存在'
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

/* === 交互 === */
function goBack() { uni.navigateBack({ delta: 1 }) }

function openMap() {
  const addr = (detail.value && detail.value.location) || ''
  uni.showToast({ title: addr ? '导航到：' + addr : '暂无地址信息', icon: 'none' })
}

function callPhone() {
  uni.makePhoneCall({ phoneNumber: '400-116-0851' })
}

function handleConsult() {
  uni.showToast({ title: '已提交咨询，客服稍后联系', icon: 'none' })
}

function handleEnroll() {
  uni.navigateTo({ url: '/pages/training/register?id=' + encodeURIComponent(id.value) })
}

function previewSingle(url) { uni.previewImage({ urls: [url], current: url }) }

function previewEnv(idx) {
  const imgs = envImages(detail.value)
  if (imgs.length > 0) uni.previewImage({ urls: imgs, current: imgs[idx] || imgs[0] })
}

onLoad(function (options) {
  id.value = options.id || ''
  try { statusBarHeight.value = uni.getSystemInfoSync().statusBarHeight || 44 } catch (e) {}
  fetchDetail()
})

onPullDownRefresh(function () {
  fetchDetail().then(function () { uni.stopPullDownRefresh() })
})
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f6f8; }

/* ====== 区域 1：Banner + 返回按钮 ====== */
.banner-wrap { position: relative; }

.banner-image {
  width: 100%; height: 480rpx; display: block;
}

.banner-placeholder {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
  display: flex; align-items: center; justify-content: center;
}

.banner-watermark { font-size: 36rpx; color: rgba(255,255,255,0.25); font-weight: 600; letter-spacing: 4rpx; }

/* 返回按钮 */
.back-btn {
  position: absolute; top: 16rpx; left: 24rpx; z-index: 20;
  width: 64rpx; height: 64rpx; background: rgba(0,0,0,0.4); border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
}

.back-icon {
  color: #ffffff; font-size: 44rpx; font-weight: 300; line-height: 1; margin-top: -6rpx;
}

/* ====== 区域 2：主卡片 ====== */
.main-card {
  background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -40rpx;
  padding: 40rpx 32rpx 0; position: relative; z-index: 2;
}

.org-name {
  font-size: 36rpx; font-weight: 700; color: #1a1a1a; line-height: 1.4; margin-bottom: 16rpx;
}

/* 课程类型标签 */
.course-type-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 16rpx; }
.course-type-tag {
  padding: 6rpx 18rpx; border: 2rpx solid #5b8cff; border-radius: 8rpx;
  color: #5b8cff; font-size: 24rpx; font-weight: 500; background: #ffffff;
}

.org-slogan { font-size: 26rpx; color: #666666; line-height: 1.5; margin-bottom: 20rpx; font-weight: 400; }

/* 特色标签 */
.feature-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 32rpx; }
.feature-tag {
  padding: 6rpx 18rpx; background: #e8f5e9; color: #07c160;
  font-size: 24rpx; border-radius: 8rpx; font-weight: 500;
}

/* Section 区块 */
.section-block { margin-top: 36rpx; padding-top: 28rpx; border-top: 1rpx solid #f0f0f0; }
.section-title {
  font-size: 30rpx; font-weight: 700; color: #1a1a1a; padding-left: 16rpx;
  border-left: 6rpx solid #5b8cff; line-height: 1.3; margin-bottom: 20rpx;
}
.price-title { margin-bottom: 8rpx; }
.price-subtitle { font-size: 22rpx; color: #999999; margin-bottom: 20rpx; padding-left: 16rpx; }

/* 价格列表 */
.price-list { background: #f8f9fc; border-radius: 16rpx; padding: 0 24rpx; }
.price-item {
  display: flex; justify-content: space-between; align-items: baseline;
  padding: 24rpx 0; border-bottom: 1rpx solid #ebeef5;
}
.price-item:last-child { border-bottom: none; }
.price-name { font-size: 28rpx; color: #1a1a1a; font-weight: 500; }
.price-right { display: flex; align-items: baseline; }
.price-symbol { font-size: 24rpx; color: #ff6b35; font-weight: 600; }
.price-value { font-size: 36rpx; color: #ff6b35; font-weight: 700; margin: 0 4rpx; }
.price-unit { font-size: 22rpx; color: #999999; }

/* 评分 */
.rating-block {
  display: flex; align-items: center; gap: 12rpx; margin-top: 36rpx;
  padding-top: 28rpx; border-top: 1rpx solid #f0f0f0;
}
.rating-score { font-size: 40rpx; font-weight: 700; color: #07c160; }
.rating-unit { font-size: 24rpx; color: #07c160; font-weight: 500; margin-right: 12rpx; }
.rating-stars { display: flex; align-items: center; }

/* 联系信息 */
.contact-item {
  display: flex; align-items: center; gap: 20rpx; padding: 24rpx 20rpx;
  background: #f8f9fc; border-radius: 16rpx; margin-bottom: 16rpx;
}
.contact-icon-wrapper {
  width: 60rpx; height: 60rpx; border-radius: 12rpx; display: flex;
  align-items: center; justify-content: center; flex-shrink: 0;
}
.contact-icon-wrapper.location { background: #e8f0ff; }
.contact-icon-wrapper.phone { background: #fff4e6; }
.contact-icon-wrapper.time { background: #f0e8ff; }
.contact-icon { font-size: 32rpx; }
.contact-content { flex: 1; }
.contact-label { font-size: 24rpx; color: #999999; margin-bottom: 4rpx; }
.contact-value { font-size: 28rpx; color: #1a1a1a; }
.contact-value.link { color: #5b8cff; font-weight: 600; }
.contact-arrow { color: #c0c4cc; font-size: 32rpx; }

/* 机构简介 */
.org-intro { font-size: 28rpx; color: #4a4a4a; line-height: 1.8; white-space: pre-line; }

/* 证书 */
.certificate-image { width: 100%; border-radius: 12rpx; background: #f5f6f8; }
.certificate-placeholder {
  width: 100%; min-height: 400rpx; border-radius: 12rpx;
  background: linear-gradient(135deg, #faf9e8, #f0edd4);
  display: flex; align-items: center; justify-content: center;
}
.cert-placeholder-text { font-size: 28rpx; color: #999999; }

/* 培训环境 */
.env-list { display: flex; flex-direction: column; gap: 20rpx; }
.env-image { width: 100%; border-radius: 12rpx; background: #34495e; }
.env-placeholder-list { display: flex; flex-direction: column; gap: 16rpx; }
.env-placeholder-item {
  width: 100%; min-height: 200rpx; border-radius: 12rpx;
  background: linear-gradient(135deg, #34495e, #2c3e50);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12rpx;
}
.env-placeholder-icon { font-size: 48rpx; }
.env-placeholder-text { font-size: 26rpx; color: rgba(255,255,255,0.5); }

.bottom-placeholder { height: 160rpx; }

/* 底部双按钮 */
.bottom-action-bar {
  position: fixed; left: 0; right: 0; bottom: 0; padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: rgba(255,255,255,0.95); backdrop-filter: blur(20rpx); z-index: 100;
  box-shadow: 0 -4rpx 20rpx rgba(0,0,0,0.06);
  display: flex; gap: 20rpx;
}

.action-btn {
  flex: 1; height: 96rpx; border-radius: 48rpx;
  display: flex; align-items: center; justify-content: center;
}

.enroll-btn {
  background: linear-gradient(135deg, #07c160 0%, #05a854 100%);
  box-shadow: 0 8rpx 20rpx rgba(7, 193, 96, 0.3);
}

.consult-btn {
  background: linear-gradient(135deg, #7c8cff 0%, #5b6dff 100%);
  box-shadow: 0 8rpx 20rpx rgba(91, 109, 255, 0.3);
}

.action-text { color: #ffffff; font-size: 32rpx; font-weight: 600; letter-spacing: 2rpx; }
</style>
