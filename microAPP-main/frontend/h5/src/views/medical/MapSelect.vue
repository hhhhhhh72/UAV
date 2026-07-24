<template>
  <div class="map-select-page">
    <van-nav-bar title="选择起降场" left-arrow @click-left="$router.back()" fixed placeholder />

    <div class="map-container" id="mapContainer"></div>

    <div class="pad-list-bottom">
      <h3>可用起降场列表</h3>
      <van-cell v-for="pad in pads" :key="pad.id" :title="pad.name" :label="`${pad.address} | 运营时间: ${pad.operating_hours}`" is-link @click="selectPad(pad)">
        <template #right-icon>
          <van-tag type="success" v-if="pad.enabled">启用</van-tag>
        </template>
      </van-cell>
      <van-empty v-if="!pads.length" description="暂无可用起降场" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMedicalStore } from '@/stores/medical'

const router = useRouter()
const route = useRoute()
const store = useMedicalStore()
const pads = computed(() => store.pads)

onMounted(async () => {
  await store.fetchPads()
})

function selectPad(pad) {
  // 通过 sessionStorage 传递选择结果
  const target = route.query.target || 'departure'
  sessionStorage.setItem(`selectedPad_${target}`, JSON.stringify(pad))
  router.back()
}
</script>

<style scoped>
.map-select-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.map-select-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.map-container {
  height: 40vh;
  background: #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
}
.map-container::after {
  content: '地图组件（Leaflet集成预留区域）';
}
.pad-list-bottom {
  padding: 16px;
}
.pad-list-bottom h3 {
  font-size: 16px;
  margin-bottom: 12px;
}
</style>
