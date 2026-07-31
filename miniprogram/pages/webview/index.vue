<template>
  <web-view :src="url" />
</template>
<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
const url = ref('')
onLoad((opts) => {
  if (opts.url) url.value = decodeURIComponent(opts.url)
})
</script>
