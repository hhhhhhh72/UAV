<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !competition"
      empty-text="赛事不存在"
      @retry="loadCompetition"
    >
      <template v-if="competition">
        <!-- ① 海军蓝 Banner -->
        <view class="banner">
          <view class="banner-nav">
            <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
            <text class="banner-nav-title">赛事报名</text>
          </view>
          <text class="banner-comp-name">{{ competition.name || competition.title }}</text>
          <text class="banner-comp-sub">{{ competition.organizer || '中国航空器拥有者及驾驶员协会' }}</text>
          <view class="banner-tags">
            <text v-for="t in (competition.tags || compTags(competition))" :key="t" class="banner-tag">{{ t }}</text>
          </view>
        </view>

        <!-- ② 主卡片 -->
        <view class="main-card">
          <!-- 参赛项目选择 -->
          <view class="event-picker" @click="showEventPicker = true">
            <view>
              <text class="picker-label">选择参赛项目</text>
              <text class="picker-value">{{ selectedEvent }}</text>
            </view>
            <text class="picker-arrow">›</text>
          </view>

          <!-- 报名条件提示 -->
          <view class="req-hint">
            <text class="req-hint-icon">!</text>
            <text class="req-hint-text">请确认已满足</text>
            <text class="req-hint-link" @click="goBack">报名条件</text>
            <text class="req-hint-text">，再填写表单</text>
          </view>

          <!-- ③ 个人信息表单 -->
          <view class="section-header">
            <view class="section-bar" />
            <text class="section-title">参赛人员信息</text>
            <text class="section-badge">*必填</text>
          </view>

          <view class="form-group">
            <view class="form-item">
              <text class="form-label">姓名</text>
              <input class="form-input" v-model="form.name" placeholder="请输入参赛人姓名" />
              <text class="form-required">*</text>
            </view>
            <view class="form-item">
              <text class="form-label">手机号</text>
              <input class="form-input" v-model="form.phone" type="number" maxlength="11" placeholder="请输入手机号码" />
              <text class="form-required">*</text>
            </view>
            <view class="form-item">
              <text class="form-label">身份证号</text>
              <input class="form-input" v-model="form.idCard" maxlength="18" placeholder="请输入身份证号码" />
              <text class="form-required">*</text>
            </view>
            <template v-if="isTeamEvent">
              <view class="form-item">
                <text class="form-label">队伍名称</text>
                <input class="form-input" v-model="form.team_name" placeholder="请输入队伍名称" />
              </view>
              <view class="form-item">
                <text class="form-label">队员人数</text>
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

          <!-- 展开更多 -->
          <view class="expand-btn" @click="showMore = !showMore">
            <text>{{ showMore ? '收起更多信息' : '展开更多信息' }}</text>
            <text class="expand-arrow">{{ showMore ? '▲' : '▼' }}</text>
          </view>

          <view v-if="showMore" class="extra-form">
            <text class="extra-label">补充信息 <text class="optional">选填</text></text>

            <view class="extra-item">
              <text class="form-label">性别</text>
              <view class="radio-group">
                <text v-for="g in ['男','女']" :key="g" class="radio-item"
                  :class="{ active: form.gender === g }" @click="form.gender = g">
                  {{ form.gender === g ? '●' : '○' }} {{ g }}
                </text>
              </view>
            </view>
            <view class="extra-item">
              <text class="form-label">出生日期</text>
              <picker mode="date" :value="form.birthday" @change="onBirthdayChange">
                <text class="picker-text">{{ form.birthday || '____年__月__日' }}</text>
              </picker>
            </view>
            <view class="extra-item">
              <text class="form-label">电子邮箱</text>
              <input class="form-input" v-model="form.email" placeholder="请输入邮箱" />
            </view>
            <view class="extra-item">
              <text class="form-label">所属单位</text>
              <input class="form-input" v-model="form.unit" placeholder="选填" />
            </view>
            <view class="extra-item extra-item-last">
              <text class="form-label">参赛经验</text>
              <view class="radio-group">
                <text v-for="l in expList" :key="l" class="radio-item radio-sm"
                  :class="{ active: form.experience === l }" @click="form.experience = l">
                  {{ form.experience === l ? '●' : '○' }} {{ l }}
                </text>
              </view>
            </view>
          </view>

          <!-- ④ 证件上传 -->
          <view class="section-header">
            <view class="section-bar" />
            <text class="section-title">证件上传</text>
            <text class="section-badge">*必传</text>
          </view>

          <view class="upload-row">
            <view class="upload-box" @click="uploadImage('photo')">
              <image v-if="form.photo" :src="form.photo" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <text class="upload-icon">照</text>
                <text class="upload-title">白底免冠证件照</text>
                <text class="upload-hint">点击上传</text>
              </view>
            </view>
            <view class="upload-box" @click="uploadImage('idCard')">
              <image v-if="form.idCardImage" :src="form.idCardImage" class="upload-preview" mode="aspectFill" />
              <view v-else class="upload-placeholder">
                <text class="upload-icon">证</text>
                <text class="upload-title">身份证正面</text>
                <text class="upload-hint">点击上传</text>
              </view>
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

          <!-- ⑤ 费用 -->
          <view class="price-section">
            <view class="price-row">
              <text class="price-label">报名费用</text>
              <text class="price-value">¥{{ currentPrice.toLocaleString() }}</text>
            </view>
            <view class="price-row">
              <text class="price-label">单位</text>
              <text class="price-unit">/人</text>
            </view>
          </view>

          <!-- ⑥ 底部按钮 -->
          <view class="bottom-bar">
            <view class="btn-outline" @click="handleConsult">联系咨询</view>
            <view class="btn-primary" @click="handleSubmit">确认报名</view>
          </view>
          <text class="privacy-text">报名信息仅用于赛事注册，受隐私政策保护</text>

          <view class="bottom-spacer" />
        </view>
      </template>
    </StateView>

    <!-- Event picker -->
    <u-picker
      :show="showEventPicker"
      :columns="eventNames"
      :model-value="selectedEvent"
      title="选择参赛项目"
      @confirm="onEventChange"
      @update:show="showEventPicker = $event"
    />
  </view>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const competition = ref(null)
const submitting = ref(false)
const showMore = ref(false)
const showEventPicker = ref(false)
const expList = ['零基础', '业余爱好者', '有参赛经验']

const eventOptions = ref([])
const selectedEvent = ref('')
const currentPrice = ref(0)
const currentEventType = ref('')

const isTeamEvent = computed(function () {
  return currentEventType.value === '团体赛'
})

/* u-picker 只接受字符串数组，派生参赛项目名称列 */
const eventNames = computed(function () {
  return eventOptions.value.map(function (e) { return e.label })
})

const form = reactive({
  name: '', phone: '', idCard: '',
  gender: '', birthday: '', email: '',
  unit: '', experience: '',
  team_name: '', member_count: 3,
  photo: '', idCardImage: '',
  agreeHealth: false, agreeRules: false,
})

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return ['多旋翼', '国家级']
}

/* 身份证自动推导 */
watch(function () { return form.idCard }, function (val) {
  if (val && val.length === 18) {
    var birth = val.substring(6, 14)
    form.birthday = birth.substring(0, 4) + '-' + birth.substring(4, 6) + '-' + birth.substring(6, 8)
    form.gender = parseInt(val.charAt(16), 10) % 2 === 0 ? '女' : '男'
  }
})

function onBirthdayChange(e) { form.birthday = e.detail.value }

function onEventChange(val) {
  var idx = eventOptions.value.findIndex(function (e) { return e.label === val })
  if (idx < 0) { showEventPicker.value = false; return }
  selectedEvent.value = eventOptions.value[idx].label
  currentPrice.value = eventOptions.value[idx].fee
  currentEventType.value = eventOptions.value[idx].type || ''
  // 切换到个人赛时重置队伍人数
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
        contact_info: (form.phone + ' ' + (form.email || '')).trim(),
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
      events: [
        { name: 'FPV竞速赛', type: '个人赛', fee: 280 },
        { name: 'FPV花飞表演赛', type: '个人赛', fee: 0 },
      ],
    },
    'comp-3': {
      id: 'comp-3', name: '2026无人机创新应用大赛', title: '2026无人机创新应用大赛',
      organizer: '工信部人才交流中心',
      tags: ['航拍', '固定翼', '国家级'],
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
      events: [
        { name: '初级编程挑战', type: '个人赛', fee: 120 },
        { name: '高级编程挑战', type: '个人赛', fee: 120 },
      ],
    },
    'comp-5': {
      id: 'comp-5', name: '国际无人机系统博览会竞技赛', title: '国际无人机系统博览会竞技赛',
      organizer: '广州市低空经济产业协会',
      tags: ['多旋翼', '固定翼', '国际赛'],
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
  // 默认个人赛时重置队伍人数
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
.page { min-height: 100vh; background: var(--color-bg); }

/* ① Banner */
.banner {
  background: linear-gradient(135deg, #1a365d, #2a4a7f);
  padding: 80rpx 32rpx 64rpx;
}

.banner-nav { display: flex; align-items: center; gap: 12rpx; margin-bottom: 24rpx; }

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.15);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-nav-title { color: rgba(255,255,255,0.9); font-size: 28rpx; font-weight: 500; }

.banner-comp-name {
  color: #ffffff; font-size: 56rpx; font-weight: 700; line-height: 1.2; margin-bottom: 8rpx;
}

.banner-comp-sub { color: rgba(255,255,255,0.7); font-size: 26rpx; margin-bottom: 12rpx; }

.banner-tags { display: flex; gap: 12rpx; }

.banner-tag {
  padding: 6rpx 18rpx; border: 1rpx solid rgba(255,255,255,0.4);
  border-radius: 20rpx; color: rgba(255,255,255,0.9); font-size: 22rpx;
}

/* ② 主卡片 */
.main-card {
  background: #ffffff; border-radius: 32rpx 32rpx 0 0; margin-top: -32rpx;
  padding: 40rpx 32rpx 32rpx; position: relative; z-index: 2;
}

.event-picker {
  background: #f8f9fc; border-radius: 18rpx; padding: 24rpx;
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 20rpx;
}

.picker-label { font-size: 24rpx; color: var(--color-primary); display: block; margin-bottom: 6rpx; }
.picker-value { font-size: 30rpx; font-weight: 600; color: var(--color-primary); }
.picker-arrow { font-size: 36rpx; color: var(--color-primary); }

.req-hint {
  display: flex; align-items: center; gap: 8rpx;
  margin-bottom: 32rpx; padding: 0 4rpx;
}

.req-hint-icon {
  width: 28rpx; height: 28rpx; border-radius: 50%;
  background: var(--color-warning); color: #ffffff;
  font-size: 22rpx; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.req-hint-text { font-size: 24rpx; color: var(--color-text-secondary); }
.req-hint-link { font-size: 24rpx; color: var(--color-primary); text-decoration: underline; font-weight: 500; }

/* ③ 表单 */
.section-header { display: flex; align-items: center; margin-bottom: 24rpx; }

.section-bar {
  width: 6rpx; height: 32rpx; background: var(--color-primary); border-radius: 3rpx; margin-right: 12rpx;
}

.section-title { font-size: 30rpx; font-weight: 700; color: var(--color-text); }
.section-badge { font-size: 22rpx; color: var(--color-warning); margin-left: 8rpx; }

.form-group { margin-bottom: 8rpx; }

.form-item {
  display: flex; align-items: center; padding: 22rpx 0; border-bottom: 1rpx solid #ebedf0;
}

.form-label { font-size: 28rpx; color: var(--color-text); width: 130rpx; flex-shrink: 0; }
.form-input { flex: 1; font-size: 28rpx; color: var(--color-text); }
.form-required { font-size: 22rpx; color: var(--color-warning); }

.stepper-wrap { display: flex; align-items: center; gap: 20rpx; }

.stepper-btn {
  width: 56rpx; height: 56rpx; border-radius: 50%; background: #f5f6f8;
  display: flex; align-items: center; justify-content: center;
  font-size: 32rpx; color: #1a1a1a; font-weight: 500;
}

.stepper-btn.disabled { color: #c0c4cc; }
.stepper-val { font-size: 30rpx; font-weight: 600; color: #1a1a1a; min-width: 40rpx; text-align: center; }

.expand-btn {
  display: flex; align-items: center; justify-content: center; gap: 8rpx;
  padding: 20rpx 0; color: var(--color-primary); font-size: 26rpx; font-weight: 500;
}

.expand-arrow { font-size: 20rpx; }

.extra-form { background: #f8fafc; border-radius: 18rpx; padding: 24rpx; margin-bottom: 8rpx; }

.extra-label { font-size: 24rpx; color: #969799; display: block; margin-bottom: 20rpx; }
.optional { color: #c0c4cc; font-size: 22rpx; }

.extra-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18rpx 0; border-bottom: 1rpx solid #ebedf0;
}

.extra-item-last { border-bottom: none; }

.radio-group { display: flex; gap: 32rpx; }
.radio-item { font-size: 26rpx; color: #c0c4cc; }
.radio-sm { font-size: 24rpx; }
.radio-item.active { color: var(--color-primary); font-weight: 500; }
.picker-text { font-size: 28rpx; color: #c0c4cc; }

/* ④ 证件上传 */
.upload-row { display: flex; gap: 20rpx; margin-bottom: 20rpx; }

.upload-box {
  flex: 1; height: 220rpx; background: #f8f9fc; border-radius: 16rpx;
  border: 2rpx dashed #d0d5dd; overflow: hidden;
}

.upload-preview { width: 100%; height: 100%; }

.upload-placeholder {
  height: 100%; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 6rpx;
}

.upload-icon { font-size: 48rpx; opacity: 0.5; }
.upload-title { font-size: 24rpx; color: #969799; }
.upload-hint { font-size: 22rpx; color: #c0c4cc; }

.checkbox-row {
  display: flex; align-items: center; gap: 12rpx; padding: 16rpx 0;
}

.checkbox-box {
  width: 36rpx; height: 36rpx; border: 2rpx solid #d0d5dd; border-radius: 8rpx;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.checkbox-box.checked { background: var(--color-primary); border-color: var(--color-primary); }
.check-mark { color: #ffffff; font-size: 24rpx; font-weight: 700; }
.checkbox-text { font-size: 26rpx; color: var(--color-text); }
.link { color: var(--color-primary); text-decoration: underline; }

/* ⑤ 费用 */
.price-section { border-top: 1rpx solid #ebedf0; padding-top: 24rpx; margin-bottom: 32rpx; }

.price-row { display: flex; justify-content: space-between; align-items: center; padding: 8rpx 0; }
.price-label { font-size: 26rpx; color: #969799; }
.price-value { font-size: 44rpx; font-weight: 700; color: var(--color-warning); }
.price-unit { font-size: 26rpx; color: var(--color-text); }

/* ⑥ 底部按钮 */
.bottom-bar { display: flex; gap: 20rpx; }

.btn-outline {
  flex: 1; height: 96rpx; border-radius: 48rpx;
  border: 2rpx solid var(--color-primary); background: #ffffff; color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-size: 32rpx; font-weight: 600;
}

.btn-primary {
  flex: 1; height: 96rpx; border-radius: 48rpx;
  background: var(--color-primary); color: #ffffff;
  display: flex; align-items: center; justify-content: center;
  font-size: 32rpx; font-weight: 600;
}

.privacy-text { display: block; text-align: center; font-size: 22rpx; color: #c0c4cc; margin-top: 16rpx; }
.bottom-spacer { height: calc(40rpx + env(safe-area-inset-bottom)); }
</style>
