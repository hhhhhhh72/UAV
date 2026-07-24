<template>
  <div class="enterprise-list-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">企业审核</span>
        <van-dropdown-menu active-color="#0071e3">
          <van-dropdown-item v-model="statusFilter" :options="statusOptions" @change="fetchEnterprises" />
        </van-dropdown-menu>
      </template>
      <template #actions>
        <van-button type="default" size="small" icon="replay" @click="fetchEnterprises">刷新</van-button>
      </template>
    </DataToolbar>

    <van-empty v-if="enterprises.length === 0" description="暂无企业数据" />

    <van-cell-group v-else inset style="border-radius: var(--card-radius);">
      <van-cell v-for="ent in enterprises" :key="ent.id" is-link @click="showDetail(ent)">
        <template #title>
          <div style="display:flex; flex-direction: column; gap: 4px;">
            <div style="font-weight: 600; color: var(--text-color);">{{ ent.name || '-' }}</div>
            <div style="font-size: 12px; color: var(--text-secondary);">{{ ent.account_name || '-' }}</div>
          </div>
        </template>
        <template #value>
          <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
            <van-tag :type="statusTagType(ent.status)" size="medium">{{ statusLabel(ent.status) }}</van-tag>
            <span style="font-size: 12px; color: var(--text-secondary);">{{ formatDate(ent.created_at) }}</span>
          </div>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- Detail Popup -->
    <van-popup :show="showDetailPopup" @update:show="v => showDetailPopup = v" position="bottom" :style="{ height: '60%' }" round>
      <div class="detail-content" v-if="currentEnterprise">
        <van-cell-group title="企业信息">
          <van-cell title="企业名称" :value="currentEnterprise.name || '-'" />
          <van-cell title="对公账户" :value="currentEnterprise.account_name || '-'" />
          <van-cell title="状态" :value="statusLabel(currentEnterprise.status)" />
          <van-cell title="协会会员" :value="currentEnterprise.is_member ? '是' : '否'" />
          <van-cell v-if="currentEnterprise.license_url" title="营业执照">
            <template #value>
              <van-button size="small" type="primary" plain @click="viewLicense(currentEnterprise.license_url)">查看</van-button>
            </template>
          </van-cell>
          <van-cell title="提交时间" :value="formatDate(currentEnterprise.created_at)" />
        </van-cell-group>

        <van-cell-group v-if="currentEnterprise.status === 'submitted'" title="审核操作">
          <div style="padding: 12px 16px; display: flex; gap: 12px;">
            <van-button type="success" block @click="reviewEnterprise('approved')">通过</van-button>
            <van-button type="danger" block @click="reviewEnterprise('rejected')">驳回</van-button>
          </div>
        </van-cell-group>
      </div>
    </van-popup>

    <!-- License Image Preview -->
    <van-image-preview
      v-model:show="showLicensePreview"
      :images="licenseImages"
      @change="onLicensePreviewChange"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast, showImagePreview } from 'vant'
import DataToolbar from '../components/DataToolbar.vue'

const enterprises = ref([])
const statusFilter = ref('submitted')
const showDetailPopup = ref(false)
const currentEnterprise = ref(null)
const showLicensePreview = ref(false)
const licenseImages = ref([])

const statusOptions = [
  { text: '待审核', value: 'submitted' },
  { text: '已通过', value: 'approved' },
  { text: '已驳回', value: 'rejected' },
  { text: '草稿', value: 'draft' },
  { text: '需补充', value: 'supplement_required' }
]

const statusLabel = (status) => {
  const map = {
    draft: '草稿',
    submitted: '待审核',
    supplement_required: '需补充',
    approved: '已通过',
    rejected: '已驳回'
  }
  return map[status] || status || '-'
}

const statusTagType = (status) => {
  const map = {
    approved: 'success',
    rejected: 'danger',
    submitted: 'warning',
    supplement_required: 'warning',
    draft: 'default'
  }
  return map[status] || 'default'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const fetchEnterprises = async () => {
  try {
    const res = await axios.get('/api/v1/admin/enterprises', {
      params: { status: statusFilter.value }
    })
    let data = res.data
    if (Array.isArray(data)) {
      enterprises.value = data
    } else if (data?.data) {
      enterprises.value = Array.isArray(data.data) ? data.data : []
    } else {
      enterprises.value = []
    }
  } catch (error) {
    showFailToast('获取企业数据失败')
    console.error(error)
  }
}

const showDetail = (ent) => {
  currentEnterprise.value = { ...ent }
  showDetailPopup.value = true
}

const viewLicense = (url) => {
  if (!url) return
  licenseImages.value = [url]
  showLicensePreview.value = true
}

const onLicensePreviewChange = () => {
  // Image preview change handler
}

const reviewEnterprise = async (action) => {
  if (!currentEnterprise.value) return
  try {
    const res = await axios.post(`/api/v1/admin/enterprises/${currentEnterprise.value.id}/review`, {
      action,
      reason: ''
    })
    if (res.data?.success !== false) {
      showSuccessToast(action === 'approved' ? '已通过' : '已驳回')
      currentEnterprise.value.status = action === 'approved' ? 'approved' : 'rejected'
      showDetailPopup.value = false
      fetchEnterprises()
    } else {
      throw new Error(res.data?.message || '操作失败')
    }
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '审核操作失败')
    console.error(error)
  }
}

onMounted(fetchEnterprises)
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
