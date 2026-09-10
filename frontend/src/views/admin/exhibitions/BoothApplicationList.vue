<template>
  <div class="booth-page">
    <div class="toolbar">
      <a-radio-group v-model="status" type="button" size="small" @change="load">
        <a-radio value="">全部</a-radio>
        <a-radio value="applied">待审核</a-radio>
        <a-radio value="approved">已通过</a-radio>
        <a-radio value="rejected">已驳回</a-radio>
      </a-radio-group>
      <a-button size="small" :loading="loading" @click="load">刷新</a-button>
    </div>

    <a-table :data="rows" :loading="loading" :pagination="false" row-key="id" :bordered="false">
      <template #columns>
        <a-table-column title="展会" :width="200">
          <template #cell="{ record }">{{ expoName(record.exhibition_id) }}</template>
        </a-table-column>
        <a-table-column title="展位号" data-index="booth_number" :width="90" />
        <a-table-column title="参展商" data-index="exhibitor_id" :width="200" />
        <a-table-column title="展品名称" data-index="exhibit_name" :width="160" />
        <a-table-column title="展品简介" data-index="exhibit_desc" :ellipsis="true" :tooltip="true" />
        <a-table-column title="状态" :width="96">
          <template #cell="{ record }">
            <a-tag :color="tagColor(record.status)" size="small">{{ statusLabel[record.status] || record.status }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="申请时间" :width="160">
          <template #cell="{ record }">{{ formatTime(record.created_at) }}</template>
        </a-table-column>
        <a-table-column title="操作" :width="130" fixed="right">
          <template #cell="{ record }">
            <a-space v-if="record.status === 'applied'" :size="4">
              <a-button type="text" size="small" status="success" @click="review(record, 'approve')">通过</a-button>
              <a-button type="text" size="small" status="danger" @click="review(record, 'reject')">驳回</a-button>
            </a-space>
            <span v-else class="done">已处理</span>
          </template>
        </a-table-column>
      </template>
      <template #empty>
        <a-empty description="暂无展位申请" />
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import axios from '@/utils/http'

/* 展位申请审核：GET /api/v1/admin/exhibitions/booths（可按 status 过滤）
   此前管理端没有任何展位申请入口 —— 申请落库了但没人能审（BUG-010）。 */
const BASE = '/api/v1/admin/exhibitions/booths'

const rows = ref([])
const loading = ref(false)
const status = ref('')
const expoMap = ref({})

const statusLabel = { applied: '待审核', approved: '已通过', rejected: '已驳回', paid: '已缴费' }

const tagColor = (s) => ({ applied: 'orangered', approved: 'green', rejected: 'red', paid: 'arcoblue' }[s] || 'gray')

const expoName = (id) => expoMap.value[id] || id || '-'

const formatTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return String(v).slice(0, 16).replace('T', ' ')
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

async function load() {
  loading.value = true
  try {
    const res = await axios.get(BASE, { params: status.value ? { status: status.value } : {} })
    const list = Array.isArray(res.data) ? res.data : (res.data && res.data.data) || []
    rows.value = Array.isArray(list) ? list : []
  } catch (e) {
    Message.error('加载展位申请失败')
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function loadExhibitions() {
  try {
    const res = await axios.get('/api/v1/admin/exhibitions', { params: { page: 1, page_size: 100 } })
    const list = (res.data && res.data.data) || res.data || []
    const map = {}
    for (const e of list) if (e && e.id) map[e.id] = e.title || e.id
    expoMap.value = map
  } catch (e) {
    expoMap.value = {}
  }
}

async function review(record, action) {
  const label = action === 'approve' ? '通过' : '驳回'
  await new Promise((resolve) => {
    Modal.confirm({
      title: `确认${label}该展位申请？`,
      content: `展会：${expoName(record.exhibition_id)}；展位号：${record.booth_number || '-'}；参展商：${record.exhibitor_id || '-'}`,
      okText: label,
      cancelText: '取消',
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
    })
  }).then(async (ok) => {
    if (!ok) return
    try {
      await axios.post(`${BASE}/${encodeURIComponent(record.id)}/review`, { action })
      Message.success(`已${label}`)
      load()
    } catch (e) {
      Message.error(`操作失败：${(e && e.message) || '请稍后重试'}`)
    }
  })
}

onMounted(() => {
  loadExhibitions()
  load()
})
</script>

<style scoped>
.booth-page { padding: 4px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.done { color: var(--color-text-3); font-size: 12px; }
</style>
