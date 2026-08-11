<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="enterprises"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      :batch-actions="batchActions"
      :batch-delete="false"
    >
      <template #name="{ record }">
        <span class="cell-name">{{ record.name || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTagType(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openEditModal(record)">编辑</a-button>
          <template v-if="record.status === 'submitted'">
            <a-divider direction="vertical" />
            <a-button type="text" status="success" size="small" @click="openReviewModal(record, 'approved')">通过</a-button>
            <a-button type="text" status="danger" size="small" @click="openReviewModal(record, 'rejected')">驳回</a-button>
          </template>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无企业数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="企业详情" :width="640" :footer="false" :mask-closable="false">
      <template v-if="currentEnterprise">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="企业名称" :span="2">{{ currentEnterprise.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="对公账户">{{ currentEnterprise.account_name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="法定代表人">{{ currentEnterprise.legal_person || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系电话">{{ currentEnterprise.contact_phone || '-' }}</a-descriptions-item>
          <a-descriptions-item label="信用代码" :span="2">{{ currentEnterprise.credit_code || '-' }}</a-descriptions-item>
          <a-descriptions-item label="产业分类">{{ currentEnterprise.industry_category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="企业规模">{{ currentEnterprise.scale || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地址" :span="2">{{ currentEnterprise.address || '-' }}</a-descriptions-item>
          <a-descriptions-item label="企业简介" :span="2">{{ currentEnterprise.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="审核状态">
            <a-tag :color="statusTagType(currentEnterprise.status)" size="small">{{ statusLabel(currentEnterprise.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentEnterprise.review_comment" label="审核意见" :span="2">
            {{ currentEnterprise.review_comment }}
          </a-descriptions-item>
          <a-descriptions-item label="协会会员">{{ currentEnterprise.is_member ? '是' : '否' }}</a-descriptions-item>
          <a-descriptions-item label="提交时间" :span="2">{{ formatDate(currentEnterprise.created_at) }}</a-descriptions-item>
          <a-descriptions-item v-if="currentEnterprise.license_url" label="营业执照" :span="2">
            <a-image
              :src="currentEnterprise.license_url"
              :width="200"
              :preview-props="{ srcList: [currentEnterprise.license_url] }"
              fit="contain"
              class="license-img"
            />
          </a-descriptions-item>
        </a-descriptions>

        <!-- 审核操作 -->
        <div v-if="currentEnterprise.status === 'submitted'" class="review-actions">
          <a-divider />
          <a-button type="primary" status="success" @click="openReviewModal(currentEnterprise, 'approved')">审核通过</a-button>
          <a-button type="primary" status="danger" @click="openReviewModal(currentEnterprise, 'rejected')">驳回</a-button>
        </div>
      </template>
    </a-modal>

    <!-- 编辑弹窗：管理员编辑企业档案（PRD FR-2.1 全部字段） -->
    <a-modal v-model:visible="editModal.visible" title="编辑企业档案" :width="720" @ok="confirmEdit" @cancel="resetEditModal" :confirm-loading="editModal.loading" :mask-closable="false">
      <a-alert
        v-if="editModal.status && editModal.status !== 'draft' && editModal.status !== 'supplement_required'"
        type="warning"
        class="edit-alert"
        :message="`当前状态：${statusLabel(editModal.status)}。编辑保存后将回到「待审核」，需重新审核后生效。`"
      />
      <a-form layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="企业名称" required>
              <a-input v-model="editModal.form.name" placeholder="请输入企业名称" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="统一社会信用代码" required>
              <a-input v-model="editModal.form.credit_code" placeholder="请输入信用代码" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="法定代表人">
              <a-input v-model="editModal.form.legal_person" placeholder="请输入法人代表" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="联系人">
              <a-input v-model="editModal.form.contact_person" placeholder="请输入联系人" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="联系电话">
              <a-input v-model="editModal.form.contact_phone" placeholder="请输入联系电话" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="邮箱">
              <a-input v-model="editModal.form.email" placeholder="请输入邮箱" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="产业分类（多选，逗号分隔）">
              <a-select v-model="editModal.form.industry_categories" multiple allow-clear placeholder="请选择分类">
                <a-option v-for="c in CATEGORY_OPTIONS" :key="c" :value="c">{{ c }}</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="能力标签（多选，逗号分隔）">
              <a-select v-model="editModal.form.capability_tags" multiple allow-clear placeholder="请选择能力标签">
                <a-option v-for="t in TAG_OPTIONS" :key="t" :value="t">{{ t }}</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="企业规模">
              <a-input v-model="editModal.form.scale" placeholder="如：50-100人" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="成立时间">
              <a-input v-model="editModal.form.founded_at" placeholder="如：2018-03" />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="注册地">
              <a-input v-model="editModal.form.address" placeholder="请输入办公地址" />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="企业简介">
              <a-textarea v-model="editModal.form.description" :rows="3" placeholder="请输入企业简介" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>

    <!-- 审核意见弹窗：驳回必填理由（持久化到 review_comment，用户端可见） -->
    <a-modal v-model:visible="reviewModal.visible" :title="reviewModal.action === 'approved' ? '审核通过' : '驳回企业'" :width="520" @ok="confirmReview" @cancel="resetReviewModal" :confirm-loading="reviewModal.loading">
      <a-form layout="vertical">
        <a-form-item :label="reviewModal.action === 'approved' ? '审核意见（选填）' : '驳回理由（必填，将展示给申请人）'" required>
          <a-textarea
            v-model="reviewModal.reason"
            :rows="4"
            max-length="200"
            show-word-limit
            :placeholder="reviewModal.action === 'approved' ? '可填写审核通过说明（选填）' : '请填写驳回原因，申请人可在小程序查看'"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { showSuccessToast, showFailToast } from '@/utils/feedback'
import { reviewEnterprise, updateEnterprise } from '@/api/admin/enterprise'
import CrudList from '../components/CrudList.vue'

// PRD FR-2.1：企业分类（8 类多选）与能力标签（预设标签库），与小程序 register 页一致
const CATEGORY_OPTIONS = ['整机研发', '零部件制造', '飞控系统', '载荷设备', '运营服务', '实训院校', '通航机场', '检测机构']
const TAG_OPTIONS = ['航拍服务', '农林植保', '电力巡检', '测绘勘察', '应急救援', '物流配送', '巡防安防', '消防救援', '桥梁检测', '环保监测']

const crudRef = ref()
const defaultParams = { status: 'all' }

// --- 状态映射 ---
const statusLabel = (s) => ({
  draft: '草稿',
  submitted: '待审核',
  supplement_required: '需补充',
  approved: '已通过',
  rejected: '已驳回'
}[s] || s || '-')

const statusTagType = (s) => ({
  approved: 'green',
  rejected: 'red',
  submitted: 'orange',
  supplement_required: 'orange',
  draft: 'gray'
}[s] || 'gray')

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 批量动作：仅批量通过（无需理由）；驳回必须逐条填写理由，避免无说明驳回
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: (row) => reviewEnterprise(row.id, { action: 'approved', reason: '' }) }
]

// 后端 listEnterprises 仅支持 status 过滤，keyword 无效已移除
const searchFields = [
  { key: 'status', label: '审核状态', type: 'select', options: [
    { value: 'all', label: '全部状态' },
    { value: 'submitted', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'rejected', label: '已驳回' },
    { value: 'draft', label: '草稿' },
    { value: 'supplement_required', label: '需补充' }
  ]}
]

const columns = [
  { title: '企业名称', dataIndex: 'name', slotName: 'name', minWidth: 180 },
  { title: '对公账户', dataIndex: 'account_name', width: 160 },
  { title: '法人', dataIndex: 'legal_person', width: 100 },
  { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
  { title: '审核状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

// --- 详情弹窗 ---
const detailVisible = ref(false)
const currentEnterprise = ref(null)

const showDetail = (ent) => {
  currentEnterprise.value = { ...ent }
  detailVisible.value = true
}

// --- 编辑弹窗 ---
const splitTags = (str) => (str ? String(str).split(',').map((t) => t.trim()).filter(Boolean) : [])

const editModal = ref({ visible: false, record: null, status: '', form: {}, loading: false })

const openEditModal = (ent) => {
  editModal.value = {
    visible: true,
    record: ent,
    status: ent.status,
    loading: false,
    form: {
      name: ent.name || '',
      credit_code: ent.credit_code || '',
      legal_person: ent.legal_person || '',
      contact_person: ent.contact_person || '',
      contact_phone: ent.contact_phone || '',
      email: ent.email || '',
      industry_categories: splitTags(ent.industry_category),
      capability_tags: splitTags(ent.capability_tags),
      scale: ent.scale || '',
      founded_at: ent.founded_at || '',
      address: ent.address || '',
      description: ent.description || '',
    },
  }
}

const resetEditModal = () => {
  editModal.value.visible = false
  editModal.value.record = null
  editModal.value.loading = false
}

const confirmEdit = async () => {
  const { record, form } = editModal.value
  if (!form.name || !form.name.trim()) {
    showFailToast('请填写企业名称')
    return
  }
  if (!form.credit_code || !form.credit_code.trim()) {
    showFailToast('请填写统一社会信用代码')
    return
  }
  editModal.value.loading = true
  try {
    await updateEnterprise(record.id, {
      name: form.name.trim(),
      credit_code: form.credit_code.trim(),
      legal_person: form.legal_person,
      contact_person: form.contact_person,
      contact_phone: form.contact_phone,
      email: form.email,
      industry_category: form.industry_categories.join(','),
      capability_tags: form.capability_tags.join(','),
      scale: form.scale,
      founded_at: form.founded_at,
      address: form.address,
      description: form.description,
    })
    showSuccessToast(record.status === 'approved' ? '已保存，将重新审核' : '已保存')
    resetEditModal()
    crudRef.value?.reload()
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '保存失败')
  } finally {
    editModal.value.loading = false
  }
}

// --- 审核操作（带意见弹窗） ---
const reviewModal = ref({ visible: false, action: 'approved', record: null, reason: '', loading: false })

const openReviewModal = (ent, action) => {
  reviewModal.value = { visible: true, action, record: ent, reason: '', loading: false }
}

const resetReviewModal = () => {
  reviewModal.value.visible = false
  reviewModal.value.reason = ''
  reviewModal.value.record = null
}

const confirmReview = async () => {
  const { record, action, reason } = reviewModal.value
  if (action === 'rejected' && !reason.trim()) {
    showFailToast('请填写驳回理由')
    return
  }
  reviewModal.value.loading = true
  try {
    await reviewEnterprise(record.id, { action, reason: reason.trim() })
    showSuccessToast(action === 'approved' ? '审核通过' : '已驳回')
    if (currentEnterprise.value?.id === record.id) {
      currentEnterprise.value.status = action === 'approved' ? 'approved' : 'rejected'
      currentEnterprise.value.review_comment = reason.trim()
    }
    detailVisible.value = false
    resetReviewModal()
    crudRef.value?.reload()
  } catch (error) {
    showFailToast(error?.response?.data?.message || error?.message || '操作失败')
  } finally {
    reviewModal.value.loading = false
  }
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.time-text { color: #86909C; font-size: 12px; }

.license-img { cursor: pointer; }

.edit-alert { margin-bottom: 16px; }

.review-actions {
  text-align: center;
  padding-top: 16px;
}
</style>
