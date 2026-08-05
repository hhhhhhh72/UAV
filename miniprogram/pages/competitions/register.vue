<template>
  <view class="page">
    <u-nav-bar title="赛事报名" show-back @back="goBack" />

    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !competition"
      empty-text="赛事不存在"
      @retry="loadCompetition"
    >
      <template v-if="competition">
        <!-- ① 赛事摘要 -->
        <view class="comp-card">
          <text class="comp-title">{{ competition.name || competition.title || '未知赛事' }}</text>
          <text v-if="competition.organizer || competition.sponsor" class="comp-org">{{ competition.organizer || competition.sponsor }}</text>
          <view v-if="compTags(competition).length > 0" class="comp-tags">
            <text v-for="t in compTags(competition)" :key="t" class="chip chip--tag">{{ t }}</text>
          </view>
          <view v-if="countdownText" class="comp-countdown">
            <text class="comp-countdown-text">{{ countdownText }}</text>
          </view>
        </view>

        <!-- ② 参赛项目选择 -->
        <view v-if="eventOptions.length > 0" class="event-picker" @click="showEventPicker = true">
          <view class="ep-bar" />
          <view class="ep-info">
            <text class="ep-label">选择参赛项目</text>
            <text class="ep-value">{{ selectedEvent }}</text>
          </view>
          <view class="ep-right">
            <view class="ep-price">
              <text class="ep-symbol">¥</text>
              <text class="ep-fee">{{ currentPrice.toLocaleString() }}</text>
            </view>
            <u-icon name="arrow" size="24rpx" color="#98A2B3" />
          </view>
        </view>

        <!-- ③ 报名条件提示 -->
        <view class="req-hint">
          <text class="req-hint-text">请先确认已满足赛事报名条件，再填写以下信息</text>
        </view>

        <!-- ④ 参赛人员信息 -->
        <view class="section">
          <text class="section-title">参赛人员信息</text>
          <view class="form-card">
            <view class="field">
              <text class="field-label">姓名<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.name" placeholder="请输入参赛人姓名" placeholder-class="field-ph" />
            </view>
            <view class="field">
              <text class="field-label">手机号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号码" placeholder-class="field-ph" />
            </view>
            <view class="field">
              <text class="field-label">身份证号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.idCard" maxlength="18" placeholder="请输入身份证号码" placeholder-class="field-ph" />
            </view>
            <template v-if="isTeamEvent">
              <view class="field">
                <text class="field-label">队伍名称<text class="field-star">*</text></text>
                <input class="field-input" v-model="form.team_name" placeholder="请输入队伍名称" placeholder-class="field-ph" />
              </view>
              <view class="field">
                <text class="field-label">队员人数<text class="field-star">*</text></text>
                <view class="stepper">
                  <view class="stepper-btn" :class="{ 'stepper-btn--disabled': form.member_count <= 2 }" @click="stepCount(-1)"><text>−</text></view>
                  <text class="stepper-val">{{ form.member_count }}</text>
                  <view class="stepper-btn" :class="{ 'stepper-btn--disabled': form.member_count >= 50 }" @click="stepCount(1)"><text>+</text></view>
                </view>
              </view>
            </template>
          </view>
        </view>

        <!-- ⑤ 证件上传 -->
        <view class="section">
          <text class="section-title">证件上传</text>
          <view class="upload-row">
            <view class="upload-box" @click="uploadImage('photo')">
              <image v-if="form.photo" :src="form.photo" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <view class="upload-icon"><text class="upload-icon-text">+</text></view>
                <text class="upload-title">白底免冠证件照</text>
                <text class="upload-hint">点击上传</text>
              </view>
              <view v-if="form.photo" class="upload-retag">重传</view>
            </view>
            <view class="upload-box" @click="uploadImage('idCard')">
              <image v-if="form.idCardImage" :src="form.idCardImage" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <view class="upload-icon"><text class="upload-icon-text">+</text></view>
                <text class="upload-title">身份证正面</text>
                <text class="upload-hint">点击上传</text>
              </view>
              <view v-if="form.idCardImage" class="upload-retag">重传</view>
            </view>
          </view>
        </view>

        <!-- ⑥ 声明 -->
        <view class="checkbox-row" @click="form.agreeHealth = !form.agreeHealth">
          <view class="checkbox-box" :class="{ 'checkbox-box--checked': form.agreeHealth }">
            <u-icon v-if="form.agreeHealth" name="check" size="24rpx" color="#ffffff" />
          </view>
          <text class="checkbox-text">本人身体健康，适合参加比赛</text>
        </view>
        <view class="checkbox-row" @click="form.agreeRules = !form.agreeRules">
          <view class="checkbox-box" :class="{ 'checkbox-box--checked': form.agreeRules }">
            <u-icon v-if="form.agreeRules" name="check" size="24rpx" color="#ffffff" />
          </view>
          <text class="checkbox-text">已阅读并同意比赛规则与免责声明</text>
        </view>

        <!-- ⑦ 费用 -->
        <view class="price-card">
          <view class="price-main">
            <text class="price-label">报名费用</text>
            <view class="price-amount">
              <text class="price-symbol">¥</text>
              <text class="price-value">{{ currentPrice.toLocaleString() }}</text>
              <text class="price-suffix">/人</text>
            </view>
            <text v-if="eventOptions.length > 0" class="price-sub">{{ eventOptions.length }} 项参赛项目可选</text>
          </view>
          <text v-if="countdownText" class="price-countdown">{{ countdownText }}</text>
        </view>

        <!-- ⑧ 提交 -->
        <view class="submit-btn" :class="{ 'submit-btn--loading': submitting }" @click="handleSubmit">
          <u-loading v-if="submitting" size="28rpx" color="#ffffff" />
          <text>{{ submitting ? '提交中...' : '确认报名' }}</text>
        </view>
        <text class="privacy-text">报名信息仅用于赛事注册，受隐私政策保护</text>
        <view class="bottom-space" />
      </template>
    </StateView>

    <!-- 参赛项目选择弹层 -->
    <u-picker
      :show="showEventPicker"
      :columns="eventLabels"
      :model-value="selectedEvent"
      title="选择参赛项目"
      @confirm="onEventConfirm"
      @cancel="onPickerCancel"
      @update:show="onPickerShow"
    />
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const competition = ref(null)
const submitting = ref(false)
const showEventPicker = ref(false)

const eventOptions = ref([])
const selectedEvent = ref('')
const currentPrice = ref(0)
const currentEventType = ref('')

const isTeamEvent = computed(function () {
  return currentEventType.value === '团体赛'
})

/* 倒计时 */
const countdownText = computed(function () {
  var d = competition.value && (competition.value.deadline || competition.value.enroll_deadline)
  if (!d) return ''
  var m = String(d).match(/(\d{4})[年.\-\/](\d{1,2})[月.\-\/](\d{1,2})/)
  if (!m) return ''
  var target = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  var days = Math.ceil((target - new Date()) / 86400000)
  if (days <= 0) return '报名已截止'
  return '报名截止倒计时 ' + days + ' 天'
})

const form = reactive({
  name: '', phone: '', idCard: '',
  team_name: '', member_count: 3,
  photo: '', idCardImage: '',
  agreeHealth: false, agreeRules: false,
})

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return []
}

function compFee(item) {
  if (item.fee != null) return item.fee
  if (item.price_fen != null) return item.price_fen / 100
  if (item.price != null) return item.price
  return 0
}

/* 项目选择（u-picker，columns 为字符串） */
const eventLabels = computed(function () {
  return eventOptions.value.map(function (o) { return o.label + ' · ¥' + o.fee })
})

function onPickerShow(v) {
  showEventPicker.value = v
}

function onPickerCancel() {
  showEventPicker.value = false
}

function onEventConfirm(label) {
  var idx = eventLabels.value.indexOf(label)
  var opt = idx >= 0 ? eventOptions.value[idx] : null
  if (!opt) { showEventPicker.value = false; return }
  selectedEvent.value = label
  currentPrice.value = opt.fee
  currentEventType.value = opt.type || ''
  if (currentEventType.value !== '团体赛') {
    form.member_count = 1
    form.team_name = ''
  } else {
    form.member_count = 3
  }
}

function stepCount(delta) {
  var next = form.member_count + delta
  if (next < 2 || next > 50) return
  form.member_count = next
}

function uploadImage(type) {
  uni.chooseImage({
    count: 1, sourceType: ['album', 'camera'],
    success: function (res) {
      form[type === 'photo' ? 'photo' : 'idCardImage'] = res.tempFilePaths[0]
    },
  })
}

/* 校验 */
function validate() {
  if (!form.name.trim()) return '请输入参赛人姓名'
  if (!/^1[3-9]\d{9}$/.test(form.phone.trim())) return '请输入正确的手机号'
  if (!/^\d{17}[\dXx]$/.test(form.idCard.trim())) return '请输入正确的身份证号'
  if (!form.photo) return '请上传白底免冠证件照'
  if (!form.idCardImage) return '请上传身份证正面'
  if (!form.agreeHealth) return '请确认健康状况'
  if (!form.agreeRules) return '请同意比赛规则与免责声明'
  return null
}

/* 提交 */
async function handleSubmit() {
  if (submitting.value) return
  var err = validate()
  if (err) { uni.showToast({ title: err, icon: 'none' }); return }

  submitting.value = true
  try {
    await request({
      url: '/api/v1/competitions/' + encodeURIComponent(id.value) + '/register',
      method: 'POST',
      data: {
        team_name: (form.team_name || form.name).trim(),
        member_count: form.member_count,
        contact_info: form.phone.trim(),
      },
    })
    uni.showToast({ title: '报名成功', icon: 'success' })
    setTimeout(function () { uni.navigateBack() }, 1500)
  } catch (e) {
    var msg = (e && e.data && e.data.message) || '报名失败，请重试'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    submitting.value = false
  }
}

function goBack() { uni.navigateBack({ delta: 1 }) }

/* 数据加载：仅真实接口 */
async function loadCompetition() {
  loading.value = true
  errorMsg.value = ''
  competition.value = null

  try {
    var res = await request({ url: '/api/v1/competitions' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    var found = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { found = items[i]; break }
    }
    competition.value = found
    if (found) loadEvents(found)
  } catch (e) {
    errorMsg.value = '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function loadEvents(item) {
  if (Array.isArray(item.events) && item.events.length > 0) {
    eventOptions.value = item.events.map(function (e) {
      return { label: e.name, fee: e.fee != null ? e.fee : 0, type: e.type || '个人赛' }
    })
  } else {
    eventOptions.value = []
  }

  if (eventOptions.value.length > 0) {
    var first = eventOptions.value[0]
    selectedEvent.value = eventLabels.value[0]
    currentPrice.value = first.fee
    currentEventType.value = first.type
  } else {
    selectedEvent.value = ''
    currentPrice.value = compFee(item)
    currentEventType.value = ''
  }

  if (currentEventType.value !== '团体赛') {
    form.member_count = 1
  }
}

onLoad(function (options) {
  id.value = options.id || ''
  loadCompetition()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ① 赛事摘要 */
.comp-card {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 28rpx;
}

.comp-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #17212B;
  line-height: 1.4;
  display: block;
  margin-bottom: 8rpx;
}

.comp-org {
  font-size: 24rpx;
  color: #667085;
  display: block;
  margin-bottom: 14rpx;
}

.comp-tags { display: flex; flex-wrap: wrap; gap: 10rpx; margin-bottom: 16rpx; }

.chip {
  display: inline-block;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  font-weight: 400;
  line-height: 1.6;
}

.chip--tag { background: #EAF3FB; color: #0A66C2; }

.comp-countdown {
  padding: 8rpx 14rpx;
  background: #FFF0E6;
  border-radius: 8rpx;
  display: inline-block;
}

.comp-countdown-text { font-size: 22rpx; color: #E96012; font-weight: 500; }

/* ② 参赛项目选择卡 */
.event-picker {
  margin: 20rpx 24rpx 0;
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 12rpx;
  padding: 24rpx;
  transition: opacity 160ms ease;
}

.event-picker:active { opacity: 0.85; }

.ep-bar {
  width: 6rpx;
  height: 56rpx;
  border-radius: 3rpx;
  background: #0A66C2;
  flex-shrink: 0;
}

.ep-info { flex: 1; min-width: 0; }

.ep-label { font-size: 22rpx; color: #667085; display: block; margin-bottom: 6rpx; }

.ep-value {
  font-size: 28rpx;
  font-weight: 500;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.ep-right { display: flex; align-items: center; gap: 12rpx; flex-shrink: 0; }

.ep-price { display: flex; align-items: baseline; }
.ep-symbol { font-size: 22rpx; color: #E96012; font-weight: 700; }
.ep-fee { font-size: 32rpx; font-weight: 700; color: #E96012; }

/* ③ 报名条件提示 */
.req-hint {
  margin: 20rpx 24rpx 0;
  background: #EAF3FB;
  border-radius: 8rpx;
  padding: 14rpx 20rpx;
}

.req-hint-text { font-size: 24rpx; color: #0A66C2; line-height: 1.5; }

/* ④ 表单 */
.section { margin: 28rpx 24rpx 0; }

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  padding-left: 16rpx;
  border-left: 6rpx solid #0A66C2;
  line-height: 1.3;
  margin-bottom: 16rpx;
  display: block;
}

.form-card {
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 8rpx 24rpx;
}

.field { padding: 20rpx 0; border-bottom: 1px solid #F4F6F8; }
.field:last-child { border-bottom: none; }

.field-label { font-size: 26rpx; color: #17212B; font-weight: 500; display: block; margin-bottom: 12rpx; }
.field-star { color: #E96012; margin-left: 4rpx; }

/* 输入控件：高 40-44px、圆角 6px、1px 边框 */
.field-input {
  height: 88rpx;
  box-sizing: border-box;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #17212B;
  width: 100%;
}

.field-input:focus { border-color: #0A66C2; }
.field-ph { color: #98A2B3; }

/* 队员人数步进器 */
.stepper { display: flex; align-items: center; gap: 24rpx; }

.stepper-btn {
  width: 76rpx;
  height: 76rpx;
  border: 1px solid #EEF1F4;
  border-radius: 12rpx;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  color: #17212B;
}

.stepper-btn:active { background: #EAF3FB; }
.stepper-btn--disabled { color: #98A2B3; }
.stepper-val { font-size: 30rpx; font-weight: 600; color: #17212B; min-width: 48rpx; text-align: center; }

/* ⑤ 证件上传 */
.upload-row { display: flex; gap: 20rpx; }

.upload-box {
  flex: 1;
  height: 240rpx;
  background: #ffffff;
  border: 2rpx dashed #D8DCE2;
  border-radius: 12rpx;
  overflow: hidden;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-preview { width: 100%; height: 100%; }

.upload-placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.upload-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 12rpx;
  background: #EAF3FB;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-icon-text { font-size: 40rpx; color: #0A66C2; font-weight: 400; line-height: 1; }
.upload-title { font-size: 24rpx; color: #344054; font-weight: 500; }
.upload-hint { font-size: 22rpx; color: #98A2B3; }

.upload-retag {
  position: absolute;
  right: 8rpx;
  bottom: 8rpx;
  padding: 4rpx 12rpx;
  background: rgba(23, 33, 43, 0.6);
  color: #ffffff;
  font-size: 20rpx;
  border-radius: 8rpx;
}

/* ⑥ 声明 */
.checkbox-row {
  margin: 0 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 16rpx 0;
}

.checkbox-box {
  width: 40rpx;
  height: 40rpx;
  border: 1px solid #D8DCE2;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #ffffff;
}

.checkbox-box--checked { background: #0A66C2; border-color: #0A66C2; }
.checkbox-text { font-size: 26rpx; color: #344054; }

/* ⑦ 费用 */
.price-card {
  margin: 8rpx 24rpx 0;
  background: #ffffff;
  border: 1px solid #EEF1F4;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.price-main { flex: 1; min-width: 0; }
.price-label { font-size: 22rpx; color: #667085; display: block; margin-bottom: 6rpx; }

.price-amount { display: flex; align-items: baseline; }
.price-symbol { font-size: 24rpx; color: #E96012; font-weight: 700; }
.price-value { font-size: 44rpx; font-weight: 700; color: #E96012; line-height: 1; }
.price-suffix { font-size: 22rpx; color: #98A2B3; margin-left: 4rpx; }

.price-sub { font-size: 22rpx; color: #98A2B3; display: block; margin-top: 6rpx; }

.price-countdown {
  padding: 8rpx 14rpx;
  background: #FFF0E6;
  color: #E96012;
  font-size: 22rpx;
  font-weight: 500;
  border-radius: 8rpx;
  flex-shrink: 0;
  margin-left: 16rpx;
}

/* ⑧ 提交按钮（高 44px、圆角 6px、单一主按钮） */
.submit-btn {
  margin: 28rpx 24rpx 0;
  height: 88rpx;
  background: #0A66C2;
  color: #ffffff;
  font-size: 32rpx;
  font-weight: 500;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  transition: opacity 160ms ease;
}

.submit-btn:active { opacity: 0.85; }
.submit-btn--loading { opacity: 0.75; }

.privacy-text { display: block; text-align: center; font-size: 22rpx; color: #98A2B3; margin: 16rpx 0 0; }
.bottom-space { height: calc(40rpx + env(safe-area-inset-bottom)); }
</style>
