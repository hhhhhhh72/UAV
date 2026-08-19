<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="test-sites"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      creatable
      add-label="新增"
      @add="openForm()"
    >
      <template #name="{ record }">
        <span class="cell-name">{{ record.name || '-' }}</span>
      </template>
      <template #type="{ record }">
        <a-tag :color="typeTag(record.site_type)" size="small">{{ typeLabel(record.site_type) }}</a-tag>
      </template>
      <template #price="{ record }">
        <span class="cell-amount">{{ formatMoney(record.price_fen) }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无场地数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="场地详情" :width="640" :footer="false" :mask-closable="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="场地名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="场地类型">
            <a-tag :color="typeTag(currentItem.site_type)" size="small">{{ typeLabel(currentItem.site_type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="费用">{{ formatMoney(currentItem.price_fen) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="设施">{{ (currentItem.facilities || []).join('、') || '-' }}</a-descriptions-item>
          <a-descriptions-item label="配套设施" :span="2">{{ currentItem.facilities || '-' }}</a-descriptions-item>
          <a-descriptions-item label="使用规则" :span="2">{{ currentItem.booking_rule || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑场地' : '新增场地'" :width="560" :mask-closable="false" :unmount-on-close="true" :on-before-cancel="beforeClose" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="场地名称" required><a-input v-model="form.name" style="width: 100%" /></a-form-item>
        <a-form-item label="地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.site_type" style="width: 100%">
            <a-option value="flying_field">飞行场地</a-option>
            <a-option value="lab">实验室</a-option>
            <a-option value="anechoic_chamber">消声室</a-option>
            <a-option value="wind_tunnel">风洞</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="available">可用</a-option>
            <a-option value="maintenance">维护中</a-option>
            <a-option value="reserved">已预约</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="费用(元)">
          <a-input-number v-model="form.priceYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="配套设施">
          <a-input v-model="form.facilitiesText" placeholder="如：充电桩、停机坪（逗号分隔）" style="width: 100%" />
        </a-form-item>
        <a-form-item label="使用规则"><a-input v-model="form.booking_rule" type="textarea" :rows="3" style="width: 100%" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="cancelForm">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('test-sites')
const defaultParams = { site_type: '' }

const typeLabel = (t) => ({ flying_field: '飞行场地', lab: '实验室', indoor: '室内场地' }[t] || t || '-')
const typeTag = (t) => ({ flying_field: 'green', lab: 'orange', indoor: 'gray' }[t] || 'gray')

const statusLabel = (s) => ({ available: '可用', maintenance: '维护中', closed: '已关闭' }[s] || s || '-')
const statusTag = (s) => ({ available: 'green', maintenance: 'orange', closed: 'red' }[s] || 'gray')

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索场地名称...', width: 220 },
  { key: 'site_type', label: '场地类型', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'flying_field', label: '飞行场地' },
    { value: 'lab', label: '实验室' },
    { value: 'indoor', label: '室内场地' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '场地名称', dataIndex: 'name', slotName: 'name', minWidth: 180 },
  { title: '类型', dataIndex: 'site_type', slotName: 'type', width: 120 },
  { title: '地区', dataIndex: 'location', minWidth: 140 },
  { title: '费用', dataIndex: 'price_fen', slotName: 'price', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', site_type: 'flying_field', location: '', priceYuan: null, facilitiesText: '', booking_rule: '', status: 'available' })
const resetForm = () => Object.assign(form, { id: '', name: '', site_type: 'flying_field', location: '', priceYuan: null, facilitiesText: '', booking_rule: '', status: 'available' })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id || '',
      name: row.name || '',
      site_type: row.site_type || 'flying_field',
      location: row.location || '',
      priceYuan: row.price_fen == null ? null : Math.round(row.price_fen) / 100,
      facilitiesText: Array.isArray(row.facilities) ? row.facilities.join('、') : '',
      booking_rule: row.booking_rule || '',
      status: row.status || 'available'
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入场地名称'); return }
  formLoading.value = true
  try {
    const payload = {
      name: form.name,
      site_type: form.site_type,
      location: form.location,
      status: form.status,
      booking_rule: form.booking_rule,
      price_fen: form.priceYuan == null || form.priceYuan === '' ? null : Math.round(Number(form.priceYuan) * 100),
      facilities: form.facilitiesText.split(/[,，、]/).map(s => s.trim()).filter(Boolean)
    }
    formEdit.value ? await api.update(form.id, payload) : await api.create(payload)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

// 未保存守卫：Esc/点 X/点遮罩/底部取消 关闭前比对快照，有改动先确认
// 注意：Arco 2.58 无 beforeClose prop（beforeClose 只是 emits 事件），
// 需用 on-before-cancel 拦截用户关闭（X/ESC/遮罩）；底部取消按钮走 cancelForm。
let formSnapshot = ''
const confirmDiscard = () => {
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
}
const cancelForm = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return }
  confirmDiscard()
}
const beforeClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  confirmDiscard()
  return false
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确定删除场地 "${row.name}" 吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); crudRef.value?.reload() } catch { Message.error('删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.cell-amount { font-weight: 600; color: #E96012; }
</style>
