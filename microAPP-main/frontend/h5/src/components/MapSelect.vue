<template>
  <div class="map-select">
    <div id="map" class="map-container"></div>
    <div class="map-footer">
      <div class="current-address" v-if="currentAddress">
        <div class="label">当前位置：</div>
        <div class="value">{{ currentAddress }}</div>
      </div>
      <van-button block type="primary" @click="confirmLocation">确认选择</van-button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import 'leaflet/dist/leaflet.css'
import L from 'leaflet'

const emit = defineEmits(['confirm'])
const currentAddress = ref('地图选点位置')
let map = null
let marker = null

onMounted(() => {
  initMap()
})

const initMap = () => {
  // Default to Hangzhou
  const lat = 30.2741
  const lng = 120.1551

  map = L.map('map').setView([lat, lng], 13)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map)

  marker = L.marker([lat, lng], { draggable: true }).addTo(map)

  map.on('click', (e) => {
    const { lat, lng } = e.latlng
    marker.setLatLng([lat, lng])
    updateAddress(lat, lng)
  })

  marker.on('dragend', (e) => {
    const { lat, lng } = e.target.getLatLng()
    updateAddress(lat, lng)
  })
}

const updateAddress = (lat, lng) => {
  // Mock address for now since we don't have a geocoding API key
  currentAddress.value = `经度:${lng.toFixed(4)}, 纬度:${lat.toFixed(4)}`
}

const confirmLocation = () => {
  emit('confirm', currentAddress.value)
}
</script>

<style scoped>
.map-select {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.map-container {
  flex: 1;
  width: 100%;
}

.map-footer {
  padding: 16px;
  background: #fff;
  border-top: 1px solid #ebedf0;
}

.current-address {
  margin-bottom: 16px;
  font-size: 14px;
}

.current-address .label {
  color: #969799;
  margin-bottom: 4px;
}

.current-address .value {
  color: #323233;
  font-weight: 500;
}
</style>
