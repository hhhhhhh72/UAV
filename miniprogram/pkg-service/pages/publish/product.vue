<template>
  <view class="redirect-page"></view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'

// 旧的单页商品发布已并入统一发布表单（pages/publish/form.vue，按原型重建，两步流程）
onLoad(() => {
  uni.redirectTo({ url: '/pages/publish/form?type=product' })
})
</script>

<style scoped>
.redirect-page { min-height: 100vh; background: #F5F6F8; }
</style>
