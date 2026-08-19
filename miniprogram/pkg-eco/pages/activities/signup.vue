<template>
  <view class="acts-page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar :title="done ? '报名结果' : '活动报名'" show-back :fixed="true" @back="goBack" />

    <!-- ═══ 成功态 ═══ -->
    <view v-if="done" class="succ">
      <view class="succ-ring">
        <u-icon name="success" size="40px" color="#fff" />
      </view>
      <text class="succ-title">报名成功</text>
      <text class="succ-desc">报名成功\n请保持手机畅通，活动开始前请留意消息提醒</text>
      <view class="receipt">
        <view class="rt">报名回执 <text class="rt-no">{{ receiptNo }}</text></view>
        <view class="rrow"><text>活动</text><text class="rv">{{ summary.title }}</text></view>
        <view class="rrow"><text>时间</text><text class="rv">{{ summary.time }}</text></view>
        <view class="rrow"><text>地点</text><text class="rv">{{ summary.loc }}</text></view>
        <view class="rrow"><text>状态</text><text class="rv cl-su">已报名</text></view>
      </view>
      <view class="btn-main" @tap="goList">返回活动列表</view>
      <view class="btn-ghost" @tap="onViewMine">查看我的报名</view>
    </view>

    <!-- ═══ 表单态 ═══ -->
    <template v-else>
      <!-- 活动摘要 -->
      <view v-if="act" class="summary">
        <text class="sum-title">{{ act.title }}</text>
        <view class="sum-meta">
          <text>{{ act.timeText }}</text>
          <text>{{ act.loc }}</text>
        </view>
      </view>

      <view class="form-card">
        <view class="form-group">
          <view class="form-label">姓名 <text class="required">*</text></view>
          <input class="form-input" v-model="form.name" placeholder="请输入报名人姓名" placeholder-class="ph" />
        </view>
        <view class="form-group">
          <view class="form-label">手机号 <text class="required">*</text></view>
          <input class="form-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号" placeholder-class="ph" />
          <text v-if="phoneError" class="form-error">请输入正确的手机号</text>
        </view>
        <view class="form-group">
          <view class="form-label">单位名称 <text class="required">*</text></view>
          <input class="form-input" v-model="form.company" placeholder="请输入所在单位（会员企业）" placeholder-class="ph" />
        </view>
        <view class="form-group">
          <view class="form-label">报名人数 <text class="required">*</text></view>
          <view class="stepper-row">
            <view class="step-btn" @tap="step(-1)">−</view>
            <text class="step-num">{{ form.persons }}</text>
            <view class="step-btn" @tap="step(1)">+</view>
            <text class="step-hint">每人最多 5 人</text>
          </view>
        </view>
        <view class="form-group">
          <view class="form-label">备注</view>
          <textarea class="form-textarea" v-model="form.remark" placeholder="如需乘车、住宿安排等请备注（选填）" placeholder-class="ph" :maxlength="200" />
        </view>
      </view>

      <view class="notice">
        <text class="notice-title">报名须知</text>
        <text class="notice-line">· 提交后即报名成功</text>
        <text class="notice-line">· 活动开始前请留意消息提醒</text>
        <text class="notice-line">· 如无法参加，请联系活动主办方</text>
      </view>

      <!-- 底部提交 -->
      <view class="submit-bar">
        <view class="submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
          {{ submitting ? '提交中...' : '确认报名' }}
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage, getStoredUser } from '@/utils/request'
import { dateOf, timeOf } from '@/utils/eventTime'

const done = ref(false)
const submitting = ref(false)
const act = ref(null)
const receiptNo = ref('')
const summary = ref({ title: '', time: '', loc: '' })
const form = ref({ name: '', phone: '', company: '', persons: 1, remark: '' })
const statusBarHeight = ref(20)
let actId = ''

const phoneError = computed(() => {
  const p = form.value.phone
  return !!p && !/^1[3-9]\d{9}$/.test(p)
})

const pad = (n) => (n < 10 ? '0' + n : '' + n)
const fmt = (key) => {
  if (!key) return ''
  const p = key.split('-')
  return p[1] + '-' + p[2]
}

async function fetchAct() {
  if (!actId) return
  try {
    // API 优先（实时数据）；接口未部署/网络失败时读列表页透传快照兜底
    let it = null
    try {
      const res = await request({ url: '/api/v1/events/' + encodeURIComponent(actId) })
      it = (res && res.data) || res
    } catch (e) { it = null }
    if (!it || !it.id) {
      try { it = uni.getStorageSync('act_detail_' + actId) || null } catch (e) { it = null }
    }
    if (it && it.id) {
      const rawTime = it.start_date || it.event_date || it.start_time || ''
      const date = dateOf(rawTime)
      const timeTxt = it.time_text || it.time || timeOf(rawTime)
      act.value = {
        title: it.title || '',
        timeText: (date ? fmt(date) + (timeTxt ? ' ' + timeTxt : '') : timeTxt) || '时间待定',
        loc: it.location || it.address || '地点待定',
      }
    }
  } catch (e) { /* 保持摘要空态 */ }
}

function step(delta) {
  const v = form.value.persons + delta
  form.value.persons = Math.max(1, Math.min(5, v))
}

async function handleSubmit() {
  if (submitting.value) return
  if (!form.value.name) return uni.showToast({ title: '请填写姓名', icon: 'none' })
  if (!/^1[3-9]\d{9}$/.test(form.value.phone)) return uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
  if (!form.value.company) return uni.showToast({ title: '请填写单位名称', icon: 'none' })

  submitting.value = true
  try {
    // 后端报名接口字段为 {name, phone, org}；人数/备注暂并入 org 展示（后端表无对应列）
    let org = form.value.company || ''
    if (form.value.persons > 1) org += '（' + form.value.persons + '人）'
    if (form.value.remark) org += '（备注：' + form.value.remark + '）'
    await request({
      url: '/api/v1/events/' + encodeURIComponent(actId) + '/register',
      method: 'POST',
      data: {
        name: form.value.name,
        phone: form.value.phone,
        org,
      },
    })
    const now = new Date()
    receiptNo.value = 'NO.' + now.getFullYear() + pad(now.getMonth() + 1) + pad(now.getDate()) + '-' + Math.floor(1000 + Math.random() * 9000)
    summary.value = act.value || { title: '', time: '', loc: '' }
    done.value = true
  } catch (e) {
    uni.showToast({ title: '报名提交失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

const onViewMine = () => {
  uni.showToast({ title: '可在「我的报名」中查看', icon: 'none', duration: 1500 })
}
const goList = () => {
  uni.redirectTo({ url: '/pkg-eco/pages/activities/list' })
}
const goBack = () => {
  if (done.value) uni.redirectTo({ url: '/pkg-eco/pages/activities/list' })
  else uni.navigateBack()
}

onLoad((options) => {
  actId = (options && options.id) || ''
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(() => uni.navigateTo({ url: '/pages/login/index' }), 500)
    return
  }
  const u = getStoredUser()
  if (u && u.phone) form.value.phone = u.phone
  if (u && u.company_name) form.value.company = u.company_name
  fetchAct()
})
</script>

<style scoped>
.acts-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 96px;
}

/* ===== 摘要 ===== */
.summary { background: #EAF3FB; border: 1px solid #D5E6F8; border-radius: 10px; padding: 14px; margin: 12px; }
.sum-title { display: block; font-size: 15px; font-weight: 700; color: #17212B; line-height: 1.4; }
.sum-meta { display: flex; gap: 12px; margin-top: 8px; font-size: 11.5px; color: #667085; flex-wrap: wrap; }

/* ===== 表单 ===== */
.form-card { background: #fff; border: 1px solid #EEF1F4; border-radius: 10px; margin: 0 12px 8px; padding: 16px; }
.form-group { margin-bottom: 16px; }
.form-group:last-child { margin-bottom: 0; }
.form-label { display: flex; align-items: center; font-size: 13px; font-weight: 600; color: #344054; margin-bottom: 6px; }
.required { color: #D92D20; font-size: 12px; margin-left: 2px; }
.form-input { width: 100%; box-sizing: border-box; min-height: 44px; padding: 10px 12px; border: 1px solid #E4E7EC; border-radius: 8px; font-size: 14px; color: #17212B; background: #fff; }
.form-textarea { width: 100%; box-sizing: border-box; min-height: 80px; padding: 10px 12px; border: 1px solid #E4E7EC; border-radius: 8px; font-size: 14px; color: #17212B; background: #fff; }
.ph { color: #98A2B3; }
.form-error { display: block; margin-top: 4px; font-size: 11px; color: #D92D20; }

/* ===== 人数步进器 ===== */
.stepper-row { display: flex; align-items: center; gap: 12px; }
.step-btn { width: 30px; height: 30px; border-radius: 50%; border: 1px solid #E4E7EC; background: #F4F6F8; color: #667085; font-size: 16px; display: flex; align-items: center; justify-content: center; }
.step-num { min-width: 24px; text-align: center; font-size: 15px; font-weight: 600; color: #17212B; }
.step-hint { font-size: 11.5px; color: #667085; margin-left: 4px; }

/* ===== 须知 ===== */
.notice { background: #fff; border: 1px solid #EEF1F4; border-radius: 10px; margin: 0 12px; padding: 12px 14px; }
.notice-title { display: block; font-size: 13px; font-weight: 600; color: #0A66C2; margin-bottom: 6px; }
.notice-line { display: block; font-size: 12px; color: #667085; line-height: 1.7; }

/* ===== 提交栏 ===== */
.submit-bar { position: fixed; left: 0; right: 0; bottom: 0; padding: 12px; padding-bottom: calc(12px + env(safe-area-inset-bottom)); background: #fff; border-top: 1px solid #EEF1F4; }
.submit-btn { height: 46px; border-radius: 8px; background: #0A66C2; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 15px; font-weight: 600; box-shadow: 0 2px 8px rgba(10,102,194,.35); }
.submit-btn.disabled { background: #98A2B3; box-shadow: none; }

/* ===== 成功态 ===== */
.succ { padding: 40px 32px; display: flex; flex-direction: column; align-items: center; text-align: center; }
.succ-ring { width: 88px; height: 88px; border-radius: 50%; background: linear-gradient(135deg, #34c759, #5bd88a); display: flex; align-items: center; justify-content: center; box-shadow: 0 14px 36px rgba(52,199,89,.35); margin-bottom: 22px; animation: pop .45s cubic-bezier(0.16, 1, 0.3, 1) both; }
@keyframes pop { from { transform: scale(.5); opacity: 0; } to { transform: scale(1); opacity: 1; } }
.succ-title { font-size: 21px; font-weight: 700; color: #17212B; margin-bottom: 8px; }
.succ-desc { font-size: 13.5px; color: #667085; line-height: 1.7; margin-bottom: 18px; white-space: pre-line; }
.receipt { width: 100%; background: #fff; border: 1px solid #EEF1F4; border-radius: 10px; padding: 16px; margin-bottom: 26px; text-align: left; box-sizing: border-box; }
.rt { font-size: 12px; color: #667085; margin-bottom: 10px; padding-bottom: 10px; border-bottom: 1px dashed #EBEDF0; display: flex; justify-content: space-between; }
.rt-no { color: #0A66C2; font-weight: 600; }
.rrow { display: flex; justify-content: space-between; font-size: 13px; padding: 5px 0; color: #667085; gap: 12px; }
.rv { color: #17212B; font-weight: 600; text-align: right; flex: 1; }
.cl-su { color: #168A55; }
.btn-main { width: 100%; height: 46px; border-radius: 8px; background: #0A66C2; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 15px; font-weight: 600; box-shadow: 0 2px 8px rgba(10,102,194,.35); }
.btn-ghost { width: 100%; height: 46px; border-radius: 8px; border: 1.5px solid #0A66C2; background: #fff; color: #0A66C2; display: flex; align-items: center; justify-content: center; font-size: 15px; font-weight: 600; margin-top: 10px; box-sizing: border-box; }
</style>
