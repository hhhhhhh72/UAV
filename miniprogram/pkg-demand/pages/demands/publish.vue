<template>
  <view class="redirect-page"></view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'

// 旧的单页需求发布已并入统一发布表单（pages/publish/form.vue，按原型重建）
// 保留本路由以兼容既有入口（首页「发布需求」、我的需求「编辑重提」）
onLoad(() => {
  uni.redirectTo({ url: '/pages/publish/form?type=demand' })
})
</script>

<style scoped>
.redirect-page { min-height: 100vh; background: #F5F6F8; }
</style>
