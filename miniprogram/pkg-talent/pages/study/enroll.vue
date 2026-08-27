<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar :title="done ? '报名结果' : '研学报名'" show-back :fixed="true" @back="goBack" />

    <!-- 成功态 -->
    <view v-if="done" class="succ">
      <view class="succ-ring">
        <u-icon name="success" size="40px" color="#fff" />
      </view>
      <text class="succ-title">报名成功</text>
      <text class="succ-desc">协会审核通过后将与您电话确认行程\n可在「我的报名」查看报名状态</text>
      <view class="btn-main" @tap="goMine">查看我的报名</view>
      <view class="btn-ghost" @tap="backToDetail">返回研学详情</view>
    </view>

    <!-- 表单态 -->
    <template v-else>
      <view v-if="tour" class="summary">
        <text class="sum-title">{{ tour.title }}</text>
        <view class="sum-meta">
          <text>{{ tour.dateText }}</text>
          <text>{{ tour.location || '地点待定' }}</text>
        </view>
        <view class="sum-meta">
          <text v-if="tour.priceText">{{ tour.priceText }}</text>
          <text v-if="tour.capacity > 0">限 {{ tour.capacity }} 人</text>
        </view>
      </view>

      <view class="form-card">
        <view class="form-group">
          <view class="form-label">报名人姓名 <text class="required">*</text></view>
          <input class="form-input" v-model="form.name" placeholder="请输入联系人姓名" placeholder-class="ph" />
        </view>
        <view class="form-group">
          <view class="form-label">联系电话 <text class="required">*</text></view>
          <input class="form-input" v-model="form.phone" type="number" maxlength="11" placeholder="用于行程确认，11位手机号" placeholder-class="ph" />
        </view>
        <view class="form-row">
          <view class="form-group half">
            <view class="form-label">成人数 <text class="required">*</text></view>
            <view class="num-box">
              <view class="num-btn" @tap="bump('adult_count', -1)">−</view>
              <text class="num-val">{{ form.adult_count }}</text>
              <view class="num-btn" @tap="bump('adult_count', 1)">＋</view>
            </view>
          </view>
          <view class="form-group half">
            <view class="form-label">儿童数</view>
            <view class="num-box">
              <view class="num-btn" @tap="bump('child_count', -1)">−</view>
              <text class="num-val">{{ form.child_count }}</text>
              <view class="num-btn" @tap="bump('child_count', 1)">＋</view>
            </view>
          </view>
        </view>
        <view class="form-group">
          <view class="form-label">备注（选填）</view>
          <textarea class="form-textarea" v-model="form.remark" placeholder="饮食禁忌、航班/车次需求等" placeholder-class="ph" />
        </view>
      </view>

      <view class="submit-bar">
        <view class="btn-primary" hover-class="btn-hover" @tap="submit">{{ submitting ? '提交中...' : '确认报名' }}</view>
      </view>
      <text class="tip">报名后 24 小时内可联系协会取消；审核通过前不收取费用。</text>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, getErrorMessage } from '../../../utils/request'
import { safeBack, requireLogin } from '../../../utils/nav'

const statusBarHeight = ref(20)
const form = ref({ name: '', phone: '', adult_count: 1, child_count: 0, remark: '' })
const submitting = ref(false)
const done = ref(false)
const tour = ref(null)
const tourId = ref('')
const fromDetail = ref(false)

function bump(key, delta) {
  const v = form.value[key] + delta
  if (key === 'adult_count') form.value.adult_count = Math.max(1, Math.min(99, v))
  else form.value.child_count = Math.max(0, Math.min(99, v))
}

const fmtMoney = (fen) => {
  const yuan = (fen || 0) / 100
  if (yuan <= 0) return ''
  return yuan >= 10000 ? (yuan / 10000).toFixed(1) + ' 万元' : yuan + ' 元'
}

async function loadTour() {
  if (!tourId.value) return
  try {
    const res = await request({ url: '/api/v1/study/tours/' + encodeURIComponent(tourId.value) })
    const t = (res && res.data) || res
    if (t && t.id) {
      tour.value = {
        ...t,
        dateText: ((t.start_date || '').slice(0, 10)) + ((t.end_date && String(t.end_date).slice(0, 10) !== String(t.start_date).slice(0, 10)) ? ' ~ ' + String(t.end_date).slice(0, 10) : ''),
        priceText: fmtMoney(t.price_fen),
      }
    }
  } catch (e) { /* 摘要缺失不阻断报名（教学为按 id 提交） */ }
}

async function submit() {
  if (submitting.value) return
  if (!requireLogin()) return
  if (!form.value.name.trim()) return uni.showToast({ title: '请填写报名人姓名', icon: 'none' })
  if (!/^1[3-9]\d{9}$/.test(form.value.phone.trim())) return uni.showToast({ title: '请填写正确的11位手机号', icon: 'none' })
  submitting.value = true
  try {
    await request({
      url: '/api/v1/study/tours/' + encodeURIComponent(tourId.value) + '/enroll',
      method: 'POST',
      data: {
        name: form.value.name.trim(),
        phone: form.value.phone.trim(),
        adult_count: form.value.adult_count,
        child_count: form.value.child_count,
        remark: form.value.remark.trim(),
      },
    })
    done.value = true
  } catch (e) {
    uni.showToast({ title: getErrorMessage(e) || '报名失败，请重试', icon: 'none', duration: 2500 })
  } finally {
    submitting.value = false
  }
}

const goMine = () => {
  uni.redirectTo({ url: '/pkg-talent/pages/training/myenrollments?tab=study' })
}
const backToDetail = () => { safeBack() }
const goBack = () => { safeBack() }

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  tourId.value = options && options.id ? decodeURIComponent(options.id) : ''
  loadTour()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F5F6F8;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}
.summary {
  margin: 20rpx 28rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx;
  box-shadow: 0 4rpx 16rpx rgba(16, 24, 40, 0.06);
}
.sum-title { display: block; font-size: 34rpx; font-weight: 700; color: #17212B; }
.sum-meta { display: flex; flex-wrap: wrap; gap: 24rpx; margin-top: 14rpx; font-size: 26rpx; color: #667085; }
.form-card {
  margin: 0 28rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 8rpx 28rpx 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(16, 24, 40, 0.06);
}
.form-group { padding-top: 24rpx; }
.form-row { display: flex; gap: 24rpx; }
.form-group.half { flex: 1; min-width: 0; }
.form-label { font-size: 28rpx; font-weight: 600; color: #344054; margin-bottom: 14rpx; }
.required { color: #D92D20; }
.form-input {
  height: 88rpx;
  background: #FAFAFA;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}
.ph { color: #98A2B3; }
.form-textarea {
  width: 100%;
  height: 180rpx;
  background: #FAFAFA;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}
.num-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 88rpx;
  background: #FAFAFA;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  padding: 0 12rpx;
}
.num-btn {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #fff;
  border: 2rpx solid #E4E7EC;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  color: #0A66C2;
}
.num-btn:active { transform: scale(.92); }
.num-val { font-size: 32rpx; font-weight: 700; color: #17212B; }
.submit-bar {
  margin: 32rpx 28rpx 0;
}
.btn-primary {
  height: 92rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.btn-hover { opacity: .9; }
.tip {
  display: block;
  text-align: center;
  margin-top: 20rpx;
  font-size: 22rpx;
  color: #98A2B3;
}
.succ {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 60rpx 0;
}
.succ-ring {
  width: 128rpx;
  height: 128rpx;
  border-radius: 50%;
  background: #0B6B41;
  display: flex;
  align-items: center;
  justify-content: center;
}
.succ-title { margin-top: 28rpx; font-size: 36rpx; font-weight: 700; color: #17212B; }
.succ-desc {
  margin-top: 14rpx;
  font-size: 26rpx;
  color: #667085;
  text-align: center;
  line-height: 1.7;
}
.btn-main {
  margin-top: 60rpx;
  width: 100%;
  height: 92rpx;
  border-radius: 999rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 30rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.btn-ghost {
  margin-top: 20rpx;
  width: 100%;
  height: 88rpx;
  border-radius: 999rpx;
  border: 2rpx solid #C6CFDA;
  background: #fff;
  color: #344054;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
