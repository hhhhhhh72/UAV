<template>
  <div class="demand-list-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">需求管理</span>
        <van-dropdown-menu active-color="#0071e3">
          <van-dropdown-item v-model="statusFilter" :options="statusOptions" @change="fetchDemands" />
        </van-dropdown-menu>
      </template>
      <template #actions>
        <van-button type="default" size="small" icon="replay" @click="fetchDemands">刷新</van-button>
      </template>
    </DataToolbar>

    <van-empty v-if="demands.length === 0" description="暂无需求数据" />

    <van-cell-group v-else inset style="border-radius: var(--card-radius);">
      <van-cell v-for="d in demands" :key="d.id" is-link @click="showDetail(d)">
        <template #title>
          <div style="display:flex; flex-direction: column; gap: 4px;">
            <div style="font-weight: 600; color: var(--text-color);">{{ d.title || '无标题' }}</div>
            <div style="font-size: 12px; color: var(--text-secondary);">
              {{ d.publisher_name || '-' }} · {{ bizTypeLabel(d.biz_type) }}
            </div>
          </div>
        </template>
        <template #value>
          <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
            <van-tag :type="statusTagType(d.status)" size="medium">{{ demandStatusLabel(d.status) }}</van-tag>
            <span style="font-size: 12px; color: var(--text-secondary);">
              {{ d.budget_fen ? '¥' + (d.budget_fen / 100).toFixed(0) : '-' }}
              · {{ formatDate(d.created_at) }}
            </span>
          </div>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- Detail Popup -->
    <van-popup :show="showDetailPopup" @update:show="v => showDetailPopup = v" position="bottom" :style="{ height: '60%' }" round>
      <div class="detail-content" v-if="currentDemand">
        <van-cell-group title="需求信息">
          <van-cell title="标题" :value="currentDemand.title || '-'" />
          <van-cell title="发布者" :value="currentDemand.publisher_name || '-'" />
          <van-cell title="业务类型" :value="bizTypeLabel(currentDemand.biz_type)" />
          <van-cell title="地区" :value="currentDemand.district || '-'" />
          <van-cell title="预算" :value="currentDemand.budget_fen ? '¥' + (currentDemand.budget_fen / 100).toFixed(2) : '-'" />
          <van-cell title="描述" :label="currentDemand.description || '-'" />
          <van-cell title="状态" :value="demandStatusLabel(currentDemand.status)" />
          <van-cell title="提交时间" :value="formatDate(currentDemand.created_at)" />
        </van-cell-group>

        <van-cell-group v-if="currentDemand.status === 'pending'" title="审核操作">
          <div style="padding: 12px 16px; display: flex; gap: 12px;">
            <van-button type="success" block @click="reviewDemand('approve')">通过</van-button>
            <van-button type="danger" block @click="reviewDemand('reject')">驳回</van-button>
          </div>
        </van-cell-group>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast } from 'vant'
import DataToolbar from '../components/DataToolbar.vue'

const demands = ref([])
const statusFilter = ref('pending')
const showDetailPopup = ref(false)
const currentDemand = ref(null)

const statusOptions = [
  { text: '待审核', value: 'pending' },
  { text: '已发布', value: 'published' },
  { text: '已匹配', value: 'matched' },
  { text: '已完成', value: 'completed' },
  { text: '已取消', value: 'cancelled' },
  { text: '已驳回', value: 'rejected' }
]

const demandStatusLabel = (status) => {
  const map = {
    pending: '待审核',
    published: '已发布',
    matched: '已匹配',
    completed: '已完成',
    cancelled: '已取消',
    rejected: '已驳回'
  }
  return map[status] || status || '-'
}

const bizTypeLabel = (bizType) => {
  const map = {
    aerial_photo: '航拍摄影',
    mapping: '测绘',
    inspection: '巡检',
    agriculture: '植保',
    logistics: '物流配送',
    training: '培训',
    competition: '赛事',
    other: '其他'
  }
  return map[bizType] || bizType || '-'
}

const statusTagType = (status) => {
  const map = {
    published: 'primary',
    matched: 'success',
    completed: 'success',
    pending: 'warning',
    cancelled: 'default',
    rejected: 'danger'
  }
  return map[status] || 'default'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const fetchDemands = async () => {
  try {
    const res = await axios.get('/api/v1/admin/demands', {
      params: { status: statusFilter.value }
    })
    let data = res.data
    if (Array.isArray(data)) {
      demands.value = data
    } else if (data?.data) {
      demands.value = Array.isArray(data.data) ? data.data : []
    } else {
      demands.value = []
    }
  } catch (error) {
    showFailToast('获取需求数据失败')
    console.error(error)
  }
}

const showDetail = (demand) => {
  currentDemand.value = { ...demand }
  showDetailPopup.value = true
}

const reviewDemand = async (action) => {
  if (!currentDemand.value) return
  try {
    const endpoint = action === 'approve'
      ? `/api/v1/admin/demands/${currentDemand.value.id}/approve`
      : `/api/v1/admin/demands/${currentDemand.value.id}/review`
    const body = action === 'approve'
      ? {}
      : { action: 'reject', reason: '' }
    const res = await axios.post(endpoint, body)
    if (res.data?.success !== false) {
      showSuccessToast(action === 'approve' ? '已通过' : '已驳回')
      currentDemand.value.status = action === 'approve' ? 'published' : 'rejected'
      showDetailPopup.value = false
      fetchDemands()
    } else {
      throw new Error(res.data?.message || '操作失败')
    }
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '审核操作失败')
    console.error(error)
  }
}

onMounted(fetchDemands)
</script>

<style scoped>
.toolbar-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  margin-right: 8px;
}

.detail-content {
  padding: 16px 0;
}
</style>
