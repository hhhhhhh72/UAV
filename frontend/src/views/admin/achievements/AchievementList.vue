<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="achievements"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增成果"
      @add="openForm()"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #field="{ record }">
        <a-tag size="small">{{ record.field || '-' }}</a-tag>
      </template>
      <template #stage="{ record }">
        <a-tag :color="stageTag(record.stage)" size="small">{{ stageLabel(record.stage) }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无成果数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="成果详情" :width="'min(640px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="成果名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="领域">{{ currentItem.field || '-' }}</a-descriptions-item>
          <a-descriptions-item label="成果类型">{{ currentItem.achieve_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="所处阶段">
            <a-tag :color="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="联系信息">{{ currentItem.contact_info || '-' }}</a-descriptions-item>
          <a-descriptions-item label="成果描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="附件" :span="2">{{ (currentItem.attachments || []).length }} 份</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑成果' : '新增成果'" :width="'min(560px, 94vw)'" :on-before-cancel="beforeClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="成果名称" required><a-input v-model="form.title" :aria-required="true" style="width: 100%" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" style="width: 100%" /></a-form-item>
        <a-form-item label="成果类型"><a-input v-model="form.achieve_type" placeholder="如：专利 / 样机 / 技术方案" style="width: 100%" /></a-form-item>
        <a-form-item label="阶段">
          <a-select v-model="form.stage" style="width: 100%">
            <a-option value="lab">实验室</a-option>
            <a-option value="pilot">中试</a-option>
            <a-option value="industrialization">产业化</a-option>
            <a-option value="launched">上市</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="联系信息"><a-input v-model="form.contact_info" placeholder="如：电话 / 邮箱" style="width: 100%" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 3 }" style="width: 100%" /></a-form-item>
        <a-form-item label="成果图片">
          <div style="width: 100%">
            <div v-for="(img, i) in form.images" :key="i" style="display: flex; gap: 6px; margin-bottom: 6px">
              <a-input v-model="form.images[i]" placeholder="/uploads/xxx.png" style="flex: 1" />
              <a-button type="text" status="danger" size="small" @click="form.images.splice(i, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <a-button type="outline" size="small" @click="form.images.push('')">+ 添加图片</a-button>
          </div>
        </a-form-item>
        <a-form-item label="附件资料">
          <div style="width: 100%">
            <div v-for="(at, i) in form.attachments" :key="i" style="display: flex; gap: 6px; margin-bottom: 6px">
              <a-input v-model="at.name" placeholder="附件名" style="width: 40%" />
              <a-input v-model="at.size" placeholder="大小" style="width: 20%" />
              <a-input v-model="at.url" placeholder="/uploads/xxx.pdf" style="flex: 1" />
              <a-button type="text" status="danger" size="small" @click="form.attachments.splice(i, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <a-button type="outline" size="small" @click="form.attachments.push({ name: '', size: '', url: '' })">+ 添加附件</a-button>
          </div>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="cancelForm">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">保存</a-button>
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
const api = useAdminApi('achievements')

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialization: '产业化', launched: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'gray', pilot: 'orangered', industrialization: 'green', launched: 'arcoblue' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 批量动作：批量进入中试 / 批量产业化——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'pilot', label: '批量进入中试', status: 'warning', api: (row) => api.update(row.id, { ...row, stage: 'pilot' }) },
  { key: 'industrialization', label: '批量产业化', status: 'success', api: (row) => api.update(row.id, { ...row, stage: 'industrialization' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索成果名称', width: 220 },
  { key: 'field', label: '领域', type: 'select', options: [
    { value: '', label: '全部' },
    { value: '无人机平台', label: '无人机平台' },
    { value: '飞控系统', label: '飞控系统' },
    { value: '导航与定位', label: '导航与定位' },
    { value: '通信链路', label: '通信链路' },
    { value: '载荷与传感器', label: '载荷与传感器' },
    { value: '能源动力', label: '能源动力' },
    { value: '人工智能', label: '人工智能' },
    { value: '新材料', label: '新材料' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '成果名称', dataIndex: 'title', slotName: 'title', minWidth: 180 },
  { title: '领域', dataIndex: 'field', slotName: 'field', width: 120 },
  { title: '阶段', dataIndex: 'stage', slotName: 'stage', width: 100 },
  { title: '成果类型', dataIndex: 'achieve_type', width: 120 },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', field: '', stage: 'lab', achieve_type: '', contact_info: '', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', field: '', stage: 'lab', achieve_type: '', contact_info: '', description: '', images: [], attachments: [] })

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      ...row,
      images: Array.isArray(row.images) ? [...row.images] : [],
      attachments: Array.isArray(row.attachments) ? row.attachments.map(a => ({ ...a })) : []
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入成果名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
      Message.success('创建成功')
    }
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
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
    title: '删除成果',
    content: `确定删除成果「${row.title || row.id}」吗？删除后不可恢复`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.time-text { color: var(--color-text-2); font-size: 12px; }
</style>
