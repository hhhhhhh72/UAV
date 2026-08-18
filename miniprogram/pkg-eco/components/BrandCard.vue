<template>
  <view class="brand-card" hover-class="tap-fade" @tap="$emit('click', item)">
    <view class="b-cover">
      <image v-if="item.cover" :src="item.cover" mode="aspectFill" class="b-cover-img" />
      <view class="bg" :class="gradClass"></view>
      <view class="b-logo">
        <image v-if="item.logo" :src="item.logo" mode="aspectFill" class="b-logo-img" />
        <text v-else>{{ item.logoText || '牌' }}</text>
      </view>
      <text class="b-char">{{ item.char || '牌' }}</text>
      <view v-if="item.hasVideo" class="b-play"><text>▶</text></view>
    </view>
    <view class="b-body">
      <text class="b-name">{{ item.name }}</text>
      <view class="b-tags">
        <text class="b-tag tag-cat">{{ item.catLabel }}</text>
        <text v-if="item.verified" class="b-tag tag-ok">已认证</text>
        <text v-else class="b-tag tag-ing">认证中</text>
      </view>
      <view v-if="item.views > 0 || item.videoCount > 0" class="b-foot">
        <text>浏览 <text class="bf-num">{{ fmt(item.views) }}</text></text>
        <text>视频 {{ item.videoCount }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
// 品牌网格卡片（对齐原型 .brand-card）
// 数据由父组件映射为展示字段（后端字段 → 展示字段），本组件只负责渲染
defineProps({
  item: { type: Object, required: true },
  gradClass: { type: String, default: 'gd-1' },
})
defineEmits(['click'])

const fmt = (n) => (n >= 10000 ? (n / 10000).toFixed(1) + 'w' : n >= 1000 ? (n / 1000).toFixed(1) + 'k' : String(n))
</script>

<style scoped>
.tap-fade { opacity: .72; }
.brand-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 20rpx;
  overflow: hidden;
  transition: transform .22s cubic-bezier(.25,.8,.3,1), box-shadow .22s;
}
.brand-card:active { transform: scale(.96); }

/* 封面 4:3 */
.b-cover { position: relative; padding-top: 75%; overflow: hidden; }
.b-cover-img { position: absolute; inset: 0; width: 100%; height: 100%; }
.bg { position: absolute; inset: 0; }
.b-logo {
  position: absolute; top: 20rpx; left: 20rpx;
  width: 68rpx; height: 68rpx; border-radius: 20rpx;
  background: rgba(255,255,255,.95);
  display: flex; align-items: center; justify-content: center;
  font-size: 30rpx; font-weight: 700; color: #17212B;
  box-shadow: 0 2px 8px rgba(0,0,0,.18);
  overflow: hidden;
}
.b-logo-img { width: 100%; height: 100%; }
.b-char {
  position: absolute; left: 20rpx; bottom: 16rpx;
  font-size: 68rpx; font-weight: 700;
  color: rgba(255,255,255,.9); text-shadow: 0 3px 10px rgba(0,0,0,.25);
  line-height: 1;
}
.b-play {
  position: absolute; right: 16rpx; bottom: 16rpx;
  width: 52rpx; height: 52rpx; border-radius: 50%;
  background: rgba(255,255,255,.95);
  display: flex; align-items: center; justify-content: center;
  color: #0A66C2; font-size: 20rpx;
  box-shadow: 0 2px 8px rgba(0,0,0,.22);
}
.b-body { padding: 20rpx 20rpx 22rpx; }
.b-name {
  font-size: 28rpx; font-weight: 600; color: #17212B;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  margin-bottom: 12rpx; display: block;
}
.b-tags { display: flex; gap: 8rpx; flex-wrap: wrap; margin-bottom: 14rpx; }
.b-tag { font-size: 19rpx; padding: 4rpx 14rpx; border-radius: 8rpx; font-weight: 500; }
.tag-cat { color: #0A66C2; background: #EAF3FB; }
.tag-ok { color: #168A55; background: #E9F7F0; }
.tag-ing { color: #E96012; background: #FFF0E6; }
.b-foot { font-size: 20rpx; color: #98A2B3; display: flex; justify-content: space-between; }
.bf-num { color: #667085; font-weight: 600; }

/* 封面渐变（占位视觉，真实数据以 images 封面图替换） */
.gd-1 { background: linear-gradient(135deg,#0d47a1,#1565c0 60%,#42a5f5); }
.gd-2 { background: linear-gradient(135deg,#004d40,#00695c 60%,#26a69a); }
.gd-3 { background: linear-gradient(135deg,#e65100,#ef6c00 60%,#fb8c00); }
.gd-4 { background: linear-gradient(135deg,#4a148c,#6a1b9a 60%,#ab47bc); }
.gd-5 { background: linear-gradient(135deg,#263238,#37474f 60%,#607d8b); }
.gd-6 { background: linear-gradient(135deg,#b71c1c,#c62828 60%,#e57373); }
.gd-7 { background: linear-gradient(135deg,#1a237e,#283593 60%,#5c6bc0); }
.gd-8 { background: linear-gradient(135deg,#004d40,#00695c 60%,#4db6ac); }
</style>
