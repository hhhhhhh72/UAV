<template>
<view class="page">
  <!-- 顶部帖子 -->
  <view class="post-card">
    <view class="post-head">
      <view class="post-title">独立吊运业务  小飞虹独家项目目...</view>
      <view class="post-poster">生成海报</view>
    </view>
    <view class="post-meta">
      <text>3天前 来自小飞虹无人机圈子社区平台</text>
    </view>
    <view class="post-stats">
      <text>{{ views }}浏览、0人点赞</text>
    </view>
    <view class="post-fields">
      <text class="field-row"><text class="field-key">货物类型</text>：树/木头</text>
      <text class="field-row"><text class="field-key">项目总量</text>：300吨</text>
      <text class="field-row"><text class="field-key">启动时间</text>：2026-07-29 10:47</text>
      <text class="field-row"><text class="field-key">地区</text>：重庆市万州区</text>
    </view>
    <text class="post-body">重庆300吨柏木头，单件重量100-400内，吊运距离200-900米，100米高度，项目只给代理运营商和签约飞手及从平台采购无人机者独家承接，欢迎加入。</text>
    <view class="post-imgs">
      <image v-for="i in 2" :key="i" src="/static/home-bg.jpg" mode="aspectFill" class="post-img-big" />
    </view>
  </view>

  <!-- 位置 + 点赞 -->
  <view class="loc-bar">
    <text>📍 两江新区金山街道加工区八路</text>
    <view class="like-btn" @tap="toggleLike">👍 {{ likes }}</view>
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
    <view class="float-kefu">💬 客服</view>
    <view class="float-row">
      <view class="float-btn1" @tap="callPhone">📞 电话</view>
      <view class="float-btn2" @tap="goHelp">💬 互帮互助</view>
    </view>
  </view>
</view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'

const views = ref(8586)
const likes = ref(0)
const activeCat = ref('news')
const cats = ref([
  { id:'news', name:'最新信息' }, { id:'lift', name:'吊运独家' }, { id:'trade', name:'买卖租赁' },
  { id:'train', name:'考证培训' }, { id:'plant', name:'植保运输' },
])
onLoad((opts) => {
  if (!opts.id) return;
  (async () => {
    try {
      const res = await request({ url:'/api/v1/demands', data:{page:1,page_size:1} })
      const list = (Array.isArray(res) ? res : (res.data||[]))
      if (list[0]) {
        views.value = list[0].views || 8586
      }
    } catch {}
  })()
})

const toggleLike = () => likes.value++
const callPhone = () => uni.showToast({ title:'电话功能开发中', icon:'none' })
const goHelp = () => uni.showToast({ title:'互帮互助开发中', icon:'none' })
</script>

<style scoped>
.page { min-height:100vh; background:#f5f6f7; padding-bottom: 100px; }

/* 帖子卡片 */
.post-card { background:#fff; padding:14px 16px 10px; }
.post-head { display:flex; justify-content:space-between; align-items:flex-start; }
.post-title { font-size:16px; font-weight:700; color:#1a1a1a; flex:1; line-height:1.4; padding-right:8px; }
.post-poster { font-size:12px; color:#0A66C2; padding:4px 8px; border:1px solid #0A66C2; border-radius:4px; flex-shrink:0; }
.post-meta { font-size:12px; color:#999; margin:6px 0 4px; }
.post-stats { font-size:12px; color:#999; margin-bottom:8px; }
.post-fields { background:#fff8e1; border-radius:6px; padding:8px 10px; margin:8px 0 10px; }
.field-row { font-size:13px; color:#333; display:block; line-height:1.7; }
.field-key { color:#f57c00; font-weight:600; }
.post-body { font-size:14px; color:#333; line-height:1.7; display:block; margin:10px 0 12px; }
.post-imgs { display:flex; flex-direction:column; gap:8px; }
.post-img-big { width:100%; height:200px; border-radius:6px; background:#f5f5f5; }

/* 位置 */
.loc-bar { background:#fff; padding:10px 16px; margin-top:8px; display:flex; justify-content:space-between; align-items:center; font-size:13px; color:#0A66C2; }
.like-btn { font-size:13px; color:#666; padding:2px 8px; background:#f5f5f5; border-radius:4px; }

/* 安全提示 */
.safe-tip { background:#fff8e1; padding:10px 16px; border-top:1px solid #ffe082; border-bottom:1px solid #ffe082; }
.safe-title { font-size:13px; color:#f57c00; font-weight:600; display:block; }
.safe-sub { font-size:11px; color:#999; display:block; margin-top:3px; }

/* 评论区 */
.comment-bar { background:#fff; padding:12px 16px; display:flex; justify-content:space-between; align-items:center; margin-top:8px; }
.comment-title { font-size:15px; font-weight:600; color:#1a1a1a; }
.comment-btn { font-size:13px; color:#0A66C2; padding:3px 10px; border:1px solid #0A66C2; border-radius:12px; }
.comment-empty { background:#fff; padding:30px 0; text-align:center; color:#999; font-size:13px; }

/* 相关帖子 */
.related-item { display:flex; align-items:center; gap:10px; background:#fff; padding:12px 16px; margin-top:8px; }
.related-logo { width:32px; height:32px; border-radius:6px; background:linear-gradient(135deg,#0A66C2,#0ea5e9); color:#fff; display:flex; align-items:center; justify-content:center; font-size:9px; flex-shrink:0; }
.related-body { flex:1; }
.related-title { font-size:13px; font-weight:500; color:#1a1a1a; display:block; }
.related-sub { font-size:11px; color:#999; margin-top:2px; }
.related-arrow { font-size:18px; color:#ccc; }

/* 分类导航 */
.cat-bar { display:flex; background:#fff; padding:10px 16px; border-top:1px solid #f0f0f0; margin-top:8px; gap:14px; white-space:nowrap; overflow-x:auto; }
.cat-tag { font-size:13px; color:#666; padding:2px 0; }
.cat-tag.on { color:#0A66C2; font-weight:600; border-bottom:2px solid #0A66C2; }

/* 悬浮 */
.float-bar { position:fixed; bottom:0; left:0; right:0; display:flex; align-items:center; gap:8px; padding:8px 12px; background:linear-gradient(to top, rgba(255,255,255,.95), rgba(255,255,255,.7)); backdrop-filter:blur(8px); }
.float-kefu { position:fixed; right:8px; bottom:80px; width:46px; height:46px; border-radius:50%; background:#0A66C2; color:#fff; display:flex; align-items:center; justify-content:center; font-size:11px; box-shadow:0 2px 8px rgba(10,102,194,.25); }
.float-row { display:flex; gap:8px; flex:1; }
.float-btn1 { flex:1; height:40px; border-radius:20px; background:linear-gradient(135deg,#4fc3f7,#0A66C2); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:500; }
.float-btn2 { flex:1; height:40px; border-radius:20px; background:linear-gradient(135deg,#4fc3f7,#0A66C2); color:#fff; display:flex; align-items:center; justify-content:center; font-size:14px; font-weight:500; }
</style>
