import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useDemandStore = defineStore('demand', () => {
  const list = ref([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchDemands(params = {}) {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/demands', { params })
      list.value = data.items || data.data || []
      total.value = data.total || list.value.length
    } finally { loading.value = false }
  }

  async function fetchDetail(id) {
    const { data } = await http.get(`/api/v1/demands/${id}`)
    return data
  }

  async function createBid(demandId, payload) {
    const { data } = await http.post(`/api/v1/demands/${demandId}/applications`, payload)
    return data
  }

  return { list, total, loading, fetchDemands, fetchDetail, createBid }
})
