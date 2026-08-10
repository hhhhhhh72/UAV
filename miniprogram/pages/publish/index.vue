<template>
  <Layout :current="2">
    <view class="publish-page">
      <view class="page-head">
        <text class="page-title">发布</text>
        <text class="page-sub">发布需求与供给，快速对接业务机会</text>
      </view>

      <!-- 发布类型功能砖 -->
      <view class="publish-grid">
        <view class="pub-item" hover-class="tap-fade" @tap="publish('demand')">
          <view class="pub-icon" :style="{ background: '#EAF3FB' }">
            <image class="pub-icon-img" :src="'/static/home/icons/demand.svg'" mode="aspectFit" />
          </view>
          <view class="pub-copy">
            <text class="pub-name">发布需求</text>
            <text class="pub-desc">巡检/植保/测绘等项目</text>
          </view>
          <text class="pub-arrow">›</text>
        </view>

        <view class="pub-item" hover-class="tap-fade" @tap="publish('service')">
          <view class="pub-icon" :style="{ background: '#FFF0E6' }">
            <image class="pub-icon-img" :src="'/static/icons/service.svg'" mode="aspectFit" />
          </view>
          <view class="pub-copy">
            <text class="pub-name">发布服务能力</text>
            <text class="pub-desc">展示巡检、测绘、航拍等可承接能力</text>
          </view>
          <text class="pub-arrow">›</text>
        </view>

        <view class="pub-item" hover-class="tap-fade" @tap="publish('product')">
          <view class="pub-icon" :style="{ background: '#E9F7F0' }">
            <image class="pub-icon-img" :src="'/static/home/icons/shop.svg'" mode="aspectFit" />
          </view>
          <view class="pub-copy">
            <text class="pub-name">发布商品设备</text>
            <text class="pub-desc">设备租赁、整机、零部件或载荷</text>
          </view>
          <text class="pub-arrow">›</text>
        </view>

        <view class="pub-item" hover-class="tap-fade" @tap="publish('course')">
          <view class="pub-icon" :style="{ background: '#F0EDFF' }">
            <image class="pub-icon-img" :src="'/static/home/icons/training.svg'" mode="aspectFit" />
          </view>
          <view class="pub-copy">
            <text class="pub-name">发布培训课程</text>
            <text class="pub-desc">CAAC/UTC/职业技能等课程招生</text>
          </view>
          <text class="pub-arrow">›</text>
        </view>
      </view>

      <!-- 我的发布管理 -->
      <view class="manage-section">
        <text class="manage-title">我的发布</text>
        <view class="manage-item" hover-class="tap-fade" @tap="go('/pkg-demand/pages/demands/mine')">
          <text class="manage-name">我的需求</text>
          <text class="manage-desc">审核状态 · 对接意向 · 成交登记</text>
          <text class="pub-arrow">›</text>
        </view>
        <view class="manage-item" hover-class="tap-fade" @tap="go('/pages/tasks/index')">
          <text class="manage-name">我的任务</text>
          <text class="manage-desc">接单与执行记录</text>
          <text class="pub-arrow">›</text>
        </view>
        <view class="manage-item" hover-class="tap-fade" @tap="go('/pages/intents/mine')">
          <text class="manage-name">我的对接意向</text>
          <text class="manage-desc">我登记过的对接记录</text>
          <text class="pub-arrow">›</text>
        </view>
      </view>
    </view>
  </Layout>
</template>

<script setup>
import Layout from '../../components/Layout.vue'

const publish = (type) => {
  if (type === 'demand') return uni.navigateTo({ url: '/pkg-demand/pages/demands/publish' })
  if (type === 'service') return uni.navigateTo({ url: '/pkg-service/pages/publish/service' })
  if (type === 'product') return uni.navigateTo({ url: '/pkg-service/pages/publish/product' })
  if (type === 'course') return uni.navigateTo({ url: '/pkg-service/pages/publish/course' })
  uni.showToast({ title: '暂未开放', icon: 'none', duration: 2000 })
}
const go = (url) => uni.navigateTo({ url })
</script>

<style scoped>
.publish-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding: 16px 12px calc(24px + env(safe-area-inset-bottom));
}

.page-head {
  padding: 8px 4px 16px;
}

.page-title {
  display: block;
  font-size: 21px;
  font-weight: 700;
  color: #17212B;
}

.page-sub {
  display: block;
  font-size: 12px;
  color: #667085;
  margin-top: 4px;
}

/* 发布类型砖 */
.publish-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pub-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 14px;
}

.pub-item.pub-muted {
  opacity: 0.85;
}

.pub-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pub-icon-img {
  width: 26px;
  height: 26px;
}

.pub-copy {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pub-name {
  font-size: 15px;
  font-weight: 600;
  color: #17212B;
}

.pub-desc {
  font-size: 11px;
  color: #98A2B3;
}

.pub-arrow {
  font-size: 18px;
  color: #98A2B3;
}

/* 我的发布管理 */
.manage-section {
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #EEF1F4;
  padding: 4px 14px;
}

.manage-title {
  display: block;
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  padding: 14px 0 10px;
}

.manage-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 0;
  border-top: 1px solid #EEF1F4;
}

.manage-name {
  font-size: 14px;
  font-weight: 500;
  color: #17212B;
}

.manage-desc {
  flex: 1;
  font-size: 11px;
  color: #98A2B3;
  text-align: right;
}

.tap-fade {
  opacity: 0.7;
}
</style>
