import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useMessageStore = defineStore('message', () => {
  const list = ref([])
  const unread = ref(0)

  async function fetchMessages() {
    const { data } = await http.get('/api/v1/messages')
    list.value = data.items || data.data || []
  }

  async function fetchUnread() {
    const { data } = await http.get('/api/v1/messages/unread-count')
    unread.value = data.count || 0
  }

  async function markRead(id) {
    await http.post(`/api/v1/messages/${id}/read`)
    unread.value = Math.max(0, unread.value - 1)
  }

  return { list, unread, fetchMessages, fetchUnread, markRead }
})
