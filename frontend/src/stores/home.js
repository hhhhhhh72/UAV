import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useHomeStore = defineStore('home', () => {
  const banners = ref([])
  const notices = ref([])
  const quickEntries = ref([])
  const demandFeed = ref([])
  const loading = ref(false)

  const systemEntries = [
    { id: 1, name: '会员资源', icon: 'user-o', color: '#1565C0' },
    { id: 2, name: '供需对接', icon: 'exchange-o', color: '#2E7D32' },
    { id: 3, name: '产学研', icon: 'certificate-o', color: '#E65100' },
    { id: 4, name: '合规政策', icon: 'shield-o', color: '#6A1B9A' },
    { id: 5, name: '人才教育', icon: 'bookmark-o', color: '#C62828' },
    { id: 6, name: '活动品牌', icon: 'star-o', color: '#00838F' },
    { id: 7, name: '应急协同', icon: 'warning-o', color: '#D84315' },
  ]

  async function fetchHome() {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/home')
      if (data.success !== false) {
        banners.value = data.banners || []
        notices.value = data.notices || []
        quickEntries.value = data.quickEntries || []
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchDemands(params = {}) {
    const { data } = await http.get('/api/v1/demands', { params })
    demandFeed.value = data.items || data.data || []
    return data
  }

  return { banners, notices, quickEntries, demandFeed, loading, systemEntries, fetchHome, fetchDemands }
})
