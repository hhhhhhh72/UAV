<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">赛事报名</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-state">
      <u-loading size="28rpx" />
      <text>加载中...</text>
    </view>

    <!-- 加载失败 -->
    <view v-else-if="errorMsg" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">加载失败</view>
      <view class="pub-empty-desc">请检查网络后重试</view>
      <view class="pub-btn pub-btn--primary retry-btn" hover-class="pub-btn--active" @tap="loadCompetition">重新加载</view>
    </view>

    <!-- 赛事不存在 -->
    <view v-else-if="!competition" class="pub-empty">
      <view class="pub-empty-title">赛事不存在</view>
    </view>

    <!-- 表单 -->
    <template v-else>
      <!-- ① 赛事卡（首屏） -->
      <view class="event-card">
        <view class="event-card-strip" />
        <view class="event-card-title">{{ competition.name || competition.title }}</view>
        <view class="event-card-tags">
          <text v-for="t in (competition.tags || compTags(competition))" :key="t" class="event-card-tag">{{ t }}</text>
        </view>
        <view class="event-card-meta">
          <text class="event-card-meta-text">主办：{{ competition.organizer || competition.sponsor || '待定' }}</text>
        </view>
        <view class="event-card-footer">
          <view class="event-card-price">
            <text v-if="currentPrice != null" class="event-card-symbol">¥</text>
            <text class="event-card-value">{{ priceText }}</text>
            <text v-if="currentPrice != null" class="event-card-unit">/人起</text>
          </view>
          <view v-if="countdownText" class="event-card-countdown">{{ countdownText }}</view>
        </view>
      </view>

      <!-- ② 参赛项目选择 -->
      <view class="pub-section">
        <view class="pub-section-title">参赛项目</view>
        <view class="pub-form-card">
          <view class="pub-field" hover-class="pub-fade" @tap="showEventPicker = true">
            <view class="pub-field-label">选择参赛项目</view>
            <view class="pub-select-field">
              <view class="ep-left">
                <view class="ep-bar" :style="{ background: eventBarColor }" />
                <view class="ep-info">
                  <view class="ep-label-row">
                    <text class="ep-value">{{ selectedEvent }}</text>
                    <view class="ep-type-badge" :class="isTeamEvent ? 'ep-type-badge--team' : ''">{{ isTeamEvent ? '团体赛' : '个人赛' }}</view>
                  </view>
                  <view class="ep-price">
                    <text v-if="currentPrice != null" class="ep-symbol">¥</text>
                    <text class="ep-fee">{{ priceText }}</text>
                  </view>
                </view>
              </view>
              <text class="pub-arrow">›</text>
            </view>
          </view>
          <!-- 报名条件提示条 -->
          <view class="req-hint">
            <text class="req-hint-text">请确认已满足</text>
            <text class="req-hint-link" @tap="goBack">报名条件</text>
            <text class="req-hint-text">，再填写表单</text>
          </view>
          <!-- 项目选择说明：后端报名契约不收 event，选择仅作费用展示，最终以现场确认为准 -->
          <view class="event-confirm-note">所选参赛项目仅作费用参考展示，最终参赛项目以现场确认为准</view>
        </view>
      </view>

      <!-- ③ 参赛人员信息 -->
      <view class="pub-section">
        <view class="pub-section-title">参赛人员信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">姓名<text class="pub-required">*</text></view>
            <input class="pub-input" v-model="form.name" placeholder="请输入参赛人姓名" placeholder-class="pub-placeholder" />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">手机号<text class="pub-required">*</text></view>
            <input class="pub-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号码" placeholder-class="pub-placeholder" />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">身份证号<text class="pub-required">*</text></view>
            <input class="pub-input" v-model="form.idCard" maxlength="18" placeholder="请输入身份证号码" placeholder-class="pub-placeholder" />
          </view>
          <template v-if="isTeamEvent">
            <view class="pub-field">
              <view class="pub-field-label">队伍名称<text class="pub-required">*</text></view>
              <input class="pub-input" v-model="form.team_name" placeholder="请输入队伍名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">队员人数<text class="pub-required">*</text></view>
              <view class="stepper-wrap">
                <view class="stepper-btn" :class="{ disabled: form.member_count <= 2 }"
                  @tap="form.member_count > 2 && form.member_count--"><text>−</text></view>
                <text class="stepper-val">{{ form.member_count }}</text>
                <view class="stepper-btn" :class="{ disabled: form.member_count >= 50 }"
                  @tap="form.member_count < 50 && form.member_count++"><text>+</text></view>
              </view>
            </view>
          </template>
        </view>
      </view>

      <!-- ④ 证件上传 -->
      <view class="pub-section">
        <view class="pub-section-title">证件上传</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">白底免冠证件照<text class="pub-required">*</text></view>
            <view class="pub-upload-row reg-upload-row">
              <view v-if="form.photo" class="pub-photo" @tap="uploadImage('photo')">
                <image :src="form.photo" class="pub-photo-img" mode="aspectFill" />
                <view class="pub-photo-tag">重传</view>
              </view>
              <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="uploadImage('photo')">＋</view>
            </view>
          </view>
          <view class="pub-field">
            <view class="pub-field-label">身份证正面<text class="pub-required">*</text></view>
            <view class="pub-upload-row reg-upload-row">
              <view v-if="form.idCardImage" class="pub-photo" @tap="uploadImage('idCard')">
                <image :src="form.idCardImage" class="pub-photo-img" mode="aspectFill" />
                <view class="pub-photo-tag">重传</view>
              </view>
              <view v-else class="pub-add-photo" hover-class="pub-fade" @tap="uploadImage('idCard')">＋</view>
            </view>
          </view>
          <view class="pub-upload-tip">白底免冠证件照与身份证正面共 2 张必传，点击已传图片可重新选择</view>
        </view>
      </view>

      <!-- 声明 -->
      <view class="pub-section">
        <view class="pub-form-card">
          <view class="agree-row" @tap="form.agreeHealth = !form.agreeHealth">
            <view class="agree-box" :class="{ checked: form.agreeHealth }">
              <text v-if="form.agreeHealth" class="agree-check">✓</text>
            </view>
            <text class="agree-text">本人身体健康，适合参加比赛</text>
          </view>
          <view class="agree-row" @tap="form.agreeRules = !form.agreeRules">
            <view class="agree-box" :class="{ checked: form.agreeRules }">
              <text v-if="form.agreeRules" class="agree-check">✓</text>
            </view>
            <text class="agree-text">已阅读并同意<text class="agree-link">比赛规则</text>与<text class="agree-link">免责声明</text></text>
          </view>
        </view>
      </view>

      <!-- ⑤ 费用（集中 + 倒计时） -->
      <view class="pub-section">
        <view class="pub-form-card">
          <view class="fee-row">
            <view class="fee-main">
              <text class="fee-label">报名费用</text>
              <view class="fee-amount">
                <text v-if="currentPrice != null" class="fee-symbol">¥</text>
                <text class="fee-value">{{ priceText }}</text>
                <text v-if="currentPrice != null" class="fee-suffix">/人</text>
              </view>
              <text class="fee-sub">{{ eventOptions.length }} 项参赛项目可选</text>
            </view>
            <view v-if="countdownText" class="fee-countdown">{{ countdownText }}</view>
          </view>
        </view>
      </view>

      <text class="privacy-text">报名信息仅用于赛事注册，受隐私政策保护</text>
    </template>

    <!-- ⑥ 固定底部操作区（双 CTA：次白描边 + 主蓝） -->
    <view v-if="competitionClosed" class="pub-closed-banner">该赛事报名已截止，暂不能提交报名</view>
    <view v-if="!loading && !errorMsg && competition" class="pub-sticky">
      <view class="pub-btn pub-btn--secondary" hover-class="pub-btn--active" @tap="handleConsult">联系咨询</view>
      <view class="pub-btn pub-btn--primary" :class="{ 'is-loading': submitting, 'is-disabled': competitionClosed }" hover-class="pub-btn--active" @tap="handleSubmit">
        {{ submitting ? '提交中...' : (competitionClosed ? '已截止' : '确认报名') }}<text v-if="!submitting && !competitionClosed" class="btn-arrow">→</text>
      </view>
    </view>

    <!-- 参赛项目选择弹层（picker-view，外壳对齐 pub-sheet；mask 用 catchtouchmove 拦截穿透，避免选完回不了顶） -->
    <view v-if="showEventPicker" class="pub-overlay" catchtouchmove="noop" @tap="showEventPicker = false">
      <view class="pub-sheet pub-sheet--picker" @tap.stop>
        <view class="pub-grab"></view>
        <view class="picker-bar">
          <text class="picker-btn" @tap="showEventPicker = false">取消</text>
          <text class="picker-title">选择参赛项目</text>
          <text class="picker-btn picker-btn--confirm" @tap="onEventConfirm">确定</text>
        </view>
        <picker-view :value="[eventIdx]" class="picker-view" @change="onEventChange">
          <picker-view-column>
            <view v-for="(item, i) in eventOptions" :key="i" class="picker-item">{{ item.label }}</view>
          </picker-view-column>
        </picker-view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref, reactive, computed } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request, authStorage, BASE_URL, getErrorMessage } from '../../../utils/request'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

/* catchtouchmove 的 noop 处理器（阻止弹层打开时页面穿透滚动） */
const noop = () => {}

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const competition = ref(null)
const submitting = ref(false)
const showEventPicker = ref(false)
let backTimer = null

const eventOptions = ref([])
const eventIdx = ref(0)
const selectedEvent = ref('')
const currentPrice = ref(0)
const currentEventType = ref('')

const isTeamEvent = computed(function () {
  return currentEventType.value === '团体赛'
})

/* 参赛项目色条：个人赛=橙、团体赛=绿（对齐发布页色板） */
const eventBarColor = computed(function () {
  return isTeamEvent.value ? '#219653' : '#F97316'
})

/* 报名截止时间：复用详情页 deadlineDate 逻辑（Date.parse 完整时间，含时刻） */
function deadlineDate(item) {
  const d = item && (item.deadline || item.enroll_deadline)
  if (!d) return null
  const t = Date.parse(String(d).replace(/-/g, '/'))
  return isNaN(t) ? null : new Date(t)
}

/* 赛事已截止：status closed/full 或报名截止时刻已过（提交按钮禁用） */
const competitionClosed = computed(function () {
  const c = competition.value
  if (!c) return false
  if (c.status === 'closed' || c.status === 'full') return true
  const dl = deadlineDate(c)
  return !!(dl && dl.getTime() < Date.now())
})

/* 倒计时（按完整时刻计算，对齐详情页「今天截止/剩余 N 天」） */
const countdownText = computed(function () {
  const dl = deadlineDate(competition.value)
  if (!dl) return ''
  const days = Math.floor((dl.getTime() - Date.now()) / 86400000)
  if (days < 0) return '报名已截止'
  if (days === 0) return '今天截止'
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
  // 项目费用未标(0/null)时回退赛事级费用（与列表/详情一致），避免默认选中显示 ¥0
  currentPrice.value = resolveFee(opt.fee)
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

/* 图片上传：POST /api/v1/files/upload。
   isPrivate=true 时（身份证影像）带 private 表单字段，文件落 uploads/private/ 且仅登录态可读；
   后端返回的 url 字段优先使用，缺失时回退 /uploads/{file_id}。 */
function uploadFile(filePath, isPrivate) {
  return new Promise(function (resolve, reject) {
    uni.uploadFile({
      url: BASE_URL + '/api/v1/files/upload',
      filePath: filePath,
      name: 'file',
      formData: { private: isPrivate ? 'true' : 'false' },
      header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
      success: function (res) {
        var data = null
        try { data = JSON.parse(res.data) } catch (e) { data = null }
        if (res.statusCode >= 200 && res.statusCode < 300 && data) {
          var url = data.url || (data.data && data.data.url)
          if (url) { resolve(url); return }
          var fid = data.file_id || (data.data && data.data.file_id)
          if (fid) { resolve('/uploads/' + fid); return }
        }
        var msg = ''
        if (data && data.error && data.error.message) msg = data.error.message
        else if (data && data.message) msg = data.message
        reject(new Error(msg || '图片上传失败（HTTP ' + res.statusCode + '）'))
      },
      fail: function (err) {
        reject(err || new Error('图片上传失败，请检查网络'))
      },
    })
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
  if (competitionClosed.value) { uni.showToast({ title: '报名已截止', icon: 'none' }); return }
  var err = validate()
  if (err) { uni.showToast({ title: err, icon: 'none' }); return }

  submitting.value = true
  try {
    /* 先上传证件照与身份证照，全部成功后才提交报名 */
    uni.showLoading({ title: '上传中...', mask: true })
    var photoUrl = ''
    var idCardUrl = ''
    try {
      photoUrl = await uploadFile(form.photo, false)
      idCardUrl = await uploadFile(form.idCardImage, true)
    } catch (uploadErr) {
      uni.showToast({ title: (uploadErr && uploadErr.message) || '图片上传失败，请重试', icon: 'none' })
      return
    } finally {
      uni.hideLoading()
    }

    await request({
      url: '/api/v1/competitions/' + encodeURIComponent(id.value) + '/register',
      method: 'POST',
      data: {
        team_name: (form.team_name || form.name).trim(),
        member_count: form.member_count,
        contact_info: form.phone.trim(),
        name: form.name.trim(),
        phone: form.phone.trim(),
        id_card: form.idCard.trim(),
        photo_url: photoUrl,
        id_card_image: idCardUrl,
      },
    })
    uni.showModal({
      title: '报名成功',
      content: '可在「我的报名」查看报名记录',
      confirmText: '查看我的报名',
      cancelText: '返回',
      success: (r) => {
        if (r.confirm) {
          uni.navigateTo({ url: '/pkg-talent/pages/training/myenrollments?tab=competition' })
        } else {
          uni.navigateBack()
        }
      },
    })
  } catch (e) {
    var msg = getErrorMessage(e) || '报名失败，请重试'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    submitting.value = false
  }
}

function handleConsult() { uni.showToast({ title: '咨询功能开发中', icon: 'none' }) }

function goBack() { safeBack() }

/* 数据加载 */
async function loadCompetition() {
  loading.value = true
  errorMsg.value = ''

  try {
    var res = await request({ url: '/api/v1/competitions/' + encodeURIComponent(id.value) })
    competition.value = res
    if (!res) errorMsg.value = '赛事不存在'
    loadEvents(res)
  } catch (e) {
    errorMsg.value = '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

/* 赛事级费用兜底：优先赛事 fee，其次 min_fee>0（与列表/详情顶部取值一致）；均无 → null */
const compFallbackFee = computed(function () {
  const it = competition.value
  if (!it) return null
  if (it.fee != null) return it.fee
  if (it.min_fee != null && it.min_fee > 0) return it.min_fee
  return null
})

/* 费用解析：赛事级已配置(>0) → 显示赛事费（整个赛事统一价，与列表/详情一致）；
   赛事级未配置(0) → 用所选项目费(>0)；两者都没有 → null（占位）。
   避免管理后台给预选赛等填了默认 0，导致报名页首屏显示 ¥0 与赛事费 380 不一致 */
function resolveFee(eventFee) {
  var f = compFallbackFee.value
  if (f != null && f > 0) return f
  if (eventFee != null && eventFee > 0) return eventFee
  return null
}

/* 费用文案：费用缺失时不编造价格，统一显示占位 */
const priceText = computed(function () {
  if (currentPrice.value == null) return '以主办方公布为准'
  return currentPrice.value.toLocaleString()
})

function loadEvents(item) {
  // 只使用后端真实下发的参赛项目；缺失时不再兜底 mock 项目/费用
  var list = (item && Array.isArray(item.events)) ? item.events : []
  eventOptions.value = list.map(function (e) {
    return { label: e.name + ' · ' + (e.type || '个人赛'), fee: e.fee != null ? e.fee : null, type: e.type || '个人赛' }
  })
  var first = eventOptions.value[0]
  // 无参赛项目明细（小程序发布/后台未配 events）或项目未标费用时：
  // 回退赛事级 fee（与列表/详情顶部一致），不再显示「以主办方公布为准」造成金额不齐
  var fee = resolveFee(first ? first.fee : null)
  selectedEvent.value = first ? first.label : ''
  currentPrice.value = fee
  currentEventType.value = first ? (first.type || '') : ''
  if (currentEventType.value !== '团体赛') {
    form.member_count = 1
  }
}

onLoad(function (options) {
  initSafeTop()
  id.value = options.id || ''
  loadCompetition()
})

onUnload(function () {
  if (backTimer) clearTimeout(backTimer)
})
</script>

<style scoped>
@import '../../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}

/* 加载中 */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 80px 0;
  color: #667085;
  font-size: 13px;
}

/* 错误重试按钮 */
.retry-btn {
  flex: none;
  margin: 12px auto 0;
  padding: 0 22px;
}

/* ① 赛事卡 */
.event-card {
  position: relative;
  overflow: hidden;
  margin: 0 0 13px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  padding: 13px;
  box-shadow: 0 2px 8px rgba(16, 24, 40, 0.025);
}

.event-card-strip {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, #074D92, #0A66C2);
}

.event-card-title {
  display: block;
  margin-bottom: 8px;
  font-size: 17px;
  font-weight: 750;
  color: #17212B;
  line-height: 1.4;
}

.event-card-tags { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 8px; }

.event-card-tag {
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 750;
  color: #0A66C2;
  background: #E8F2FC;
  border: 1px solid rgba(10, 102, 194, 0.25);
}

.event-card-meta { margin-bottom: 10px; }
.event-card-meta-text { font-size: 12px; color: #667085; line-height: 1.5; }

.event-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid #EEF1F4;
}

.event-card-price { display: flex; align-items: baseline; }
.event-card-symbol { font-size: 12px; color: #F97316; font-weight: 700; }
.event-card-value { font-size: 20px; color: #F97316; font-weight: 800; margin: 0 3px; }
.event-card-unit { font-size: 11px; color: #98A2B3; }

.event-card-countdown {
  font-size: 11px;
  color: #FF3B30;
  background: rgba(255, 59, 48, 0.08);
  padding: 3px 9px;
  border-radius: 999px;
  font-weight: 600;
}

/* ② 参赛项目选择 */
.ep-left {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.ep-bar { width: 3px; height: 32px; border-radius: 2px; flex-shrink: 0; }
.ep-info { flex: 1; min-width: 0; }

.ep-label-row { display: flex; align-items: center; gap: 6px; margin-bottom: 3px; }

.ep-value {
  font-size: 14px;
  font-weight: 650;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ep-type-badge {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 700;
  color: #0A66C2;
  background: #E8F2FC;
  border: 1px solid rgba(10, 102, 194, 0.25);
  padding: 0 7px;
  border-radius: 999px;
  line-height: 1.7;
}

.ep-type-badge--team {
  color: #219653;
  background: #E9F7F0;
  border-color: rgba(33, 150, 83, 0.25);
}

.ep-price { display: flex; align-items: baseline; }
.ep-symbol { font-size: 11px; color: #F97316; font-weight: 700; }
.ep-fee { font-size: 17px; font-weight: 800; color: #F97316; }

/* 报名条件提示条 */
.req-hint {
  display: flex;
  align-items: center;
  gap: 3px;
  border-top: 1px solid #EEF1F4;
  background: #EAF3FB;
  padding: 9px 13px;
  font-size: 11px;
}

/* 参赛项目选择说明：仅作费用参考展示，最终以现场确认为准 */
.event-confirm-note {
  padding: 9px 13px;
  border-top: 1px solid #EEF1F4;
  font-size: 11px;
  color: #E96012;
  background: #FFF8F3;
}

.req-hint-text { color: #667085; }
.req-hint-link { color: #0A66C2; font-weight: 700; text-decoration: underline; }

/* ③ 队员人数 stepper */
.stepper-wrap { display: flex; align-items: center; gap: 12px; }

.stepper-btn {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #F5F6F8;
  border: 1px solid #D7E1EA;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  color: #17212B;
  font-weight: 500;
}

.stepper-btn.disabled { color: #C8D7E6; }
.stepper-val { font-size: 14px; font-weight: 700; color: #17212B; min-width: 28px; text-align: center; }

/* ④ 证件上传 */
.reg-upload-row { padding: 0; }

.pub-photo-tag {
  position: absolute;
  right: 3px;
  bottom: 3px;
  padding: 1px 6px;
  background: rgba(23, 33, 43, 0.55);
  color: #fff;
  font-size: 9px;
  border-radius: 4px;
}

/* 声明 */
.agree-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 13px;
  border-top: 1px solid #EEF1F4;
  font-size: 13px;
  color: #667085;
}

.agree-row:first-child { border-top: 0; }

.agree-box {
  width: 18px;
  height: 18px;
  border: 1px solid #A9B9C9;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.agree-box.checked { background: #0A66C2; border-color: #0A66C2; }
.agree-check { color: #fff; font-size: 12px; font-weight: 700; line-height: 1; }
.agree-link { color: #0A66C2; text-decoration: underline; }

/* ⑤ 费用 */
.fee-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 13px;
}

.fee-label { font-size: 12px; color: #667085; display: block; margin-bottom: 3px; }
.fee-amount { display: flex; align-items: baseline; }
.fee-symbol { font-size: 12px; color: #F97316; font-weight: 700; }
.fee-value { font-size: 22px; font-weight: 800; color: #F97316; line-height: 1; }
.fee-suffix { font-size: 11px; color: #98A2B3; margin-left: 3px; }
.fee-sub { font-size: 11px; color: #98A2B3; display: block; margin-top: 4px; }

.fee-countdown {
  padding: 4px 10px;
  background: rgba(255, 59, 48, 0.08);
  color: #FF3B30;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
  flex-shrink: 0;
}

/* ⑥ 底部操作区 */
.pub-sticky .pub-btn--secondary { flex: 1; }
.is-loading { opacity: 0.7; }
.is-disabled { background: #B0B8C4 !important; box-shadow: none !important; pointer-events: none; }
.btn-arrow { margin-left: 6px; font-size: 15px; }

/* 已截止横幅（固定于底部操作区上方） */
.pub-closed-banner {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(env(safe-area-inset-bottom) + 132rpx);
  z-index: 98;
  padding: 12rpx 16rpx;
  background: #FDECEC;
  color: #B42318;
  font-size: 24rpx;
  text-align: center;
  border-radius: 10rpx;
}

.privacy-text {
  display: block;
  text-align: center;
  font-size: 11px;
  color: #98A2B3;
  margin: 4px 0 0;
}

/* 参赛项目选择弹层（picker-view） */
.pub-sheet--picker { max-height: none; }

.picker-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0 10px;
  border-bottom: 1px solid #EEF1F4;
}

.picker-btn { font-size: 14px; color: #667085; padding: 4px; }
.picker-btn--confirm { color: #0A66C2; font-weight: 700; }
.picker-title { font-size: 15px; font-weight: 750; color: #17212B; }

.picker-view { height: 200px; }

.picker-item {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: #17212B;
}

/* ===== 对齐培训模块质感 ===== */
/* 页面底色：往培训模块浅蓝灰对齐 */
.pub-page { background: #F5F8FC; }

/* 表单卡：16px 圆角 + 柔和投影（对齐培训报名页卡片语言） */
.pub-form-card {
  background: #ffffff;
  border: 1px solid #E8EDF3;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(16, 24, 40, 0.06);
}

.pub-section-title { font-size: 15px; font-weight: 750; color: #17212B; margin: 0 0 10px; }
.pub-section-note { color: #98A2B3; margin: -3px 0 10px; font-size: 11px; }

/* 双 CTA：12px 圆角（对齐培训模块按钮） */
.pub-btn--primary, .pub-btn--secondary { border-radius: 12px; }

/* 底部固定栏：毛白底 + 上投影 */
.pub-sticky {
  background: rgba(255, 255, 255, 0.96);
  border-top: 1px solid #E8EDF3;
  box-shadow: 0 -6px 18px rgba(16, 24, 40, 0.06);
}

/* 赛事摘要卡：16px 圆角 + 更饱满投影（呼应培训报名页摘要卡） */
.event-card {
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(10, 102, 194, 0.08);
}
</style>
