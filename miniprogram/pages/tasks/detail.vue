<template>
<view class="page">
  <!-- 顶部帖子（真实数据） -->
  <view v-if="post" class="post-card">
    <view class="post-head">
      <view class="post-title">{{ post.title || '未命名需求' }}</view>
      <view class="post-poster">生成海报</view>
    </view>
    <view class="post-meta">
      <text>{{ (post.created_at || '').slice(0, 16).replace('T', ' ') }} · 需求大厅</text>
    </view>
    <view class="post-stats">
      <text>{{ views }}浏览、0人点赞</text>
    </view>
    <view class="post-fields">
      <text class="field-row"><text class="field-key">业务类型</text>：{{ bizTypeLabel(post.biz_type) }}</text>
      <text class="field-row"><text class="field-key">预算</text>：¥{{ fmtFen(post.budget_fen) }}</text>
      <text class="field-row"><text class="field-key">地区</text>：{{ post.district || '未填写' }}</text>
    </view>
    <text class="post-body">{{ post.description || '暂无详细描述，请联系发布者' }}</text>
    <view class="post-imgs">
      <image v-for="(img, i) in postImgs" :key="i" :src="img" mode="aspectFill" class="post-img-big" />
    </view>
  </view>
  <view v-else class="post-card" style="padding: 40rpx 0; text-align: center">
    <u-empty description="需求不存在或已下线" />
  </view>

  <!-- 位置 + 点赞 -->
  <view class="loc-bar">
    <view class="loc-text"><u-icon name="location" size="28rpx" color="var(--color-primary)" /><text>{{ post?.district || '位置未填写' }}</text></view>
    <view class="like-btn" @tap="toggleLike">赞 {{ likes }}</view>
  </view>

  <!-- 安全提示 -->
  <view class="safe-tip">
    <text class="safe-title">如遇无效、虚假、诈骗信息，请立即举报</text>
    <text class="safe-sub">为了您的资金安全，请见面交易，切勿提前支付任何费用</text>
  </view>

  <!-- 评论区 -->
  <view class="comment-bar">
    <text class="comment-title">全部评论</text>
    <text class="comment-btn">评论</text>
  </view>
  <view class="comment-empty">还没有评论...</view>

  <!-- 相关帖子 -->
  <view class="related-item">
    <view class="related-logo">小飞虾</view>
    <view class="related-body">
      <text class="related-title">小飞虾独家项目目...</text>
      <text class="related-sub">发布454条</text>
    </view>
    <text class="related-arrow">›</text>
  </view>

  <!-- 分类导航 -->
  <view class="cat-bar">
    <text class="cat-tag" v-for="c in cats" :key="c.id" :class="{on: c.id===activeCat}" @tap="activeCat=c.id">{{ c.name }}</text>
  </view>

  <!-- 悬浮按钮 -->
  <view class="float-bar">
    <view class="float-kefu">客服</view>
    <view class="float-row">
      <view class="float-btn1" @tap="callPhone">电话</view>
      <view class="float-btn2" @tap="goHelp">互帮互助</view>
    </view>
  </view>
</view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, BASE_URL } from '@/utils/request'

const views = ref(0)
const likes = ref(0)
const post = ref(null)
const activeCat = ref('news')
const cats = ref([
  { id:'news', name:'最新信息' }, { id:'lift', name:'吊运独家' }, { id:'trade', name:'买卖租赁' },
  { id:'train', name:'考证培训' }, { id:'plant', name:'植保运输' },
])
const postImgs = computed(() => {
  try {
    const arr = typeof post.value?.images === 'string' ? JSON.parse(post.value.images) : (post.value?.images || [])
    return arr.map(u => (u && u.startsWith('http') ? u : BASE_URL + u)).slice(0, 2)
  } catch (e) { return [] }
})
const bizTypeLabel = (t) => ({ lift: '吊运', aerial: '航拍', plant: '植保', patrol: '巡检', survey: '测绘', logistics: '物流', purchase: '采购' }[t] || t || '其他')
const fmtFen = (f) => f ? (f / 100).toLocaleString('en-US') : '面议'

onLoad((opts) => {
  if (!opts.id) return;
  (async () => {
    try {
      const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(opts.id) })
      const d = (res && res.data) || res
      if (d && d.id) {
        post.value = d
        views.value = d.views || 0
      }
    } catch {}
  })()
})

const toggleLike = () => likes.value++
const callPhone = () => uni.showToast({ title:'电话功能开发中', icon:'none' })
const goHelp = () => uni.showToast({ title:'互帮互助开发中', icon:'none' })
</script>

<style scoped>
.page { min-height:100vh; background:var(--color-bg); padding-bottom: 100px; }

/* 帖子卡片 */
.post-card { background:#fff; padding:14px 16px 10px; }
.post-head { display:flex; justify-content:space-between; align-items:flex-start; }
.post-title { font-size:16px; font-weight:700; color:var(--color-text); flex:1; line-height:1.4; padding-right:8px; }
.post-poster { font-size:12px; color:var(--color-primary); padding:4px 8px; border:1px solid var(--color-primary); border-radius:4px; flex-shrink:0; }
.post-meta { font-size:12px; color:#999; margin:6px 0 4px; }
.post-stats { font-size:12px; color:#999; margin-bottom:8px; }
.post-fields { background:#fff8e1; border-radius:6px; padding:8px 10px; margin:8px 0 10px; }
.field-row { font-size:13px; color:#333; display:block; line-height:1.7; }
.field-key { color:var(--color-warning); font-weight:600; }
.post-body { font-size:14px; color:#333; line-height:1.7; display:block; margin:10px 0 12px; }
.post-imgs { display:flex; flex-direction:column; gap:8px; }
.post-img-big { width:100%; height:200px; border-radius:6px; background:#f5f5f5; }

/* 位置 */
.loc-bar { background:#fff; padding:10px 16px; margin-top:8px; display:flex; justify-content:space-between; align-items:center; font-size:13px; color:var(--color-primary); }
.loc-text { display:flex; align-items:center; gap:6rpx; }
.like-btn { font-size:13px; color:#666; padding:2px 8px; background:#f5f5f5; border-radius:4px; }

/* 安全提示 */
.safe-tip { background:#fff8e1; padding:10px 16px; border-top:1px solid #ffe082; border-bottom:1px solid #ffe082; }
.safe-title { font-size:13px; color:var(--color-warning); font-weight:600; display:block; }
.safe-sub { font-size:11px; color:#999; display:block; margin-top:3px; }

/* 评论区 */
.comment-bar { background:#fff; padding:12px 16px; display:flex; justify-content:space-between; align-items:center; margin-top:8px; }
.comment-title { font-size:15px; font-weight:600; color:var(--color-text); }
.comment-btn { font-size:13px; color:var(--color-primary); padding:3px 10px; border:1px solid var(--color-primary); border-radius:12px; }
.comment-empty { background:#fff; padding:30px 0; text-align:center; color:#999; font-size:13px; }

/* 相关帖子 */
.related-item { display:flex; align-items:center; gap:10px; background:#fff; padding:12px 16px; margin-top:8px; }
.related-logo { width:32px; height:32px; border-radius:6px; background:var(--color-primary); color:#fff; display:flex; align-items:center; justify-content:center; font-size:9px; flex-shrink:0; }
.related-body { flex:1; }
.related-title { font-size:13px; font-weight:500; color:var(--color-text); display:block; }
.related-sub { font-size:11px; color:#999; margin-top:2px; }
.related-arrow { font-size:18px; color:#ccc; }

/* 分类导航 */
.cat-bar { display:flex; background:#fff; padding:10px 16px; border-top:1px solid #f0f0f0; margin-top:8px; gap:14px; white-space:nowrap; overflow-x:auto; }
.cat-tag { font-size:13px; color:#666; padding:2px 0; }
.cat-tag.on { color:var(--color-primary); font-weight:600; border-bottom:2px solid var(--color-primary); }

/* 悬浮 */
.float-bar { position:fixed; bottom:0; left:0; right:0; display:flex; align-items:center; gap:8px; padding:8px 12px; background:rgba(255,255,255,0.95); backdrop-filter:blur(8px); }
.float-kefu { position:fixed; right:8px; bottom:80px; width:46px; height:46px; border-radius:50%; background:var(--color-primary); color:#fff; display:flex; align-items:center; justify-content:center; font-size:11px; box-shadow:0 2px 8px rgba(10,102,194,.25); }
.float-row { display:flex; gap:8px; flex:1; }
.float-btn1 { flex:1; height:40px; border-radius:20px; background:#4fc3f7); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:500; }
.float-btn2 { flex:1; height:40px; border-radius:20px; background:#4fc3f7); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:500; }
</style>
