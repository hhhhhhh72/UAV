<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar :title="done ? '申请结果' : '展位申请'" show-back :fixed="true" @back="goBack" />

    <!-- ═══ 成功态 ═══ -->
    <view v-if="done" class="succ">
      <view class="succ-ring"><u-icon name="success" size="40px" color="#fff" /></view>
      <text class="succ-title">展位申请已提交</text>
      <text class="succ-desc">申请信息已提交，审核结果将通过消息中心通知
如需更换展位，请在审核通过前联系协会秘书处</text>
      <view class="receipt">
        <view class="rt">申请回执 <text class="rt-no">{{ receiptNo }}</text></view>
        <view class="rrow"><text>展会</text><text class="rv">{{ summary.title }}</text></view>
        <view class="rrow"><text>展位号</text><text class="rv">{{ summary.booth }}</text></view>
        <view class="rrow"><text>展品</text><text class="rv">{{ summary.exhibit }}</text></view>
        <view class="rrow"><text>状态</text><text class="rv cl-su">待审核</text></view>
      </view>
      <view class="btn-main" @tap="goList">返回展会列表</view>
      <view class="btn-ghost" @tap="onViewMine">查看我的申请</view>
    </view>

    <!-- ═══ 表单态 ═══ -->
    <template v-else>
      <!-- 加载 -->
      <view v-if="loading" class="skw">
        <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w60"></view></view>
        <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w70"></view></view>
      </view>

      <!-- 错误 -->
      <view v-else-if="err || !detail" class="st">
        <u-empty :description="err ? '加载失败，请检查网络' : '该展会不存在'">
          <view v-if="err" class="stb" @tap="fetchData">重新加载</view>
          <view v-else class="stb" @tap="goBack">返回</view>
        </u-empty>
      </view>

      <template v-else>
        <!-- 展会摘要 -->
        <view class="a-sum">
          <text class="a-title">{{ detail.title }}</text>
          <view class="a-meta">
            <text>{{ detail.dateShort }}</text>
            <text>{{ detail.location }}</text>
            <text>{{ detail.priceText || '' }}</text>
          </view>
        </view>

        <!-- 选择展位 -->
        <view class="sec-title">
          <text>选择展位</text>
          <view class="legend">
            <text><view class="dot ok"></view>可选</text>
            <text><view class="dot no"></view>已订</text>
            <text><view class="dot sel"></view>已选</text>
          </view>
        </view>
        <view v-if="detail.cells.length" class="b-grid">
          <view
            v-for="c in detail.cells"
            :key="c.no"
            class="b-cell"
            :class="cellClass(c)"
            @tap="pickBooth(c)"
          >{{ c.no }}</view>
        </view>
        <view v-else class="no-booth">展位信息待主办方更新</view>
        <text v-if="detail.capped" class="b-note">仅展示部分展位，完整展位以现场平面图为准</text>

        <!-- 申请信息（对齐后端契约：booth_number / exhibit_name / exhibit_desc） -->
        <view class="form-card">
          <view class="form-group">
            <view class="form-label">展位号 <text class="required">*</text><text class="form-hint">点击上方展位图选择</text></view>
            <input class="form-input readonly" v-model="form.booth_number" placeholder="请先选择展位" placeholder-class="ph" disabled />
          </view>
          <view class="form-group">
            <view class="form-label">展品名称 <text class="required">*</text></view>
            <input class="form-input" v-model="form.exhibit_name" maxlength="40" placeholder="如：X-200 工业级无人机" placeholder-class="ph" />
          </view>
          <view class="form-group">
            <view class="form-label">展品简介 <text class="required">*</text></view>
            <textarea class="form-textarea" v-model="form.exhibit_desc" maxlength="200" placeholder="一句话介绍展品亮点、参数或解决方案" placeholder-class="ph" />
          </view>
        </view>

        <view class="note-box">
          <text class="note-title">申请须知</text>
          <text class="note-line">· 提交后进入协会审核，审核结果通过消息中心通知</text>
          <text class="note-line">· 企业名称与联系人信息取自企业档案，无需重复填写</text>
          <text class="note-line">· 展位一经提交不可更换，请核对展品信息后再提交</text>
        </view>
        <view style="height: 130px"></view>

        <!-- 底部操作栏：已选展位信息 + 提交 -->
        <view class="bb booth-bar">
          <view class="bb-info">
            <text class="bb-info-label">已选展位</text>
            <text class="bb-info-value">{{ selected ? selected.no : '请选择' }}</text>
            <text class="bb-info-price">{{ detail.priceText || '' }} / 标准展位</text>
          </view>
          <view class="bp" :class="{ disabled: !selected || submitting }" hover-class="tap-fade" @tap="handleSubmit">
            {{ submitting ? '提交中...' : '提交申请' }}
          </view>
        </view>
      </template>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage } from '@/utils/request'
import { MOCK_EXHIBITIONS, MOCK_BOOTHS_BY_EXPO, fmtRange, fmtFen, buildBoothCells } from '@/utils/mockExhibitions'

const id = ref('')
const detail = ref(null)
const loading = ref(true)
const err = ref(false)
const done = ref(false)
const submitting = ref(false)
const selected = ref(null)
const statusBarHeight = ref(20)
const receiptNo = ref('')
const summary = ref({ title: '', booth: '', exhibit: '' })

const form = ref({ booth_number: '', exhibit_name: '', exhibit_desc: '' })

const cellClass = (c) => (c.occupied ? 'no' : selected.value && selected.value.no === c.no ? 'sel' : 'ok')

// ===== 数据获取 =====
// 接口替换点：GET /api/v1/exhibitions/{id} + GET /api/v1/exhibitions/{id}/booths
const fetchData = async () => {
  if (!id.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const [expoRes, boothRes] = await Promise.all([
      request({ url: '/api/v1/exhibitions/' + encodeURIComponent(id.value) }),
      request({ url: '/api/v1/exhibitions/' + encodeURIComponent(id.value) + '/booths' }).catch(() => []),
    ])
    const it = expoRes && expoRes.id ? expoRes : null
    if (it) {
      detail.value = buildDetail(it, Array.isArray(boothRes) ? boothRes : [])
      return
    }
    throw new Error('empty')
  } catch {
    // 回退：演示数据 / 列表缓存（仅开发环境回退演示数据）
    const cached = uni.getStorageSync('exhibition_cache_' + id.value)
    if (import.meta.env.DEV) {
      const mock = MOCK_EXHIBITIONS.find((x) => x.id === id.value)
      const src = mock || cached
      if (src) {
        detail.value = buildDetail(src, MOCK_BOOTHS_BY_EXPO[id.value] || [])
      } else {
        err.value = true
      }
    } else if (cached) {
      detail.value = buildDetail(cached, [])
    } else {
      err.value = true
    }
  } finally {
    loading.value = false
  }
}

const buildDetail = (it, booths = []) => {
  const occupied = new Set()
  ;(booths || []).forEach((b) => {
    if (b && b.booth_number && b.status !== 'rejected') occupied.add(String(b.booth_number).toLowerCase())
  })
  const { cells, capped } = buildBoothCells(it.booth_count, occupied)
  const start = String(it.start_date || '')
  const end = String(it.end_date || '')
  return {
    id: it.id,
    title: it.title || '',
    dateShort: fmtRange(start, end),
    location: it.location || '',
    priceText: fmtFen(it.booth_price_fen),
    cells,
    capped,
  }
}

// ===== 选位 =====
const pickBooth = (c) => {
  if (c.occupied) {
    uni.showToast({ title: '该展位已被预订，请选择其他展位', icon: 'none' })
    return
  }
  selected.value = c
  form.value.booth_number = c.no
}

// ===== 提交 =====
// 接口替换点：POST /api/v1/exhibitions/{id}/booths (body: booth_number / exhibit_name / exhibit_desc)
const handleSubmit = async () => {
  if (!selected.value) {
    uni.showToast({ title: '请先选择展位', icon: 'none' })
    return
  }
  if (!form.value.exhibit_name.trim()) {
    uni.showToast({ title: '请填写展品名称', icon: 'none' })
    return
  }
  if (!form.value.exhibit_desc.trim()) {
    uni.showToast({ title: '请填写展品简介', icon: 'none' })
    return
  }
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(() => uni.navigateTo({ url: '/pages/login/index' }), 500)
    return
  }
  submitting.value = true
  try {
    await request({
      url: '/api/v1/exhibitions/' + encodeURIComponent(id.value) + '/booths',
      method: 'POST',
      data: {
        booth_number: form.value.booth_number,
        exhibit_name: form.value.exhibit_name.trim(),
        exhibit_desc: form.value.exhibit_desc.trim(),
      },
    })
    receiptNo.value = 'B' + String(Date.now()).slice(-10)
    summary.value = {
      title: detail.value.title,
      booth: form.value.booth_number,
      exhibit: form.value.exhibit_name.trim(),
    }
    done.value = true
  } catch {
    uni.showToast({ title: '提交失败，请重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

// ===== 其他 =====
const goList = () => {
  const pages = getCurrentPages()
  if (pages.length > 2) uni.navigateBack({ delta: 2 })
  else uni.navigateBack()
}
const onViewMine = () => uni.showToast({ title: '我的申请功能建设中', icon: 'none' })
const goBack = () => uni.navigateBack()

onLoad((options) => {
  if (options && options.id) id.value = decodeURIComponent(options.id)
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  fetchData()
})
</script>

<style scoped>
.page { min-height: 100vh; background: #F4F6F8; padding-bottom: 40rpx; }

/* ===== 摘要 ===== */
.a-sum { background: #EAF3FB; border: 1px solid #D5E6F8; border-radius: 20rpx; padding: 28rpx; margin: 24rpx; }
.a-title { font-size: 30rpx; font-weight: 700; color: #17212B; line-height: 1.4; display: block; }
.a-meta { display: flex; gap: 24rpx; margin-top: 12rpx; font-size: 22rpx; color: #667085; flex-wrap: wrap; }

/* ===== 选择展位 ===== */
.sec-title { font-size: 30rpx; font-weight: 700; color: #17212B; display: flex; align-items: center; justify-content: space-between; padding: 0 32rpx; margin: 28rpx 0 16rpx; }
.legend { display: flex; gap: 24rpx; font-size: 20rpx; color: #667085; font-weight: 400; align-items: center; }
.legend .dot { display: inline-block; width: 18rpx; height: 18rpx; border-radius: 4rpx; margin-right: 6rpx; vertical-align: middle; }
.legend .dot.ok { background: #34C759; }
.legend .dot.no { background: #D8DDE3; }
.legend .dot.sel { background: #0A66C2; }
.b-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 16rpx; padding: 0 32rpx; }
.b-cell { aspect-ratio: 1.4; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; font-size: 20rpx; font-weight: 600; color: #fff; transition: transform .15s, box-shadow .15s; }
.b-cell.ok { background: linear-gradient(135deg,#2ea44f,#34c759); }
.b-cell.no { background: #D8DDE3; color: #98A2B3; }
.b-cell.sel { background: linear-gradient(135deg,#0d47a1,#1565c0); box-shadow: 0 0 0 4rpx rgba(10,102,194,.35); transform: scale(1.04); }
.b-cell.ok:active { transform: scale(.94); }
.b-note { display: block; padding: 12rpx 32rpx 0; font-size: 20rpx; color: #98A2B3; }
.no-booth { padding: 40rpx 32rpx; font-size: 24rpx; color: #98A2B3; }

/* ===== 表单 ===== */
.form-card { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; margin: 20rpx 24rpx 0; padding: 32rpx; }
.form-group { margin-bottom: 28rpx; }
.form-group:last-child { margin-bottom: 0; }
.form-label { display: flex; align-items: center; font-size: 26rpx; font-weight: 600; color: #344054; margin-bottom: 12rpx; }
.required { color: #D92D20; font-size: 24rpx; margin-left: 4rpx; }
.form-hint { font-size: 20rpx; color: #98A2B3; font-weight: 400; margin-left: 16rpx; }
.form-input { width: 100%; box-sizing: border-box; min-height: 84rpx; padding: 20rpx 24rpx; border: 1px solid #E4E7EC; border-radius: 16rpx; font-size: 28rpx; color: #17212B; background: #fff; }
.form-input.readonly { background: #F4F6F8; color: #667085; }
.form-textarea { width: 100%; box-sizing: border-box; min-height: 152rpx; padding: 20rpx 24rpx; border: 1px solid #E4E7EC; border-radius: 16rpx; font-size: 28rpx; color: #17212B; background: #fff; }
.ph { color: #98A2B3; }

/* ===== 须知 ===== */
.note-box { background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; margin: 20rpx 24rpx 0; padding: 24rpx 28rpx; }
.note-title { font-size: 24rpx; font-weight: 600; color: #0A66C2; margin-bottom: 12rpx; display: block; }
.note-line { font-size: 22rpx; color: #667085; line-height: 1.7; display: block; }

/* ===== 底部操作栏 ===== */
.bb { position: fixed; left: 0; right: 0; bottom: 0; display: flex; align-items: center; padding: 16rpx 24rpx; padding-bottom: calc(16rpx + env(safe-area-inset-bottom)); background: #fff; border-top: 1rpx solid #EEF1F4; gap: 20rpx; z-index: 50; }
.bb-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.bb-info-label { font-size: 20rpx; color: #98A2B3; }
.bb-info-value { font-size: 30rpx; font-weight: 700; color: #0A66C2; }
.bb-info-price { font-size: 22rpx; color: #667085; }
.bp { flex: none; width: 400rpx; height: 88rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 30rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.35); }
.bp.disabled { background: #98A2B3; box-shadow: none; }

/* ===== 成功态 ===== */
.succ { padding: 80rpx 48rpx; display: flex; flex-direction: column; align-items: center; text-align: center; }
.succ-ring { width: 176rpx; height: 176rpx; border-radius: 50%; background: linear-gradient(135deg, #34c759, #5bd88a); display: flex; align-items: center; justify-content: center; box-shadow: 0 28rpx 72rpx rgba(52,199,89,.35); margin-bottom: 44rpx; animation: pop .45s cubic-bezier(0.16, 1, 0.3, 1) both; }
@keyframes pop { from { transform: scale(.5); opacity: 0; } to { transform: scale(1); opacity: 1; } }
.succ-title { font-size: 42rpx; font-weight: 700; color: #17212B; margin-bottom: 16rpx; }
.succ-desc { font-size: 26rpx; color: #667085; line-height: 1.7; margin-bottom: 36rpx; white-space: pre-line; }
.receipt { width: 100%; background: #fff; border: 1px solid #EEF1F4; border-radius: 20rpx; padding: 32rpx; margin-bottom: 44rpx; text-align: left; box-sizing: border-box; }
.rt { font-size: 22rpx; color: #667085; margin-bottom: 20rpx; padding-bottom: 20rpx; border-bottom: 2rpx dashed #EBEDF0; display: flex; justify-content: space-between; }
.rt-no { color: #0A66C2; font-weight: 600; }
.rrow { display: flex; justify-content: space-between; font-size: 26rpx; padding: 10rpx 0; color: #667085; gap: 24rpx; }
.rv { color: #17212B; font-weight: 600; text-align: right; flex: 1; }
.cl-su { color: #168A55; }
.btn-main { width: 100%; height: 92rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 600; box-shadow: 0 4rpx 16rpx rgba(10,102,194,.35); }
.btn-ghost { width: 100%; height: 92rpx; border-radius: 16rpx; border: 3rpx solid #0A66C2; background: #fff; color: #0A66C2; display: flex; align-items: center; justify-content: center; font-size: 30rpx; font-weight: 600; margin-top: 20rpx; box-sizing: border-box; }

/* ===== 骨架 ===== */
.skw { padding-top: 20rpx; }
.sk-sec { margin: 24rpx; padding: 32rpx; background: #fff; border-radius: 24rpx; }
.sk-l { height: 28rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
.sk-l.w100 { width: 100%; }
.sk-l.w70 { width: 70%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 48rpx; min-height: 600rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 24rpx; }
</style>
