<template>
  <view class="wod-page">
    <view v-if="loading" class="wod-state">加载中...</view>

    <template v-else-if="w">
      <view class="wod-hero" :class="'wod-hero--' + w.status">
        <text class="wod-status-label">{{ statusLabel(w.status) }}</text>
        <text class="wod-no">{{ w.order_no || shortId(w.id) }}</text>
        <text class="wod-amount">{{ amountText(w) }}</text>
      </view>

      <view class="wod-card">
        <view class="wod-row"><text class="wod-k">作业需求</text><text class="wod-v">{{ w.demand_title || w.demand_id || '-' }}</text></view>
        <view class="wod-row"><text class="wod-k">需求方</text><text class="wod-v">{{ w.publisher_name || w.publisher_id }}</text></view>
        <view class="wod-row"><text class="wod-k">接单方</text><text class="wod-v">{{ w.worker_name || w.worker_id }}</text></view>
        <view class="wod-row"><text class="wod-k">创建时间</text><text class="wod-v">{{ formatTime(w.created_at) }}</text></view>
        <view v-if="w.rework_note" class="wod-row"><text class="wod-k">整改要求</text><text class="wod-v wod-v--warn">{{ w.rework_note }}</text></view>
        <view v-if="w.cancel_reason" class="wod-row"><text class="wod-k">取消原因</text><text class="wod-v wod-v--danger">{{ w.cancel_reason }}</text></view>
      </view>

      <view v-if="w.result_photos && w.result_photos.length" class="wod-card">
        <text class="wod-card-title">作业成果</text>
        <view class="wod-photos">
          <image
            v-for="(p, i) in w.result_photos"
            :key="i"
            class="wod-photo"
            :src="resolveImg(p)"
            mode="aspectFill"
            @tap="previewPhoto(i)"
          />
        </view>
      </view>

      <!-- 操作区 -->
      <view v-if="actions.length" class="wod-card">
        <text class="wod-card-title">操作</text>
        <view v-for="(a, i) in actions" :key="i" class="wod-btn" :class="{ primary: a.primary, danger: a.danger }" hover-class="wod-btn-hover" @tap="a.run">{{ a.label }}</view>
      </view>

      <view v-if="w.status === 'completed'" class="wod-done">订单已完成 · 互评功能 V1.1 开放</view>
    </template>

    <view v-else class="wod-state">工单不存在或无权查看</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, getStoredUser, getErrorMessage, authStorage, BASE_URL } from '../../utils/request'

const id = ref('')
const w = ref(null)
const loading = ref(true)
const needRefresh = ref(false)

const user = getStoredUser()
const myId = user && (user.id || user.user_id)
const isPub = computed(() => w.value && myId && w.value.publisher_id === myId)
const isWorker = computed(() => w.value && myId && w.value.worker_id === myId)

const shortId = (v) => (v || '').length > 10 ? (v || '').slice(-8) : (v || '-')
const statusLabel = (s) => ({ pending: '待开始', ongoing: '进行中', awaiting_accept: '待验收', completed: '已完成', cancelled: '已取消' }[s] || s || '')
const amountText = (wd) => (wd && wd.amount_fen && wd.amount_fen > 0 ? '¥' + ((wd.amount_fen / 100).toLocaleString()) + ' 元' : '金额面议')
const formatTime = (iso) => (iso || '').slice(0, 16).replace('T', ' ')
const resolveImg = (p) => (p && p.indexOf('/') === 0 ? BASE_URL + p : p)

const previewPhoto = (i) => uni.previewImage({ urls: (w.value.result_photos || []).map(resolveImg), current: resolveImg(w.value.result_photos[i]) })

const toast = (t) => uni.showToast({ title: t, icon: 'none' })

async function act(fn, doneMsg) {
  try {
    await fn()
    toast(doneMsg || '操作成功')
    needRefresh.value = true
    load()
  } catch (e) {
    toast(getErrorMessage(e) || '操作失败，请重试')
  }
}

const uploadPhoto = (localPath) => {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + '/api/v1/files/upload',
      filePath: localPath,
      name: 'file',
      formData: { private: 'false' },
      header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
      success: (r) => {
        let data = null
        try { data = JSON.parse(r.data) } catch (e) {}
        if (r.statusCode >= 400 || !data || (!data.file_id && !(data.data && data.data.file_id))) { reject(new Error('上传失败')); return }
        const fid = data.file_id || (data.data && data.data.file_id)
        const url = data.url || (data.data && data.data.url) || ('/uploads/' + fid)
        resolve(url.startsWith('/') ? url : '/' + url.replace(/^.*?\/uploads\//, 'uploads/'))
      },
      fail: (e) => reject(new Error('上传失败，请重试'))
    })
  })
}

let chosenPhotos = []
const chooseAndUpload = async () => {
  try {
    const paths = await new Promise((resolve, reject) => uni.chooseImage({ count: 3, sizeType: ['compressed'], success: (r) => resolve(r.tempFilePaths), fail: (e) => reject(e) }))
    uni.showLoading({ title: '上传中...', mask: true })
    chosenPhotos = []
    for (const p of paths) chosenPhotos.push(await uploadPhoto(p))
    return true
  } catch (e) {
    uni.hideLoading()
    toast((e && e.message) || '上传失败，请重试')
    return false
  } finally {
    uni.hideLoading()
  }
}

const actions = computed(() => {
  if (!w.value) return []
  const acts = []
  const st = w.value.status
  // 飞手：待开始 → 开始作业；进行中 → 完成作业（可上传成果照片）
  if (isWorker.value && st === 'pending') acts.push({ label: '开始作业', primary: true, run: () => act(() => request({ url: '/api/v1/work-orders/' + w.value.id + '/start', method: 'POST' }), '已开始作业') })
  if (isWorker.value && st === 'ongoing') acts.push({
    label: '完成作业（可上传成果）', primary: true,
    run: async () => {
      const ok = await chooseAndUpload()
      if (!ok && chosenPhotos.length === 0) return
      await act(() => request({ url: '/api/v1/work-orders/' + w.value.id + '/complete', method: 'POST', data: { result_photos: chosenPhotos } }), '作业已提交，等待验收')
    }
  })
  // 需求方：待验收 → 验收通过 / 要求整改
  if (isPub.value && st === 'awaiting_accept') {
    acts.push({ label: '验收通过', primary: true, run: () => act(() => request({ url: '/api/v1/work-orders/' + w.value.id + '/accept', method: 'POST' }), '验收完成') })
    acts.push({
      label: '要求整改', danger: true,
      run: () => uni.showModal({
        title: '要求整改',
        editable: true,
        placeholderText: '填写整改要求',
        success: (r) => {
          if (!r.confirm || !(r.content || '').trim()) return
          act(() => request({ url: '/api/v1/work-orders/' + w.value.id + '/rework', method: 'POST', data: { note: r.content.trim() } }), '已发出整改要求')
        }
      })
    })
  }
  // 双方：未完结可取消（填写原因）
  if (st !== 'completed' && st !== 'cancelled') {
    acts.push({
      label: '取消订单', danger: true,
      run: () => uni.showModal({
        title: '取消订单',
        editable: true,
        placeholderText: '请填写取消原因',
        success: (r) => {
          if (!r.confirm || !(r.content || '').trim()) return
          act(() => request({ url: '/api/v1/work-orders/' + w.value.id + '/cancel', method: 'POST', data: { reason: r.content.trim() } }), '已取消订单')
        }
      })
    })
  }
  return acts
})

async function load() {
  loading.value = true
  try {
    const res = await request({ url: '/api/v1/work-orders/' + encodeURIComponent(id.value) })
    w.value = res || null
  } catch (e) {
    w.value = null
  } finally {
    loading.value = false
  }
}

onLoad((options) => {
  id.value = (options && options.id) || ''
  load()
})
</script>

<style>
page { background: var(--color-bg); }
.wod-page { min-height: 100vh; padding: 20rpx; box-sizing: border-box; }
.wod-state { text-align: center; padding: 120rpx 0; color: #98A2B3; font-size: 26rpx; }
.wod-hero { border-radius: 10px; padding: 32rpx 28rpx; color: #fff; background: linear-gradient(160deg, #0a3a6b, #074d92); }
.wod-hero--completed { background: linear-gradient(160deg, #0B6B41, #168A55); }
.wod-hero--cancelled { background: linear-gradient(160deg, #6B3A0A, #D92D20); }
.wod-status-label { display: block; font-size: 24rpx; opacity: .9; }
.wod-no { display: block; font-size: 26rpx; margin-top: 8rpx; opacity: .85; }
.wod-amount { display: block; font-size: 44rpx; font-weight: 800; margin-top: 16rpx; }
.wod-card { background: #fff; border: 1rpx solid #E4E7EC; border-radius: 10px; padding: 24rpx; margin-top: 20rpx; box-shadow: 0 4px 20px rgba(16,24,40,.06); }
.wod-card-title { display: block; font-size: 26rpx; font-weight: 600; color: #17212B; margin-bottom: 12rpx; }
.wod-row { display: flex; justify-content: space-between; gap: 20rpx; padding: 10rpx 0; }
.wod-k { flex-shrink: 0; font-size: 24rpx; color: #667085; }
.wod-v { font-size: 24rpx; color: #17212B; text-align: right; word-break: break-all; }
.wod-v--warn { color: #B54708; }
.wod-v--danger { color: #D92D20; }
.wod-photos { display: flex; flex-wrap: wrap; gap: 12rpx; }
.wod-photo { width: 200rpx; height: 150rpx; border-radius: 8rpx; background: #F4F6F8; }
.wod-btn { height: 88rpx; border-radius: 44rpx; display: flex; align-items: center; justify-content: center; font-size: 28rpx; font-weight: 600; background: #EEF1F4; color: #344054; margin-bottom: 16rpx; }
.wod-btn.primary { background: #0A66C2; color: #fff; }
.wod-btn.danger { background: #FEF3F2; color: #D92D20; }
.wod-btn-hover { opacity: .85; }
.wod-done { text-align: center; font-size: 24rpx; color: #98A2B3; margin-top: 24rpx; }
</style>
