import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useEnterpriseStore = defineStore('enterprise', () => {
  const myEnterprises = ref([])
  const loading = ref(false)

  async function fetchMy() {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/enterprises')
      myEnterprises.value = data.items || data.data || data
    } finally { loading.value = false }
  }

  async function apply(form) {
    const { data } = await http.post('/api/v1/enterprises', form)
    return data
  }

  return { myEnterprises, loading, fetchMy, apply }
})
