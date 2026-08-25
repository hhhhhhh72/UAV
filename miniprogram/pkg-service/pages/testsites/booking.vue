<template>
  <view class="tsb-page" :class="{ 'no-motion': noMotion }">
    <u-nav-bar title="预约测试" show-back @back="goBack" />

    <!-- 骨架（对齐组灰条，与 detail 同构） -->
    <view v-if="loading">
      <view class="sk-section">
        <view class="sk-l w40"></view>
        <view class="sk-l w60"></view>
        <view class="sk-l w80"></view>
      </view>
      <view class="sk-section">
        <view class="sk-l w50"></view>
        <view class="sk-l w70"></view>
        <view class="sk-l w90"></view>
      </view>
    </view>

    <!-- 错误：渲染真实原因；场地不存在 → 返回列表（重试是死路） -->
    <view v-else-if="errorMsg" class="st">
      <u-empty :description="errorMsg">
        <view class="stb" @tap="onRetry">{{ retryLabel }}</view>
      </u-empty>
    </view>

    <template v-else-if="site">
      <!-- 预约摘要条（蓝雾；未选项目显示占位，提交时才校验） -->
      <view class="sum-bar">
        <text class="sum-name">{{ site.name }}</text>
        <text class="sum-sep">·</text>
        <text class="sum-pick" :class="{ pending: !form.date }">{{ form.date || '未选日期' }}</text>
        <text class="sum-sep">·</text>
        <text class="sum-pick" :class="{ pending: !form.timeSlots.length }">{{ form.timeSlots.length ? form.timeSlots.join('、') : '未选时段' }}</text>
      </view>

      <!-- Step 1 预约类型 -->
      <view class="step-card">
        <view class="step-label"><text class="step-num">1</text><text>预约类型</text></view>
        <view class="type-grid">
          <view
            class="type-card"
            :class="{ selected: bookingType === 'personal' }"
            @tap="switchType('personal')"
          >
            <view class="type-icon fig-single">
              <view class="fig-head"></view>
              <view class="fig-body"></view>
            </view>
            <text class="type-name">个人预约</text>
            <text class="type-desc">单飞手 / 独立开发者</text>
          </view>
          <view
            class="type-card"
            :class="{ selected: bookingType === 'group' }"
            @tap="switchType('group')"
          >
            <view class="type-icon fig-group">
              <view class="fig fig-a"><view class="fig-head"></view><view class="fig-body"></view></view>
              <view class="fig fig-b"><view class="fig-head"></view><view class="fig-body"></view></view>
            </view>
            <text class="type-name">团体预约</text>
            <text class="type-desc">公司团队 / 培训教学 / 入驻飞手</text>
          </view>
        </view>
      </view>

      <!-- Step 2 选择日期（chips + 完整日历） -->
      <view class="step-card">
        <view class="step-label"><text class="step-num">2</text><text>选择日期</text></view>
        <view class="pick-card">
          <view class="pick-card-title">
            <text class="pc-title">预约日期</text>
            <text class="pc-hint">点击「日历」查看完整月份</text>
          </view>
          <view class="date-chips">
            <view
              v-for="c in dateChips"
              :key="c.key"
              class="date-chip"
              :class="{ selected: form.date === c.key }"
              @tap="pickChip(c.key)"
            >
              <text class="dc-day">{{ c.label }}</text>
              <text class="dc-week">{{ c.week }}</text>
            </view>
            <view class="cal-btn" @tap="openCalendar">
              <view class="cal-icon"></view>
              <text>日历</text>
            </view>
          </view>
        </view>
      </view>

      <!-- Step 3 选择时段（多选） -->
      <view class="step-card">
        <view class="step-label"><text class="step-num">3</text><text>选择时段</text></view>
        <view class="pick-card">
          <view class="pick-card-title">
            <text class="pc-title">意向时段</text>
            <text class="pc-hint">可多选</text>
          </view>
          <view class="slot-grid">
            <view
              v-for="slot in slots"
              :key="slot"
              class="slot-item"
              :class="{ selected: form.timeSlots.indexOf(slot) >= 0 }"
              @tap="toggleSlot(slot)"
            >{{ slot }}</view>
          </view>
          <text class="form-hint">占用情况以场地方确认为准，提交后 24 小时内电话联系确认</text>
        </view>
      </view>

      <!-- Step 4 填写预约信息 -->
      <view class="step-card">
        <view class="step-label"><text class="step-num">4</text><text>填写预约信息</text></view>

        <!-- 个人预约 -->
        <view v-if="bookingType === 'personal'" class="type-form">
          <view class="form-group">
            <view class="form-label">姓名 <text class="required">*</text></view>
            <input class="form-input" v-model="form.pName" placeholder="请输入姓名" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">手机号 <text class="required">*</text></view>
            <input class="form-input" v-model="form.pPhone" type="number" maxlength="11" placeholder="请输入手机号" placeholder-class="ph" />
            <text v-if="phoneErr.p" class="form-error">请输入正确的手机号</text>
          </view>
          <view class="form-group">
            <view class="form-label">无人机型号 <text class="required">*</text></view>
            <input class="form-input" v-model="form.pModel" placeholder="请输入无人机型号" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">飞手执照<text class="optional">（选填）</text></view>
            <view class="upload-zone" @tap="chooseCert('license')">
              <image v-if="certPreview.license" class="up-img" :src="certPreview.license" mode="aspectFill" />
              <template v-else>
                <text class="up-plus">+</text>
                <text class="up-tip">点击上传执照照片</text>
              </template>
              <text v-if="certPreview.license" class="up-note">已上传，点击更换</text>
            </view>
          </view>
        </view>

        <!-- 团体预约 -->
        <view v-else class="type-form">
          <view class="form-group">
            <view class="form-label">团队名称 <text class="required">*</text></view>
            <input class="form-input" v-model="form.gTeam" placeholder="请输入团队名称" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">参与人数 <text class="required">*</text></view>
            <input class="form-input" v-model="form.gPeople" type="number" placeholder="请输入参与人数" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">设备清单 <text class="required">*</text></view>
            <textarea class="form-textarea" v-model="form.gEquipList" placeholder="请列出测试涉及的设备名称和数量" placeholder-class="ph"></textarea>
          </view>
          <view class="form-group">
            <view class="form-label">联系人姓名 <text class="required">*</text></view>
            <input class="form-input" v-model="form.gContactName" placeholder="请输入联系人姓名" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">联系人手机 <text class="required">*</text></view>
            <input class="form-input" v-model="form.gContactPhone" type="number" maxlength="11" placeholder="请输入联系人手机" placeholder-class="ph" />
            <text v-if="phoneErr.g" class="form-error">请输入正确的手机号</text>
          </view>
          <view class="form-group">
            <view class="form-label">资质证明<text class="optional">（选填 · 培训教学类必填）</text></view>
            <view class="upload-zone" @tap="chooseCert('qual')">
              <image v-if="certPreview.qual" class="up-img" :src="certPreview.qual" mode="aspectFill" />
              <template v-else>
                <text class="up-plus">+</text>
                <text class="up-tip">点击上传营业执照 / 办学许可</text>
              </template>
              <text v-if="certPreview.qual" class="up-note">已上传，点击更换</text>
            </view>
          </view>
        </view>

        <!-- 公共：测试内容描述 + 设备备注 -->
        <view class="form-group">
          <view class="form-label">测试内容描述 <text class="required">*</text><text class="optional">（≤500字）</text></view>
          <textarea class="form-textarea" v-model="form.desc" maxlength="500" placeholder="描述测试目的、项目、飞行参数等..." placeholder-class="ph"></textarea>
          <text class="char-count">{{ form.desc.length }}/500</text>
        </view>
        <view class="form-group">
          <view class="form-label">设备需求备注<text class="optional">（选填）</text></view>
          <textarea class="form-textarea short" v-model="form.equipmentNote" placeholder="如需特殊测试设备请注明..." placeholder-class="ph"></textarea>
        </view>
      </view>

      <!-- 资金说明 -->
      <view class="notice-block">
        <text class="notice-title">费用说明</text>
        <text class="notice-line">· 测试费用在线下向场地方支付，平台不参与资金流转</text>
        <text class="notice-line">· 费用标准以场地方实际计费方式（按次 / 按时段）为准</text>
      </view>
    </template>

    <!-- 底部提交栏 -->
    <view v-if="site" class="submit-bar">
      <view class="submit-btn" :class="{ disabled: !canSubmit || submitting }" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '提交预约' }}
      </view>
    </view>

    <!-- 预约须知（进入页面自动弹出，同意后方可提交） -->
    <u-popup :show="noticeShow" position="center" :close-on-click-overlay="false">
      <view class="notice-dialog">
        <text class="nd-title">预约须知</text>
        <text class="nd-sub">请阅读以下须知后继续预约</text>
        <view class="notice-list">
          <view class="nl-item">1. 试飞前须完成实名登记及飞行报备</view>
          <view class="nl-item">2. 测试期间须遵守场地安全规范</view>
          <view class="nl-item">3. 设备或人身安全事故由预约方自行承担责任</view>
          <view class="nl-item">4. 取消预约需提前 24 小时通知场地管理方</view>
          <view class="nl-item">5. 预约时段开始后 30 分钟未到场视为自动放弃，费用不予退还</view>
        </view>
        <view class="nd-actions">
          <view class="nd-btn ghost" @tap="declineNotice">暂不预约</view>
          <view class="nd-btn" @tap="agreeNotice">我已阅读并同意</view>
        </view>
      </view>
    </u-popup>

    <!-- 完整日历（底部弹出，本月 + 下月） -->
    <u-popup :show="calShow" position="bottom" round :close-on-click-overlay="false" @close="calShow = false">
      <view class="cal-sheet">
        <view class="sheet-hd">
          <text class="sheet-title">选择日期</text>
          <text class="sheet-close" @tap="calShow = false">✕</text>
        </view>
        <view v-for="(m, mi) in calCells" :key="mi" class="cal-month-block">
          <text class="cal-month">{{ m.monthLabel }}</text>
          <view class="cal-week">
            <text v-for="w in weekNames" :key="w" class="cal-week-cell">{{ w }}</text>
          </view>
          <view class="cal-grid">
            <view v-for="n in m.lead" :key="'lead' + n" class="cal-cell empty"></view>
            <view
              v-for="c in m.days"
              :key="c.key"
              class="cal-cell"
              :class="{ past: c.past, selected: c.key === pendingCalKey }"
              @tap="calPick(c)"
            >
              {{ c.day }}
            </view>
          </view>
        </view>
        <view class="cal-confirm" @tap="confirmCalendar">确定</view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage, getStoredUser } from '@/utils/request'
import { safeBack } from '@/utils/nav'
import { useReduceMotion } from '@/utils/motion'

// 预约时段（多选；硬编码规则，后端无时段数据——照产品原型 6 段）
const SLOTS = [
  '08:00-10:00',
  '10:00-12:00',
  '13:00-15:00',
  '15:00-17:00',
  '17:00-19:00',
  '19:00-21:00',
]

const site = ref(null)
const loading = ref(true)
const errorMsg = ref('')
const submitting = ref(false)
const slots = SLOTS
const { noMotion, checkMotion } = useReduceMotion()

// 预约类型（原型：个人 / 团体 二选一）
const bookingType = ref('personal')
const form = reactive({
  date: '',
  timeSlots: [],
  desc: '',
  equipmentNote: '',
  // 个人
  pName: '',
  pPhone: '',
  pModel: '',
  pLicenseUrl: '',
  // 团体
  gTeam: '',
  gPeople: '',
  gEquipList: '',
  gContactName: '',
  gContactPhone: '',
  gQualUrl: '',
})
// 上传预览（本地临时路径，与提交 URL 分离）
const certPreview = reactive({ license: '', qual: '' })

// 须知弹窗
const noticeShow = ref(false)
const noticeAgreed = ref(false)

// 日历
const calShow = ref(false)
const calCells = ref([])
const pendingCalKey = ref('')
const weekNames = ['日', '一', '二', '三', '四', '五', '六']

let siteId = ''

const phoneErr = computed(() => {
  return {
    p: !!form.pPhone && !/^1[3-9]\d{9}$/.test(form.pPhone),
    g: !!form.gContactPhone && !/^1[3-9]\d{9}$/.test(form.gContactPhone),
  }
})

// 错误恢复：场地不存在/下架 → 返回列表（重试是死路）；网络错误 → 重新加载
const retryLabel = computed(() => {
  if (errorMsg.value === '场地不存在或已下架') return '返回场地列表'
  return '重新加载'
})

// 缺失项枚举（提交被拦时逐项提示，替代笼统 toast）
function missingFields() {
  const miss = []
  if (!form.desc.trim()) miss.push('测试内容描述')
  if (!form.date) miss.push('预约日期')
  if (form.timeSlots.length === 0) miss.push('意向时段')
  if (!noticeAgreed) miss.push('同意预约须知')
  if (bookingType.value === 'personal') {
    if (!form.pName.trim()) miss.push('联系人姓名')
    if (!/^1[3-9]\d{9}$/.test(form.pPhone)) miss.push('11 位联系电话')
    if (!form.pModel.trim()) miss.push('无人机型号')
  } else {
    if (!form.gTeam.trim()) miss.push('团队名称')
    if (!(Number(form.gPeople) >= 1)) miss.push('参与人数')
    if (!form.gEquipList.trim()) miss.push('设备清单')
    if (!form.gContactName.trim()) miss.push('联系人姓名')
    if (!/^1[3-9]\d{9}$/.test(form.gContactPhone)) miss.push('11 位联系电话')
  }
  return miss
}

// 完整度校验（照原型：全部满足才可提交）
const canSubmit = computed(() => {
  if (!form.desc.trim() || !form.date || form.timeSlots.length === 0 || !noticeAgreed) return false
  if (bookingType.value === 'personal') {
    return !!form.pName.trim() && /^1[3-9]\d{9}$/.test(form.pPhone) && !!form.pModel.trim()
  }
  return (
    !!form.gTeam.trim() &&
    Number(form.gPeople) >= 1 &&
    !!form.gEquipList.trim() &&
    !!form.gContactName.trim() &&
    /^1[3-9]\d{9}$/.test(form.gContactPhone)
  )
})

function pad(n) { return n < 10 ? '0' + n : '' + n }
function keyOf(d) { return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) }
function today() { return keyOf(new Date()) }
const WEEK = '日一二三四五六'

// 日期 chips：今天起 4 天（原型）
const dateChips = computed(() => {
  const out = []
  const now = new Date()
  for (let i = 0; i < 4; i++) {
    const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + i)
    out.push({
      key: keyOf(d),
      label: i === 0 ? '今天' : (d.getMonth() + 1) + '/' + d.getDate(),
      week: '周' + WEEK[d.getDay()],
    })
  }
  return out
})

function pickChip(key) {
  if (key === form.date) return
  if (form.timeSlots.length === 0) {
    form.date = key
    return
  }
  // 清空已选时段是破坏性操作：先确认（同 switchType 的模式）
  uni.showModal({
    title: '更换日期',
    content: '更换日期将清空已选时段，是否继续？',
    confirmText: '更换',
    confirmColor: '#0A66C2',
    success: function (res) {
      if (!res.confirm) return
      form.date = key
      form.timeSlots = []
    },
  })
}

// 完整日历：本月 + 下月，过去置灰
function buildCalendar() {
  const now = new Date()
  const todayKey = keyOf(now)
  const cells = []
  for (let off = 0; off < 2; off++) {
    const first = new Date(now.getFullYear(), now.getMonth() + off, 1)
    const days = new Date(now.getFullYear(), now.getMonth() + off + 1, 0).getDate()
    const lead = first.getDay()
    const monthLabel = off === 0
      ? now.getFullYear() + '年' + (now.getMonth() + 1) + '月（本月）'
      : (first.getFullYear()) + '年' + (first.getMonth() + 1) + '月'
    const dayCells = []
    for (let d = 1; d <= days; d++) {
      const dt = new Date(now.getFullYear(), now.getMonth() + off, d)
      const k = keyOf(dt)
      dayCells.push({
        day: d,
        key: k,
        past: k < todayKey,
        label: (now.getMonth() + off + 1) + '月' + d + '日',
      })
    }
    cells.push({ monthLabel, lead, days: dayCells })
  }
  return cells
}

function openCalendar() {
  calCells.value = buildCalendar()
  pendingCalKey.value = form.date || ''
  calShow.value = true
}

function calPick(c) {
  if (c.past) return
  pendingCalKey.value = c.key
}

function confirmCalendar() {
  if (!pendingCalKey.value) {
    uni.showToast({ title: '请先选择日期', icon: 'none' })
    return
  }
  if (pendingCalKey.value === form.date) {
    calShow.value = false
    return
  }
  if (form.timeSlots.length === 0) {
    form.date = pendingCalKey.value
    calShow.value = false
    return
  }
  // 清空已选时段是破坏性操作：先确认
  uni.showModal({
    title: '更换日期',
    content: '更换日期将清空已选时段，是否继续？',
    confirmText: '更换',
    confirmColor: '#0A66C2',
    success: function (res) {
      if (!res.confirm) return
      form.date = pendingCalKey.value
      form.timeSlots = []
      calShow.value = false
    },
  })
}

function toggleSlot(slot) {
  if (!form.date) {
    uni.showToast({ title: '请先选择日期', icon: 'none' })
    return
  }
  const idx = form.timeSlots.indexOf(slot)
  if (idx >= 0) {
    form.timeSlots.splice(idx, 1)
  } else {
    form.timeSlots.push(slot)
  }
}

// 切换预约类型：确认后清空「旧类型」专属字段（desc/equipmentNote/date/timeSlots 为共享字段，保留）
function switchType(type) {
  if (type === bookingType.value) return
  uni.showModal({
    title: '切换预约类型',
    content: '切换后将清空当前类型的表单内容，是否继续？',
    confirmText: '确认切换',
    confirmColor: '#0A66C2',
    success: function (res) {
      if (!res.confirm) return
      clearFormFields(bookingType.value)
      bookingType.value = type
    },
  })
}

function clearFormFields(type) {
  // 只清 type 专属字段，不碰共享字段（避免手滑切类型丢掉测试描述与设备备注）
  if (type === 'personal') {
    form.pName = ''
    form.pPhone = ''
    form.pModel = ''
    form.pLicenseUrl = ''
    certPreview.license = ''
  } else {
    form.gTeam = ''
    form.gPeople = ''
    form.gEquipList = ''
    form.gContactName = ''
    form.gContactPhone = ''
    form.gQualUrl = ''
    certPreview.qual = ''
  }
}

// 执照 / 资质上传：POST /api/v1/files/upload → /uploads/{file_id}
function chooseCert(key) {
  uni.chooseImage({
    count: 1,
    sourceType: ['album', 'camera'],
    success: function (res) {
      uploadCert(key, res.tempFilePaths[0])
    },
  })
}

async function uploadCert(key, filePath) {
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }
  uni.showLoading({ title: '上传中...' })
  try {
    const data = await new Promise(function (resolve, reject) {
      uni.uploadFile({
        url: BASE_URL + '/api/v1/files/upload',
        filePath: filePath,
        name: 'file',
        header: { Authorization: 'Bearer ' + token },
        success: function (r) {
          if (r.statusCode >= 200 && r.statusCode < 300) {
            try { resolve(JSON.parse(r.data)) } catch (e) { reject(e) }
          } else {
            reject(new Error('upload failed ' + r.statusCode))
          }
        },
        fail: reject,
      })
    })
    const fid = data && (data.file_id || (data.data && data.data.file_id))
    if (!fid) {
      uni.showToast({ title: '上传失败，请重试', icon: 'none' })
      return
    }
    const url = '/uploads/' + fid
    if (key === 'license') {
      form.pLicenseUrl = url
      certPreview.license = filePath
    } else {
      form.gQualUrl = url
      certPreview.qual = filePath
    }
  } catch (e) {
    uni.showToast({ title: '上传失败，请重试', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

function agreeNotice() {
  noticeAgreed.value = true
  noticeShow.value = false
}

// 须知弹窗出口：不认可可离开本页，而非被唯一「同意」按钮锁死
function declineNotice() {
  noticeShow.value = false
  setTimeout(function () { safeBack() }, 300)
}

async function fetchSite() {
  if (!siteId) {
    errorMsg.value = '场地参数缺失，请返回后重试'
    loading.value = false
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await request({ url: '/api/v1/test-sites/' + encodeURIComponent(siteId) })
    const d = (res && res.data) || res
    if (d && d.id) {
      site.value = d
      // 进入页面自动弹出预约须知（照原型）
      setTimeout(function () { noticeShow.value = true }, 300)
    } else {
      errorMsg.value = '场地不存在或已下架'
    }
  } catch (e) {
    const code = e && (e.statusCode || e.status)
    errorMsg.value = code === 404 ? '场地不存在或已下架' : '网络异常，请稍后重试'
  } finally {
    loading.value = false
  }
}

function onRetry() {
  if (retryLabel.value === '返回场地列表') {
    uni.reLaunch({ url: '/pkg-service/pages/testsites/list' })
    return
  }
  fetchSite()
}

async function handleSubmit() {
  if (submitting.value) return // 硬化：submit-btn 是 view 非原生 button，disabled 只改视觉——重入守卫防双击双 POST
  var token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  if (!siteId || !site.value) {
    uni.showToast({ title: '场地信息缺失', icon: 'none' })
    return
  }
  if (!canSubmit.value) {
    const miss = missingFields()
    const tip = miss.length
      ? '请补充：' + miss.slice(0, 3).join('、') + (miss.length > 3 ? ' 等' + miss.length + ' 项' : '')
      : '请完善预约信息'
    uni.showToast({ title: tip, icon: 'none' })
    return
  }

  const isGroup = bookingType.value === 'group'
  const contactName = isGroup ? form.gContactName : form.pName
  const contactPhone = isGroup ? form.gContactPhone : form.pPhone

  submitting.value = true
  try {
    // 提交字段与 docs/接口文档/场地预约接口变更需求.md 一致：
    // 新字段后端合入前被静默忽略（不报错），合入后即完整落库
    await request({
      url: '/api/v1/test-sites/' + encodeURIComponent(siteId) + '/book',
      method: 'POST',
      data: {
        purpose: form.desc,
        date: form.date,
        time_slot: form.timeSlots[0] || '',
        time_slots: form.timeSlots,
        booking_type: bookingType.value,
        model: form.pModel,
        license_url: form.pLicenseUrl,
        team_name: form.gTeam,
        people_count: form.gPeople ? Number(form.gPeople) : 0,
        equipment_list: form.gEquipList,
        qualification_url: form.gQualUrl,
        equipment_note: form.equipmentNote,
        contact_name: contactName,
        contact_phone: contactPhone,
      },
    })
    // 保存预约摘要，供确认页（pay）展示
    uni.setStorageSync('testBookingDraft', {
      siteId: siteId,
      siteName: site.value.name,
      booking_type: bookingType.value,
      date: form.date,
      time_slot: form.timeSlots.join('、'),
      time_slots: form.timeSlots,
      purpose: form.desc,
      contact_name: contactName,
      contact_phone: contactPhone,
      price_fen: site.value.price_fen,
    })
    // 保栈跳转（评审 P1-2）：pay 返回落回本页，已填表单不销毁
    uni.navigateTo({ url: '/pkg-service/pages/testsites/pay' })
  } catch (e) {
    const code = e && (e.statusCode || e.status)
    if (code === 409) {
      uni.showToast({ title: '该时段已被预约，请换一个时段', icon: 'none' })
    } else {
      uni.showToast({ title: '预约提交失败，请稍后重试', icon: 'none' })
    }
  } finally {
    submitting.value = false
  }
}

function goBack() {
  safeBack()
}

// 从 testBookingDraft 回填表单（评审 P1-2/P2：从 pay/result 重开或重试时表单不丢；?date= 深链优先，不覆盖）
function restoreDraft(hasDeepLinkDate) {
  let stored = null
  try {
    stored = uni.getStorageSync('testBookingDraft')
  } catch (e) {
    stored = null
  }
  if (!stored || !stored.siteId || stored.siteId !== siteId) return
  if (stored.booking_type === 'personal' || stored.booking_type === 'group') bookingType.value = stored.booking_type
  if (!hasDeepLinkDate && stored.date) form.date = stored.date
  if (Array.isArray(stored.time_slots) && stored.time_slots.length) form.timeSlots = stored.time_slots.slice()
  if (stored.purpose) form.desc = stored.purpose
  const isG = bookingType.value === 'group'
  if (stored.contact_name) {
    if (isG) form.gContactName = stored.contact_name
    else form.pName = stored.contact_name
  }
  if (stored.contact_phone) {
    if (isG) form.gContactPhone = stored.contact_phone
    else form.pPhone = stored.contact_phone
  }
}

onLoad((options) => {
  checkMotion()
  siteId = (options && options.id) || ''
  // 预填日期：detail 日期格可带 ?date= 直达；早于今天（陈旧链接）回退到今天
  const optDate = (options && options.date) || ''
  const minDate = today()
  form.date = optDate >= minDate ? optDate : ''
  // 草稿回填：仅同场地；深链带日期时草稿日期不覆盖
  restoreDraft(!!optDate)

  var token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(function () {
      uni.navigateTo({ url: '/pages/login/index' })
    }, 500)
    return
  }
  // 预填联系电话（微信登录用户可能无手机号，留空需手动填写）
  var u = getStoredUser()
  if (u && u.phone) {
    form.pPhone = u.phone
  }
  fetchSite()
})
</script>

<style scoped>
.tsb-page {
  min-height: 100vh;
  background: #fff; /* 白上白：白底页面 + 描边软角卡片（对齐组） */
  padding-bottom: 240rpx; /* 固定底栏之上留呼吸，滚动到底不被遮 */
}

/* ===== 骨架（对齐组） ===== */
.sk-section {
  background: #fff;
  margin: 16rpx 24rpx 0;
  padding: 32rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
}
.sk-l {
  height: 24rpx;
  background: #EDF0F3;
  border-radius: 8rpx;
  margin-bottom: 20rpx;
  animation: sk-pulse 1.4s linear infinite;
}
.sk-l.w40 { width: 40%; }
.sk-l.w50 { width: 50%; }
.sk-l.w60 { width: 60%; }
.sk-l.w70 { width: 70%; }
.sk-l.w80 { width: 80%; }
.sk-l.w90 { width: 90%; }
@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }

/* ===== 错误态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.stb {
  margin-top: 32rpx;
  min-height: 88rpx;
  padding: 0 48rpx;
  border-radius: 16rpx; /* 对齐组按钮：16rpx，非全圆 */
  background: #0A66C2;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
}
.stb:active { opacity: 0.85; }

/* ===== 白上白卡片（2rpx 描边 + 20rpx 软角 + 蓝调双层阴影，与 detail/list 同构） ===== */
.step-card,
.notice-block {
  background: #fff;
  margin: 16rpx 24rpx 0;
  padding: 28rpx 32rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx;
  box-shadow:
    0 2rpx 6rpx rgba(10, 30, 60, 0.04),
    0 12rpx 32rpx rgba(10, 30, 60, 0.05);
  animation: fade-in 0.22s ease-out backwards;
}
.step-card:nth-of-type(2) { animation-delay: 20ms; }
.step-card:nth-of-type(3) { animation-delay: 40ms; }
.step-card:nth-of-type(4) { animation-delay: 60ms; }
.step-card:nth-of-type(5) { animation-delay: 60ms; }
.notice-block {
  margin-bottom: 0;
  animation-delay: 60ms;
}
@keyframes fade-in {
  from { opacity: 0; transform: translateY(16rpx); }
  to { opacity: 1; transform: translateY(0); }
}

/* ===== 预约摘要条（蓝雾） ===== */
.sum-bar {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8rpx;
  margin: 16rpx 24rpx 0;
  padding: 20rpx 24rpx;
  background: #EAF3FB; /* 蓝雾浅底（对齐组 tint） */
  border-radius: 20rpx;
  font-size: 26rpx;
  color: #0A66C2;
  animation: fade-in 0.22s ease-out backwards;
}
.sum-name {
  font-weight: 600;
  max-width: 50%; /* 硬化：超长场地名截断，不把「未选日期/时段」挤到下一行 */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sum-sep {
  opacity: 0.55;
}
.sum-pick {
  color: #074D92; /* 深蓝：蓝雾底上的数据强调（AA 深色变体） */
  font-weight: 600;
  transition: color 0.2s ease-out; /* 选择反馈：未选灰 → 已选深蓝平滑（delight，非瞬跳） */
}
.sum-pick.pending {
  color: #667085;
  font-weight: 400;
}

/* ===== 步骤标题（原型 step-number 圆圈，承载流程顺序） ===== */
.step-label {
  display: flex;
  align-items: center;
  gap: 12rpx;
  font-size: 30rpx; /* 对齐组例外：卡片标题 */
  font-weight: 700;
  color: #17212B;
  margin-bottom: 20rpx;
}
.step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: #0A66C2;
  color: #fff;
  font-size: 24rpx;
  font-weight: 700;
  flex-shrink: 0;
}

/* ===== Step 1 预约类型：二选一卡片 ===== */
.type-grid {
  display: flex;
  gap: 16rpx;
}
.type-card {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  padding: 28rpx 16rpx 24rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 20rpx; /* 与对齐组卡片同档 */
  background: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  /* 规范（Motion-License）：属性限定 + 短过渡，禁 all */
  transition: border-color 0.15s ease-out, background-color 0.15s ease-out;
}
.type-card.selected {
  border-color: #0A66C2;
  background: #EAF3FB; /* 蓝雾选中（与 slot/日期格同语言） */
}
.type-icon {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12rpx;
  color: #98A2B3;
  transition: color 0.15s ease-out;
}
.type-card.selected .type-icon { color: #0A66C2; }
.type-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #17212B;
  margin-bottom: 6rpx;
}
.type-desc {
  font-size: 24rpx;
  color: #667085;
  line-height: 1.5;
}

/* CSS 绘制人形（设计系统禁 emoji；单人 + 双人） */
.fig-single {
  position: relative;
}
.fig-single .fig-head {
  position: absolute;
  top: 6rpx;
  left: 50%;
  margin-left: -9rpx;
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
  background: currentColor;
}
.fig-single .fig-body {
  position: absolute;
  bottom: 8rpx;
  left: 50%;
  margin-left: -16rpx;
  width: 32rpx;
  height: 22rpx;
  border-radius: 16rpx 16rpx 0 0;
  background: currentColor;
}
.fig-group {
  position: relative;
}
.fig-group .fig {
  position: absolute;
  width: 30rpx;
  height: 38rpx;
}
.fig-group .fig-a {
  left: 4rpx;
  bottom: 8rpx;
}
.fig-group .fig-b {
  right: 4rpx;
  bottom: 8rpx;
}
.fig-group .fig-head {
  position: absolute;
  top: 0;
  left: 50%;
  margin-left: -7rpx;
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  background: currentColor;
}
.fig-group .fig-body {
  position: absolute;
  bottom: 0;
  left: 50%;
  margin-left: -13rpx;
  width: 26rpx;
  height: 18rpx;
  border-radius: 12rpx 12rpx 0 0; /* 阶梯值：13rpx 不在字阶，12rpx 渲染同 */
  background: currentColor;
}

/* ===== 日期 / 时段选择卡（原型 pick-card） ===== */
.pick-card {
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx; /* 内嵌选择卡：field 档 */
  padding: 20rpx 20rpx 24rpx;
}
.pick-card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}
.pc-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
}
.pc-hint {
  font-size: 24rpx;
  color: #667085;
}

/* 日期 chips（今天起 4 天）+ 日历按钮（与 detail 日期格同族） */
.date-chips {
  display: flex;
  gap: 12rpx;
}
.date-chip,
.cal-btn {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  min-height: 88rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  /* 规范（Motion-License）：属性限定 + 短过渡 */
  transition: border-color 0.15s ease-out, background-color 0.15s ease-out, color 0.15s ease-out;
}
.date-chip.selected {
  border-color: #0A66C2;
  background: #EAF3FB;
  color: #0A66C2;
}
.dc-day {
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
}
.date-chip.selected .dc-day { color: #0A66C2; }
.dc-week {
  font-size: 20rpx;
  color: #667085;
}
.date-chip.selected .dc-week { color: #0A66C2; }

.cal-btn {
  border-color: #E4E7EC; /* 实线（Solid-Line：正常态禁虚线） */
  color: #0A66C2;
  font-size: 24rpx;
  font-weight: 600;
}
.cal-icon {
  width: 30rpx;
  height: 28rpx;
  border: 2rpx solid currentColor;
  border-radius: 8rpx; /* 阶梯值：6rpx 不在阶梯，8rpx 渲染同 */
  position: relative;
  box-sizing: border-box;
}
.cal-icon::before {
  content: '';
  position: absolute;
  left: 4rpx;
  right: 4rpx;
  top: 9rpx;
  height: 2rpx;
  background: currentColor;
}

/* 时段：多选格（选中 = 蓝描边 + 蓝雾底 + 蓝字） */
.slot-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.slot-item {
  flex: none;
  width: calc((100% - 24rpx) / 3);
  box-sizing: border-box;
  min-height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid #E4E7EC;
  border-radius: 12rpx; /* 与日期格同族：field 档 */
  font-size: 26rpx;
  color: #344054;
  background: #fff;
  /* 规范（Motion-License）：属性限定 + 短过渡 */
  transition: border-color 0.15s ease-out, background-color 0.15s ease-out, color 0.15s ease-out;
}
.slot-item.selected {
  border-color: #0A66C2;
  background: #EAF3FB;
  color: #0A66C2;
  font-weight: 600;
}
.form-hint {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #667085;
}

/* ===== 表单（对齐组 input-field 规范：#fafafa 浅底 + 16rpx） ===== */
.type-form {
  border-bottom: 2rpx solid #f0f1f3; /* 实线分隔：类型差异区 */
  padding-bottom: 24rpx;
  margin-bottom: 24rpx;
}
.form-group {
  margin-bottom: 24rpx;
}
.form-group:last-child {
  margin-bottom: 0;
}
.form-label {
  display: flex;
  align-items: center;
  gap: 6rpx;
  font-size: 26rpx;
  font-weight: 600;
  color: #344054;
  margin-bottom: 12rpx;
}
.required {
  color: #D92D20;
  font-size: 24rpx;
}
.optional {
  font-size: 24rpx;
  font-weight: 400;
  color: #98A2B3;
}
.form-input {
  width: 100%;
  box-sizing: border-box;
  min-height: 88rpx;
  padding: 0 24rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx; /* input-field 规范 */
  font-size: 28rpx;
  color: #17212B;
  background: #fafafa;
}
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 160rpx;
  padding: 20rpx 24rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  font-size: 28rpx;
  color: #17212B;
  background: #fafafa;
  line-height: 1.6;
}
.form-textarea.short {
  min-height: 100rpx;
}
.ph {
  color: #98A2B3;
}
.form-error {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #D92D20;
}
.char-count {
  display: block;
  text-align: right;
  font-size: 24rpx;
  color: #98A2B3;
  margin-top: 8rpx;
}

/* 上传区（实线框——Solid-Line 禁虚线；上传后显示缩略图） */
.upload-zone {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  min-height: 120rpx;
  border: 2rpx solid #E4E7EC;
  border-radius: 16rpx;
  background: #fafafa;
  padding: 16rpx 24rpx;
}
.up-img {
  width: 120rpx;
  height: 120rpx;
  border-radius: 8rpx;
  flex-shrink: 0;
}
.up-plus {
  font-size: 40rpx;
  color: #98A2B3;
  line-height: 1;
}
.up-tip {
  font-size: 26rpx;
  color: #667085;
}
.up-note {
  font-size: 24rpx;
  color: #0A66C2;
}

/* ===== 资金说明（蓝雾） ===== */
.notice-block {
  background: #EAF3FB; /* 蓝雾浅底（对齐组 tint） */
  border: none;
  box-shadow: none;
}
.notice-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #0A66C2;
  display: block;
  margin-bottom: 8rpx;
}
.notice-line {
  display: block;
  font-size: 24rpx;
  color: #344054;
  line-height: 1.7;
}

/* ===== 预约须知弹窗 ===== */
.notice-dialog {
  padding: 40rpx 32rpx 32rpx;
  text-align: center;
}
.nd-title {
  display: block;
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 8rpx;
}
.nd-sub {
  display: block;
  font-size: 24rpx;
  color: #667085;
  margin-bottom: 20rpx;
}
.notice-list {
  text-align: left;
  background: #EAF3FB; /* 蓝雾浅底 */
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
  margin-bottom: 24rpx;
}
.nl-item {
  font-size: 24rpx;
  color: #344054;
  line-height: 1.9;
  border-bottom: 2rpx solid rgba(10, 102, 194, 0.1);
}
.nl-item:last-child {
  border-bottom: none;
}
.nd-actions {
  display: flex;
  gap: 16rpx;
}

.nd-btn {
  flex: 1;
  height: 88rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮 */
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.nd-btn:active { transform: scale(0.96); }

/* 须知弹窗次要出口：描边按钮（对齐组 outline 配方） */
.nd-btn.ghost {
  background: #fff;
  color: #0A66C2;
  border: 2rpx solid #0A66C2;
  box-shadow: none;
}

/* ===== 完整日历 sheet ===== */
.cal-sheet {
  padding: 12rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
  max-height: 78vh;
  overflow-y: auto;
}
.sheet-hd {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8rpx;
}
.sheet-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
}
.sheet-close {
  font-size: 32rpx;
  color: #98A2B3;
  padding: 8rpx;
  line-height: 1;
}
.cal-month-block {
  margin-top: 16rpx;
}
.cal-month {
  display: block;
  text-align: center;
  font-size: 26rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 8rpx;
}
.cal-week {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  margin-bottom: 4rpx;
}
.cal-week-cell {
  font-size: 20rpx;
  color: #667085;
  padding: 8rpx 0;
}
.cal-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  row-gap: 4rpx;
}
.cal-cell {
  min-height: 88rpx; /* 触达标准（评估 Sam：72rpx < 88rpx 标准） */
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26rpx;
  color: #344054;
  border-radius: 50%;
  box-sizing: border-box;
}
.cal-cell.empty { pointer-events: none; }
.cal-cell.past {
  color: #C1C7D0;
}
.cal-cell.selected {
  background: #0A66C2;
  color: #fff;
  font-weight: 600;
}
.cal-confirm {
  margin-top: 20rpx;
  height: 88rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮 */
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.cal-confirm:active { transform: scale(0.96); }

/* ===== 底部提交栏 ===== */
.submit-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24rpx 32rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 2rpx solid #f0f1f3;
  box-shadow: 0 -4rpx 20rpx rgba(10, 30, 60, 0.05);
  z-index: 50;
}
.submit-btn {
  height: 88rpx;
  border-radius: 16rpx; /* 对齐组主行动按钮：16rpx */
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 600;
  box-shadow: inset 0 2rpx 0 rgba(255,255,255,.22), inset 0 -4rpx 10rpx rgba(7,77,146,.18), 0 4rpx 14rpx rgba(10,102,194,.25);
  transition: transform 0.15s ease-out;
}
.submit-btn:active { transform: scale(0.96); }
.submit-btn.disabled {
  background: #EEF1F4;
  color: #5D6B82; /* ≈4.9:1，禁用原因可读 */
  box-shadow: none;
}
.submit-btn.disabled:active { transform: none; }

/* 减弱动效（无障碍）：装饰动画与过渡全关 */
.no-motion .sk-l,
.no-motion .sum-bar,
.no-motion .step-card,
.no-motion .notice-block { animation: none; }
.no-motion .type-card,
.no-motion .date-chip,
.no-motion .slot-item { transition: none; }
</style>
