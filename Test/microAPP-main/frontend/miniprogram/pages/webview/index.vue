<template>
  <web-view v-if="decodedUrl" :src="decodedUrl" />
  <view v-else class="empty-state">暂无链接</view>

  <!-- web-view 页面不再提供“首页”图标（按需求仅保留详情/填表页） -->
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const decodedUrl = ref('')

onLoad((options) => {
  if (options?.src) {
    decodedUrl.value = decodeURIComponent(options.src)
  }
})

</script>

<style scoped>
.empty-state {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #969799;
}

</style>

