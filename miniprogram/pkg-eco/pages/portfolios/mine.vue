<template>
  <view class="page" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="我的品牌" show-back :fixed="true" @back="goBack" />

    <!-- 加载骨架 -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w70"></view></view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- 空态：尚未创建 -->
    <view v-else-if="!d" class="m-empty">
      <view class="m-empty-ic"><text>牌</text></view>
      <text class="m-empty-t">尚未创建企业品牌</text>
      <text class="m-empty-d">创建品牌名片，向行业展示企业产品、荣誉与案例</text>
      <view class="m-empty-btn" hover-class="tap-fade" @tap="goEdit('')">立即创建品牌</view>
      <text class="m-note">创建后需经协会审核，审核通过后对外展示</text>
    </view>

    <!-- 已创建 -->
    <template v-else>
      <view class="m-hero" :class="d.grad">
        <view class="m-hero-inner">
          <view class="m-logo"><text>{{ d.logoText }}</text></view>
          <view class="m-txt">
            <text class="m-tag" :class="'st--' + d.status">{{ d.statusLabel }}</text>
            <text class="m-name">{{ d.name }}</text>
            <text class="m-sub">{{ d.desc }}</text>
          </view>
        </view>
      </view>

      <!-- 数据统计 -->
      <view class="m-stat">
        <view class="m-si"><text class="m-sv">{{ d.products }}</text><text class="m-sl">展示产品</text></view>
        <view class="m-si"><text class="m-sv">{{ d.honors }}</text><text class="m-sl">荣誉资质</text></view>
        <view class="m-si"><text class="m-sv">{{ d.videoCount }}</text><text class="m-sl">宣传视频</text></view>
        <view class="m-si"><text class="m-sv">{{ d.caseCount }}</text><text class="m-sl">品牌案例</text></view>
      </view>

      <!-- 管理入口 -->
      <view class="m-sec">
        <view class="m-row" hover-class="tap-fade" @tap="goEdit(d.id)">
          <view class="m-ic"><text>编</text></view>
          <text class="m-t">编辑品牌资料</text>
          <text class="m-ar">›</text>
        </view>
        <view class="m-row" hover-class="tap-fade" @tap="goDetail(d.id)">
          <view class="m-ic ic-blue"><text>览</text></view>
          <text class="m-t">预览品牌主页</text>
          <text class="m-ar">›</text>
        </view>
        <view class="m-row" hover-class="tap-fade" @tap="goEdit(d.id)">
          <view class="m-ic ic-green"><text>联</text></view>
          <text class="m-t">联系方式管理</text>
          <text class="m-ar">›</text>
        </view>
      </view>

      <view class="m-note note-bottom">品牌内容由企业自主维护，提交后经协会审核即可对外展示；审核中仍可修改。</view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { request, authStorage } from '@/utils/request'

// ===== 状态 =====
const statusBarHeight = ref(20)
const loading = ref(true)
const err = ref(false)
const d = ref(null)

// 状态文案（对齐后端 status：draft / pending / published / rejected）
const STATUS_LABEL = { published: '已发布', pending: '待审核', review: '待审核', draft: '草稿', rejected: '未通过' }
const statusKey = (s) => (STATUS_LABEL[s] ? s : 'draft')
const gradCls = (g) => 'gd-' + String(g || 'gd1').replace(/^gd-?/, '')

const mapMine = (it) => {
  const key = statusKey(it.status)
  return {
    id: it.id,
    name: it.name || it.company_name || '',
    logoText: it.logo_text || it.char || (it.name ? String(it.name).charAt(0) : '牌'),
    status: key,
    statusLabel: STATUS_LABEL[key] || '草稿',
    desc: it.desc || it.description || '',
    products: Array.isArray(it.products) ? it.products.length : 0,
    honors: Array.isArray(it.honors) ? it.honors.length : 0,
    videoCount: it.video_count ?? 0,
    caseCount: it.case_count ?? 0,
    grad: gradCls(it.grad),
    raw: it,
  }
}

// ===== 数据获取 =====
// 接口：GET /api/v1/portfolios/mine（返回当前企业品牌列表，取第一条）
const fetchData = async () => {
  loading.value = true
  err.value = false
  try {
    const res = await request({ url: '/api/v1/portfolios/mine' })
    const list = Array.isArray(res) ? res : (res && res.data) || []
    d.value = list.length ? mapMine(list[0]) : null
  } catch {
    err.value = true
  } finally {
    loading.value = false
  }
}

// ===== 交互 =====
const goEdit = (id) => {
  if (d.value && d.value.raw) uni.setStorageSync('portfolio_edit_cache', d.value.raw)
  uni.navigateTo({ url: '/pkg-eco/pages/portfolios/edit' + (id ? '?id=' + encodeURIComponent(id) : '') })
}
const goDetail = (id) => {
  if (d.value && d.value.raw) uni.setStorageSync('portfolio_cache_' + id, d.value.raw)
  uni.navigateTo({ url: '/pkg-eco/pages/portfolios/detail?id=' + encodeURIComponent(id) })
}
const goBack = () => uni.navigateBack()

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  const token = authStorage.getAccessToken()
  if (!token) {
    loading.value = false
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(() => uni.navigateTo({ url: '/pages/login/index' }), 500)
    return
  }
  fetchData()
})
onShow(() => {
  // 从编辑页返回时刷新（未登录不重复请求）
  if (authStorage.getAccessToken() && (d.value || !loading.value)) fetchData()
})
onPullDownRefresh(async () => {
  await fetchData()
  uni.stopPullDownRefresh()
})
</script>

<style scoped>
.page { min-height: 100vh; background: #F4F6F8; padding-bottom: 60rpx; }

/* ===== 骨架 ===== */
.skw { padding-top: 20rpx; }
.sk-h { height: 260rpx; background: #f0f1f3; animation: shimmer 1.5s infinite; border-radius: 24rpx; margin: 24rpx; }
.sk-sec { margin: 24rpx; padding: 32rpx; background: #fff; border-radius: 24rpx; }
.sk-l { height: 28rpx; background: #f0f1f3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s infinite; }
.sk-l.w80 { width: 80%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
.sk-l.w100 { width: 100%; }
.sk-l.w70 { width: 70%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 错误 ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 40rpx; min-height: 600rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 16rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 24rpx; }

/* ===== 空态 ===== */
.m-empty { display: flex; flex-direction: column; align-items: center; padding: 180rpx 48rpx 40rpx; text-align: center; }
.m-empty-ic {
  width: 152rpx; height: 152rpx; border-radius: 44rpx;
  background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5);
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 60rpx; font-weight: 800;
  box-shadow: 0 24rpx 56rpx rgba(21,101,192,.35); margin-bottom: 36rpx;
}
.m-empty-t { font-size: 32rpx; font-weight: 700; color: #17212B; margin-bottom: 16rpx; }
.m-empty-d { font-size: 25rpx; color: #667085; line-height: 1.6; margin-bottom: 44rpx; }
.m-empty-btn {
  height: 88rpx; padding: 0 72rpx; border-radius: 16rpx; background: #0A66C2; color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 28rpx; font-weight: 600;
  box-shadow: 0 12rpx 36rpx rgba(10,102,194,.35);
}

/* ===== 品牌状态卡 ===== */
.m-hero { position: relative; height: 300rpx; margin: 24rpx; border-radius: 28rpx; overflow: hidden; }
.m-hero-inner {
  position: absolute; inset: 0;
  display: flex; align-items: center; gap: 24rpx; padding: 0 32rpx;
  background: linear-gradient(90deg, rgba(8,34,62,.6), rgba(8,34,62,0) 80%);
}
.m-logo {
  width: 104rpx; height: 104rpx; border-radius: 28rpx;
  background: rgba(255,255,255,.95);
  display: flex; align-items: center; justify-content: center;
  font-size: 44rpx; font-weight: 800; color: #17212B;
  flex: none; box-shadow: 0 8rpx 24rpx rgba(0,0,0,.25);
}
.m-txt { flex: 1; min-width: 0; }
.m-tag {
  display: inline-block; font-size: 19rpx; color: #fff;
  padding: 4rpx 16rpx; border-radius: 12rpx; margin-bottom: 12rpx;
}
.st--published { background: #168A55; }
.st--pending { background: #E96012; }
.st--draft { background: #667085; }
.st--rejected { background: #D92D20; }
.m-name { font-size: 32rpx; font-weight: 700; color: #fff; line-height: 1.3; display: block; }
.m-sub {
  font-size: 22rpx; color: rgba(255,255,255,.85); margin-top: 8rpx;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}

/* ===== 数据统计 ===== */
.m-stat { display: flex; margin: 0 24rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 24rpx; padding: 24rpx 0; }
.m-si { flex: 1; text-align: center; position: relative; }
.m-si + .m-si::before { content: ''; position: absolute; left: 0; top: 16rpx; bottom: 16rpx; width: 1rpx; background: #EBEDF0; }
.m-sv { font-size: 32rpx; font-weight: 700; color: #17212B; display: block; }
.m-sl { font-size: 20rpx; color: #98A2B3; margin-top: 4rpx; display: block; }

/* ===== 管理入口 ===== */
.m-sec { margin: 24rpx; background: #fff; border: 1px solid #EEF1F4; border-radius: 24rpx; padding: 0 28rpx; }
.m-row { display: flex; align-items: center; gap: 20rpx; padding: 26rpx 0; border-bottom: 1rpx solid #EBEDF0; }
.m-row:last-child { border-bottom: none; }
.m-ic {
  width: 56rpx; height: 56rpx; border-radius: 16rpx;
  background: #EAF3FB; color: #0A66C2;
  display: flex; align-items: center; justify-content: center;
  font-size: 24rpx; font-weight: 600; flex: none;
}
.m-ic.ic-blue { background: #F6F4FF; color: #7A5AF8; }
.m-ic.ic-green { background: #E9F7F0; color: #168A55; }
.m-t { flex: 1; font-size: 26rpx; color: #17212B; font-weight: 500; }
.m-ar { color: #98A2B3; font-size: 30rpx; }

/* ===== 说明 ===== */
.m-note { font-size: 22rpx; color: #98A2B3; line-height: 1.6; padding: 0 28rpx; display: block; }
.note-bottom { margin-top: 24rpx; }

/* 封面渐变（占位视觉，真实数据以 cover_url 替换） */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-3 { background: linear-gradient(135deg,#e65100,#ef6c00 60%,#fb8c00); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
.gd-6 { background: linear-gradient(135deg,#b71c1c,#c62828 60%,#e57373); }
.gd-7 { background: linear-gradient(135deg,#1a237e,#283593 60%,#5c6bc0); }
.gd-8 { background: linear-gradient(135deg,#004d40,#00695c 60%,#4db6ac); }
</style>
