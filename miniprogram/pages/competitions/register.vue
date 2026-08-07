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
        <!-- ① 赛事卡（作为首屏，替代原深蓝 Hero） -->
        <view class="event-card">
          <view class="event-card-strip" />
          <view class="event-card-header">
            <text class="event-card-title">{{ competition.name || competition.title }}</text>
          </view>
          <view class="event-card-tags">
            <text v-for="t in (competition.tags || compTags(competition))" :key="t" class="event-card-tag">{{ t }}</text>
          </view>
          <view class="event-card-meta">
            <text class="event-card-meta-text">🏛️ 主办：{{ competition.organizer || '待定' }}</text>
          </view>
          <view class="event-card-footer">
            <view class="event-card-price">
              <text class="event-card-symbol">¥</text>
              <text class="event-card-value">{{ currentPrice.toLocaleString() }}</text>
              <text class="event-card-unit">/人起</text>
            </view>
            <view v-if="countdownText" class="event-card-countdown">{{ countdownText }}</view>
          </view>
        </view>

        <!-- ② 主卡片 -->
        <view class="main-card">
          <!-- 参赛项目选择（色条+价格+箭头） -->
          <view class="event-picker" hover-class="press-feedback" :hover-stay-time="120" @click="showEventPicker = true">
            <view class="ep-bar" :style="{ background: eventBarColor }" />
            <view class="ep-info">
              <view class="ep-label-row">
                <text class="ep-label">选择参赛项目</text>
                <view class="ep-type-badge" :class="isTeamEvent ? 'ep-type-badge--team' : ''">{{ isTeamEvent ? '团体赛' : '个人赛' }}</view>
              </view>
              <text class="ep-value">{{ selectedEvent }}</text>
            </view>
            <view class="ep-right">
              <view class="ep-price">
                <text class="ep-symbol">¥</text>
                <text class="ep-fee">{{ currentPrice.toLocaleString() }}</text>
              </view>
              <text class="ep-arrow">▼</text>
            </view>
          </view>

          <!-- 报名条件提示条 -->
          <view class="req-hint">
            <view class="req-hint-bar" />
            <text class="req-hint-text">请确认已满足</text>
            <text class="req-hint-link" @click="goBack">报名条件</text>
            <text class="req-hint-text">，再填写表单</text>
          </view>

          <!-- ③ 个人信息表单 -->
          <view class="section-header">
            <view class="section-bar" />
            <text class="section-title">参赛人员信息</text>
            <text class="section-badge">必填</text>
          </view>

          <view class="form-group">
            <view class="field">
              <text class="field-label">姓名<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.name" placeholder="请输入参赛人姓名" />
            </view>
            <view class="field">
              <text class="field-label">手机号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号码" />
            </view>
            <view class="field">
              <text class="field-label">身份证号<text class="field-star">*</text></text>
              <input class="field-input" v-model="form.idCard" maxlength="18" placeholder="请输入身份证号码" />
            </view>
            <template v-if="isTeamEvent">
              <view class="field">
                <text class="field-label">队伍名称<text class="field-star">*</text></text>
                <input class="field-input" v-model="form.team_name" placeholder="请输入队伍名称" />
              </view>
              <view class="field">
                <text class="field-label">队员人数<text class="field-star">*</text></text>
                <view class="stepper-wrap">
                  <view class="stepper-btn" :class="{ disabled: form.member_count <= 2 }"
                    @click="form.member_count > 2 && form.member_count--"><text>−</text></view>
                  <text class="stepper-val">{{ form.member_count }}</text>
                  <view class="stepper-btn" :class="{ disabled: form.member_count >= 50 }"
                    @click="form.member_count < 50 && form.member_count++"><text>+</text></view>
                </view>
              </view>
            </template>
          </view>

          <!-- ④ 证件上传 -->
          <view class="section-header">
            <view class="section-bar" />
            <text class="section-title">证件上传</text>
            <text class="section-badge">必传</text>
          </view>

          <view class="upload-row">
            <view class="upload-box" @click="uploadImage('photo')">
              <image v-if="form.photo" :src="form.photo" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <view class="upload-cam"><text class="upload-cam-icon">＋</text></view>
                <text class="upload-title">白底免冠证件照</text>
                <text class="upload-hint">点击上传</text>
              </view>
              <view v-if="form.photo" class="upload-retag">重传</view>
            </view>
            <view class="upload-box" @click="uploadImage('idCard')">
              <image v-if="form.idCardImage" :src="form.idCardImage" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <view class="upload-cam"><text class="upload-cam-icon">＋</text></view>
                <text class="upload-title">身份证正面</text>
                <text class="upload-hint">点击上传</text>
              </view>
              <view v-if="form.idCardImage" class="upload-retag">重传</view>
            </view>
          </view>

          <!-- 声明 -->
          <view class="checkbox-row" @click="form.agreeHealth = !form.agreeHealth">
            <view class="checkbox-box" :class="{ checked: form.agreeHealth }">
              <text v-if="form.agreeHealth" class="check-mark">✓</text>
            </view>
            <text class="checkbox-text">本人身体健康，适合参加比赛</text>
          </view>
          <view class="checkbox-row" @click="form.agreeRules = !form.agreeRules">
            <view class="checkbox-box" :class="{ checked: form.agreeRules }">
              <text v-if="form.agreeRules" class="check-mark">✓</text>
            </view>
            <text class="checkbox-text">已阅读并同意<text class="link">比赛规则</text>与<text class="link">免责声明</text></text>
          </view>

          <!-- ⑤ 费用（集中 + 倒计时） -->
          <view class="price-section">
            <view class="price-main">
              <text class="price-label">报名费用</text>
              <view class="price-amount">
                <text class="price-symbol">¥</text>
                <text class="price-value">{{ currentPrice.toLocaleString() }}</text>
                <text class="price-suffix">/人</text>
              </view>
              <text class="price-sub">{{ eventOptions.length }} 项参赛项目可选</text>
            </view>
            <view v-if="countdownText" class="price-countdown">{{ countdownText }}</view>
          </view>

          <!-- ⑥ 底部 CTA -->
          <view class="bottom-bar">
            <view class="btn-outline" hover-class="press-feedback" :hover-stay-time="120" @click="handleConsult">联系咨询</view>
            <view class="btn-primary" :class="{ 'is-loading': submitting }" hover-class="press-feedback" :hover-stay-time="120" @click="handleSubmit">
              {{ submitting ? '提交中...' : '确认报名' }}<text v-if="!submitting" class="btn-arrow">→</text>
            </view>
          </view>
          <text class="privacy-text">报名信息仅用于赛事注册，受隐私政策保护</text>

          <view class="bottom-spacer" />
        </view>
      </template>
    </StateView>

    <!-- 参赛项目选择弹层 -->
    <u-popup :show="showEventPicker" position="bottom" round @close="showEventPicker = false">
      <view class="picker-panel">
        <view class="picker-bar">
          <text class="picker-btn" @click="showEventPicker = false">取消</text>
          <text class="picker-title">选择参赛项目</text>
          <text class="picker-btn picker-btn--confirm" @click="onEventConfirm">确定</text>
        </view>
        <picker-view :value="[eventIdx]" class="picker-view" @change="onEventChange">
          <picker-view-column>
            <view v-for="(item, i) in eventOptions" :key="i" class="picker-item">{{ item.label }}</view>
          </picker-view-column>
        </picker-view>
      </view>
    </u-popup>
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
const eventIdx = ref(0)
const selectedEvent = ref('')
const currentPrice = ref(0)
const currentEventType = ref('')

const isTeamEvent = computed(function () {
  return currentEventType.value === '团体赛'
})

/* 参赛项目色条：个人赛=橙、团体赛=紫 */
const eventBarColor = computed(function () {
  return isTeamEvent.value ? '#8B5CF6' : '#F97316'
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
  return ['多旋翼', '国家级']
}

/* 项目选择（picker-view） */
function onEventChange(e) {
  eventIdx.value = e.detail.value[0]
}

function onEventConfirm() {
  var opt = eventOptions.value[eventIdx.value]
  if (!opt) { showEventPicker.value = false; return }
  selectedEvent.value = opt.label
  currentPrice.value = opt.fee
  currentEventType.value = opt.type || ''
  if (currentEventType.value !== '团体赛') {
    form.member_count = 1
    form.team_name = ''
  } else {
    form.member_count = 3
  }
  showEventPicker.value = false
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

function handleConsult() { uni.showToast({ title: '咨询功能开发中', icon: 'none' }) }

function goBack() { uni.navigateBack({ delta: 1 }) }

/* 数据加载 */
async function loadCompetition() {
  loading.value = true
  errorMsg.value = ''

  try {
    var res = await request({ url: '/api/v1/competitions' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    var found = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { found = items[i]; break }
    }
    competition.value = found
    if (!found) useMock()
    loadEvents(found)
  } catch (e) {
    useMock()
  } finally {
    loading.value = false
  }
}

function useMock() {
  var mockMap = {
    'comp-1': {
      id: 'comp-1', name: '2026全国无人机职业技能大赛', title: '2026全国无人机职业技能大赛',
      organizer: '中国航空器拥有者及驾驶员协会',
      tags: ['多旋翼', '固定翼', '国家级'],
      deadline: '2026年9月1日',
      events: [
        { name: '多旋翼竞速赛', type: '个人赛', fee: 380 },
        { name: '固定翼编队赛', type: '团体赛', fee: 680 },
        { name: '航拍创作赛', type: '个人赛', fee: 280 },
      ],
    },
    'comp-2': {
      id: 'comp-2', name: '首届西南无人机FPV竞速挑战赛', title: '首届西南无人机FPV竞速挑战赛',
      organizer: '四川省航空运动协会',
      tags: ['竞速FPV', '多旋翼'],
      deadline: '2026年9月20日',
      events: [
        { name: 'FPV竞速赛', type: '个人赛', fee: 280 },
        { name: 'FPV花飞表演赛', type: '个人赛', fee: 0 },
      ],
    },
    'comp-3': {
      id: 'comp-3', name: '2026无人机创新应用大赛', title: '2026无人机创新应用大赛',
      organizer: '工信部人才交流中心',
      tags: ['航拍', '固定翼', '国家级'],
      deadline: '2026年7月20日',
      events: [
        { name: '航拍创作赛', type: '个人赛', fee: 0 },
        { name: '应急救援方案赛', type: '团体赛', fee: 0 },
        { name: '农业植保方案赛', type: '个人赛', fee: 0 },
      ],
    },
    'comp-4': {
      id: 'comp-4', name: '青少年无人机编程挑战赛', title: '青少年无人机编程挑战赛',
      organizer: '上海市教育委员会',
      tags: ['多旋翼', '航拍'],
      deadline: '2026年10月25日',
      events: [
        { name: '初级编程挑战', type: '个人赛', fee: 120 },
        { name: '高级编程挑战', type: '个人赛', fee: 120 },
      ],
    },
    'comp-5': {
      id: 'comp-5', name: '国际无人机系统博览会竞技赛', title: '国际无人机系统博览会竞技赛',
      organizer: '广州市低空经济产业协会',
      tags: ['多旋翼', '固定翼', '国际赛'],
      deadline: '2026年11月20日',
      events: [
        { name: '专业组竞速赛', type: '个人赛', fee: 580 },
        { name: '公开组竞速赛', type: '个人赛', fee: 380 },
        { name: '编队飞行赛', type: '团体赛', fee: 1200 },
      ],
    },
    'comp-6': {
      id: 'comp-6', name: '2026贵州无人机应急救援演练赛', title: '2026贵州无人机应急救援演练赛',
      organizer: '贵州省应急管理厅',
      tags: ['多旋翼', '航拍'],
      deadline: '2026年5月30日',
      events: [
        { name: '搜索定位赛', type: '团体赛', fee: 0 },
        { name: '物资投送赛', type: '团体赛', fee: 0 },
      ],
    },
  }
  competition.value = mockMap[id.value] || mockMap['comp-1']
  loadEvents(competition.value)
}

function loadEvents(item) {
  if (item && Array.isArray(item.events) && item.events.length > 0) {
    eventOptions.value = item.events.map(function (e) {
      return { label: e.name + ' · ' + (e.type || '个人赛'), fee: e.fee || 380, type: e.type || '个人赛' }
    })
  } else {
    eventOptions.value = [
      { label: '多旋翼竞速赛 · 个人赛', fee: 380, type: '个人赛' },
      { label: '固定翼编队赛 · 团体赛', fee: 680, type: '团体赛' },
      { label: '航拍创作赛 · 个人赛', fee: 280, type: '个人赛' },
    ]
  }
  selectedEvent.value = eventOptions.value[0].label
  currentPrice.value = eventOptions.value[0].fee
  currentEventType.value = eventOptions.value[0].type || ''
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
  --anim-fast: 160ms;
  --anim-base: 240ms;
  --anim-slow: 320ms;
  --ease-out: cubic-bezier(0.25, 0.46, 0.45, 0.94);
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f6f8 0%, #E8F2FC 100%);
}

/* ================================================================= */
/* ① 赛事卡                                                           */
/* ================================================================= */
/* ① 赛事卡（首屏，替代原深蓝 Hero）                                   */
/* ================================================================= */
.event-card {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border: 1rpx solid rgba(10, 31, 68, 0.06);
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(10, 31, 68, 0.06);
  position: relative;
  overflow: hidden;
}

.event-card-strip {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 6rpx;
  background: linear-gradient(180deg, #074D92, #0A66C2);
}

.event-card-header { margin-bottom: 12rpx; }

.event-card-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  line-height: 1.4;
  display: block;
}

.event-card-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 12rpx; }

.event-card-tag {
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 500;
  color: #0A66C2;
  background: rgba(10, 102, 194, 0.06);
  border: 1rpx solid rgba(10, 102, 194, 0.25);
}

.event-card-meta { margin-bottom: 16rpx; }
.event-card-meta-text { font-size: 26rpx; color: #969799; line-height: 1.5; }

.event-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 16rpx;
  border-top: 1rpx solid #ebedf0;
}

.event-card-price { display: flex; align-items: baseline; }
.event-card-symbol { font-size: 22rpx; color: #E96012; font-weight: 700; }
.event-card-value { font-size: 36rpx; color: #E96012; font-weight: 800; margin: 0 4rpx; }
.event-card-unit { font-size: 22rpx; color: #969799; }

.event-card-countdown {
  font-size: 22rpx;
  color: #EF4444;
  background: #FEE2E2;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
  font-weight: 500;
}

/* ================================================================= */
/* ② 主卡片                                                           */
/* ================================================================= */
.main-card {
  background: #ffffff;
  border-radius: 0;
  padding: 24rpx 24rpx 32rpx;
  position: relative;
  z-index: 2;
  animation: pageIn var(--anim-slow) var(--ease-out) both;
}

/* 参赛项目选择卡 */
.event-picker {
  display: flex;
  align-items: center;
  gap: 16rpx;
  background: #fafafa;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.ep-bar { width: 6rpx; height: 56rpx; border-radius: 3rpx; flex-shrink: 0; }

.ep-info { flex: 1; min-width: 0; }

.ep-label-row { display: flex; align-items: center; gap: 8rpx; margin-bottom: 6rpx; }
.ep-label { font-size: 22rpx; color: #969799; }

/* 参赛类型徽章：个人赛=蓝、团体赛=紫 */
.ep-type-badge {
  font-size: 18rpx;
  color: #0A66C2;
  background: rgba(10, 102, 194, 0.08);
  border: 1rpx solid rgba(10, 102, 194, 0.25);
  padding: 0 10rpx;
  border-radius: 999rpx;
  line-height: 1.6;
}

.ep-type-badge--team {
  color: #8B5CF6;
  background: rgba(139, 92, 246, 0.08);
  border-color: rgba(139, 92, 246, 0.25);
}

.ep-value {
  font-size: 30rpx;
  font-weight: 600;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.ep-right { display: flex; align-items: center; gap: 12rpx; flex-shrink: 0; }

.ep-price { display: flex; align-items: baseline; }

.ep-symbol { font-size: 22rpx; color: #E96012; font-weight: 700; }
.ep-fee { font-size: 34rpx; font-weight: 800; color: #E96012; }
.ep-arrow { font-size: 20rpx; color: #98A2B3; }

/* 报名条件提示条 */
.req-hint {
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: rgba(10, 102, 194, 0.06);
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 28rpx;
}

.req-hint-bar { width: 4rpx; height: 28rpx; background: #0A66C2; border-radius: 2rpx; flex-shrink: 0; }
.req-hint-text { font-size: 24rpx; color: #969799; }
.req-hint-link { font-size: 24rpx; color: #0A66C2; font-weight: 500; text-decoration: underline; }

/* ③ 表单 */
.section-header { display: flex; align-items: center; margin-bottom: 24rpx; }

.section-bar { width: 6rpx; height: 32rpx; background: #0A66C2; border-radius: 3rpx; margin-right: 12rpx; }
.section-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.section-badge {
  font-size: 20rpx;
  color: #EF4444;
  background: #FEE2E2;
  padding: 2rpx 12rpx;
  border-radius: 999rpx;
  margin-left: 12rpx;
  font-weight: 500;
}

.form-group { margin-bottom: 8rpx; }

.field { margin-bottom: 20rpx; }

.field-label { font-size: 26rpx; color: #17212B; font-weight: 500; display: block; margin-bottom: 10rpx; }
.field-star { color: #EF4444; margin-left: 4rpx; }

.field-input {
  background: #fafafa;
  border: 2rpx solid #ebedf0;
  border-radius: 24rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  color: #17212B;
  transition: border-color var(--anim-fast) ease, box-shadow var(--anim-fast) ease;
}

.field-input:focus {
  border-color: #0A66C2;
  box-shadow: 0 0 0 8rpx rgba(10, 102, 194, 0.12);
}

.stepper-wrap { display: flex; align-items: center; gap: 20rpx; }

.stepper-btn {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: #fafafa;
  border: 2rpx solid #ebedf0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  color: #17212B;
  font-weight: 500;
}

.stepper-btn.disabled { color: #CBD5E1; }
.stepper-val { font-size: 30rpx; font-weight: 600; color: #17212B; min-width: 40rpx; text-align: center; }

/* ④ 证件上传 */
.upload-row { display: flex; gap: 20rpx; margin-bottom: 20rpx; }

.upload-box {
  flex: 1;
  height: 240rpx;
  background: linear-gradient(180deg, #f5f6f8, #E8F2FC);
  border-radius: 16rpx;
  border: 2rpx dashed #CBD5E1;
  overflow: hidden;
  position: relative;
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

.upload-cam {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: #EEF3FA;
  border: 1rpx solid rgba(10, 102, 194, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-cam-icon { font-size: 32rpx; color: #0A66C2; font-weight: 300; }
.upload-title { font-size: 24rpx; color: #17212B; font-weight: 500; }
.upload-hint { font-size: 22rpx; color: #98A2B3; }

.upload-retag {
  position: absolute;
  right: 8rpx;
  bottom: 8rpx;
  padding: 4rpx 12rpx;
  background: rgba(10, 31, 68, 0.6);
  color: #ffffff;
  font-size: 20rpx;
  border-radius: 8rpx;
}

/* 声明 */
.checkbox-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 14rpx 0;
}

.checkbox-box {
  width: 40rpx;
  height: 40rpx;
  border: 2rpx solid #CBD5E1;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background var(--anim-fast) ease, border-color var(--anim-fast) ease;
}

.checkbox-box.checked { background: #0A66C2; border-color: #0A66C2; }
.check-mark { color: #ffffff; font-size: 26rpx; font-weight: 700; }
.checkbox-text { font-size: 26rpx; color: #969799; }
.link { color: #0A66C2; text-decoration: underline; }

/* ⑤ 费用 */
.price-section {
  background: #fafafa;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-top: 20rpx;
  margin-bottom: 24rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.price-label { font-size: 24rpx; color: #969799; display: block; margin-bottom: 6rpx; }

.price-amount { display: flex; align-items: baseline; }
.price-symbol { font-size: 24rpx; color: #E96012; font-weight: 700; }
.price-value { font-size: 44rpx; font-weight: 800; color: #E96012; line-height: 1; }
.price-suffix { font-size: 22rpx; color: #969799; margin-left: 4rpx; }

.price-sub { font-size: 22rpx; color: #98A2B3; display: block; margin-top: 6rpx; }

.price-countdown {
  padding: 8rpx 16rpx;
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
  font-size: 24rpx;
  font-weight: 600;
  border-radius: 999rpx;
  flex-shrink: 0;
}

/* ⑥ 底部按钮 */
.bottom-bar { display: flex; gap: 20rpx; }

.btn-outline {
  flex: 1;
  height: 96rpx;
  border-radius: 50rpx;
  border: 2rpx solid #0A66C2;
  background: #ffffff;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.btn-primary {
  flex: 1;
  height: 96rpx;
  border-radius: 50rpx;
  background: linear-gradient(135deg, #074D92, #0A66C2);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  box-shadow: 0 8rpx 24rpx rgba(10, 102, 194, 0.3);
  transition: transform var(--anim-fast) ease, opacity var(--anim-fast) ease;
}

.btn-primary.is-loading { opacity: 0.7; }
.btn-arrow { margin-left: 8rpx; font-size: 30rpx; }

.privacy-text { display: block; text-align: center; font-size: 22rpx; color: #98A2B3; margin-top: 16rpx; }
.bottom-spacer { height: calc(40rpx + env(safe-area-inset-bottom)); }

/* 项目选择弹层 */
.picker-panel { background: #ffffff; }
.picker-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #ebedf0;
}
.picker-btn { font-size: 28rpx; color: #969799; }
.picker-btn--confirm { color: #0A66C2; font-weight: 600; }
.picker-title { font-size: 30rpx; font-weight: 600; color: #17212B; }
.picker-view { height: 400rpx; }
.picker-item {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #17212B;
}

/* ================================================================= */
/* 动效                                                              */
/* ================================================================= */
@keyframes pageIn {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.2; }
  50%      { opacity: 0.8; }
}

.press-feedback {
  transform: scale(0.98);
  opacity: 0.92;
}

@media (prefers-reduced-motion: reduce) {
  .event-card, .main-card, .btn-primary, .btn-outline {
    animation: none !important;
    transition: none !important;
  }
}
</style>
