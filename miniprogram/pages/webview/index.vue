<template>
  <web-view v-if="valid" :src="url" />
  <view v-else class="err-page">
    <view class="err-mark">!</view>
    <text class="err-title">链接不可用</text>
    <text class="err-desc">该链接不在平台允许访问的域名白名单内</text>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

// 默认白名单：仅平台业务域名。调用方可通过 allowed_domains 显式追加业务白名单（逗号分隔）。
const DEFAULT_ALLOWED_HOSTS = ['api.cqnarc.cn']

const url = ref('')
const valid = ref(false)

// 提取 https:// 后的主机名（去端口、统一小写），非 https 或不合法返回空串
function parseHost(u) {
  if (typeof u !== 'string') return ''
  const m = /^https:\/\/([^\/?#]+)/.exec(u)
  if (!m) return ''
  return m[1].replace(/:\d+$/, '').toLowerCase()
}

onLoad((opts) => {
  let raw = ''
  try {
    raw = opts && opts.url ? decodeURIComponent(opts.url) : ''
  } catch (e) {
    raw = ''
  }
  const extra = String((opts && opts.allowed_domains) || '')
    .split(',')
    .map((s) => s.trim().toLowerCase())
    .filter(Boolean)
  const allowedHosts = [...DEFAULT_ALLOWED_HOSTS, ...extra]
  const host = parseHost(raw)
  if (raw.startsWith('https://') && host && allowedHosts.includes(host)) {
    url.value = raw
    valid.value = true
  } else {
    valid.value = false
  }
})
</script>

<style scoped>
.err-page {
  min-height: 100vh;
  background: #F4F6F8;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 48rpx;
}
.err-mark {
  width: 124rpx;
  height: 124rpx;
  border-radius: 50%;
  background: #FEF3F2;
  color: #D92D20;
  font-size: 54rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12rpx;
}
.err-title { font-size: 30rpx; font-weight: 700; color: #17212B; }
.err-desc { font-size: 24rpx; color: #98A2B3; text-align: center; line-height: 1.6; }
</style>
