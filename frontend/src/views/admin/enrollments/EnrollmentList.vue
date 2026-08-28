<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="enrollments"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      :selectable="false"
      :batch-delete="false"
    >
      <template #courseId="{ record }">
        <span>{{ record.course_id || '-' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #idCard="{ record }">
        <span>{{ maskIdCard(record.id_card) || '-' }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <!-- 完成结业：仅 enrolled/paid 可结（释放托管学费 + 发证；pending/approved/rejected 不可） -->
          <a-button
            v-if="record.status === 'enrolled' || record.status === 'paid'"
            type="text"
            size="small"
            status="success"
            @click="completeEnrollment(record)"
          >完成结业</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无报名记录" />
      </template>
    </CrudList>

    <!-- 报名详情（含证件资料） -->
    <a-modal v-model:visible="detailVisible" title="报名详情" :width="'min(640px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="姓名" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="电话">{{ revealPII ? (currentItem.phone || '-') : (maskPhone(currentItem.phone) || '-') }}</a-descriptions-item>
          <a-descriptions-item label="身份证号">{{ revealPII ? (currentItem.id_card || '-') : (maskIdCard(currentItem.id_card) || '-') }}</a-descriptions-item>
          <a-descriptions-item label="敏感信息" :span="2">
            <a-checkbox v-model="revealPII" :disabled="!currentItem">显示完整证件信息（谨慎操作）</a-checkbox>
          </a-descriptions-item>
          <a-descriptions-item label="性别">{{ currentItem.gender || '-' }}</a-descriptions-item>
          <a-descriptions-item label="生日">{{ currentItem.birthday || '-' }}</a-descriptions-item>
          <a-descriptions-item label="邮箱">{{ currentItem.email || '-' }}</a-descriptions-item>
          <a-descriptions-item label="学历">{{ currentItem.education || '-' }}</a-descriptions-item>
          <a-descriptions-item label="从业经验" :span="2">{{ currentItem.experience || '-' }}</a-descriptions-item>
          <a-descriptions-item label="课程ID" :span="2">{{ currentItem.course_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证件照">
            <a-image v-if="currentItem.photo_url" :src="fullUrl(currentItem.photo_url)" alt="证件照" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="身份证照片">
            <a-image v-if="currentItem.id_card_image" :src="fullUrl(currentItem.id_card_image)" alt="身份证照片" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="无犯罪证明" :span="2">{{ currentItem.no_crime || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 编辑弹窗（基础信息 + 状态） -->
    <a-modal v-model:visible="formVisible" title="编辑报名记录" :width="'min(560px, 94vw)'" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="姓名"><a-input v-model="form.name" style="width: 100%" /></a-form-item>
        <a-form-item label="电话"><a-input v-model="form.phone" style="width: 100%" /></a-form-item>
        <a-form-item label="身份证号"><a-input v-model="form.id_card" style="width: 100%" /></a-form-item>
        <a-form-item label="性别"><a-input v-model="form.gender" placeholder="如：男" style="width: 100%" /></a-form-item>
        <a-form-item label="生日"><a-date-picker v-model="form.birthday" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" /></a-form-item>
        <a-form-item label="邮箱"><a-input v-model="form.email" style="width: 100%" /></a-form-item>
        <a-form-item label="学历"><a-input v-model="form.education" placeholder="如：本科" style="width: 100%" /></a-form-item>
        <a-form-item label="从业经验"><a-input v-model="form.experience" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <!-- 无 approved：与后端状态机冲突（approved 既不能结业也不能回退，edit 置 approved 即卡死），
                 结业走专用按钮 POST /enrollments/{id}/complete（enrolled/paid → completed） -->
            <a-option value="pending">待审核</a-option>
            <a-option value="paid">已缴费</a-option>
            <a-option value="enrolled">已入学</a-option>
            <a-option value="rejected">已拒绝</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="handleCancel">取消</a-button>
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
import http from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('enrollments')

// 图片路径同源相对引用：生产 nginx 已反代 /uploads/，开发 vite 已代理 /uploads；
// 早期版本回退拼接 localhost:8080 会在生产环境指向管理员本机导致图片加载失败
const fullUrl = (u) => (u && u.startsWith('http') ? u : u || '')

// 身份证/电话脱敏：列表与详情默认脱敏（安全审计 P1——详情默认不显完整身份证）
const maskIdCard = (c) => {
  if (!c) return ''
  if (c.length <= 10) return c
  return c.slice(0, 6) + '********' + c.slice(-4)
}
const maskPhone = (p) => {
  if (!p) return ''
  if (p.length < 7) return p
  return p.slice(0, 3) + '****' + p.slice(-4)
}
// 详情"显示完整"开关：默认脱敏，显式点击才展开（敏感凭证最小暴露）
const revealPII = ref(false)

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 报名记录为纯查看型数据，无批量动作（selectable/batch-delete 已关闭）
const batchActions = []

const statusLabel = {
  enrolled: '已入学', approved: '已通过', rejected: '已拒绝',
  pending: '待审核', paid: '已缴费', completed: '已完成'
}
const statusTag = (s) => ({ enrolled: 'green', approved: 'green', paid: 'arcoblue', pending: 'orangered', rejected: 'red', completed: 'gray' }[s] || 'gray')

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索姓名/电话...', width: 200 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'pending', label: '待审核' },
    { value: 'approved', label: '已通过' },
    { value: 'paid', label: '已缴费' },
    { value: 'enrolled', label: '已入学' },
    { value: 'rejected', label: '已拒绝' }
  ]}
]

const columns = [
  { title: '姓名', dataIndex: 'name', minWidth: 100 },
  { title: '电话', dataIndex: 'phone', width: 130 },
  { title: '身份证号', dataIndex: 'id_card', slotName: 'idCard', width: 190 },
  { title: '课程ID', dataIndex: 'course_id', slotName: 'courseId', minWidth: 180 },
  { title: '性别', dataIndex: 'gender', width: 70 },
  { title: '学历', dataIndex: 'education', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '报名时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 170, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { revealPII.value = false; currentItem.value = row; detailVisible.value = true }

// 完成结业：POST /api/v1/enrollments/{id}/complete（后端释放托管学费 + 发放结业证书，
// enrolled/paid → completed；completed 幂等重试可补齐缺失的释放/发证步骤）
const completeEnrollment = (row) => {
  Modal.confirm({
    title: '完成结业',
    content: `确认学员「${row.name || '-'}」已完成结业？将释放托管学费并发放结业证书。`,
    okText: '确认完成',
    cancelText: '取消',
    onOk: async () => {
      try {
        await http.post('/api/v1/enrollments/' + encodeURIComponent(row.id) + '/complete')
        Message.success('已完成结业')
        crudRef.value?.reload()
      } catch (e) {
        Message.error(e?.response?.data?.message || '操作失败')
      }
    },
  })
}

// 编辑：可编辑字段 + 图片/证明字段（photo_url/id_card_image/no_crime）从原记录带入
// —— 后端 PUT 为全字段覆盖，不提交会清空原值（课程ID不可改，后端忽略）
const formVisible = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', phone: '', id_card: '', gender: '', birthday: '', email: '', education: '', experience: '', photo_url: '', id_card_image: '', no_crime: '', status: 'pending' })
const resetForm = () => Object.assign(form, { id: '', name: '', phone: '', id_card: '', gender: '', birthday: '', email: '', education: '', experience: '', photo_url: '', id_card_image: '', no_crime: '', status: 'pending' })

// 记录中 birthday 为 ISO 时间串或空，归一化为 YYYY-MM-DD 供 a-date-picker 显示
const toDateInput = (d) => (d ? String(d).slice(0, 10) : '')

const openForm = (row) => {
  resetForm()
  Object.assign(form, {
    id: row.id, name: row.name || '', phone: row.phone || '', id_card: row.id_card || '',
    gender: row.gender || '', birthday: toDateInput(row.birthday), email: row.email || '',
    education: row.education || '', experience: row.experience || '',
    photo_url: row.photo_url || '', id_card_image: row.id_card_image || '',
    no_crime: row.no_crime || '', status: row.status || 'pending'
  })
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入姓名'); return }
  formLoading.value = true
  try {
    await api.update(form.id, {
      name: form.name, phone: form.phone, id_card: form.id_card, gender: form.gender,
      birthday: form.birthday, email: form.email, education: form.education,
      experience: form.experience, photo_url: form.photo_url, id_card_image: form.id_card_image,
      no_crime: form.no_crime, status: form.status
    })
    Message.success('更新成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

// 未保存守卫：X/Esc/遮罩/取消 关闭前，表单有改动则确认（onBeforeCancel 返回 false 阻断关闭）
let formSnapshot = ''
const guardClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}
// 底部取消按钮：走守卫，确认无改动/放弃修改后才真正关闭
const handleCancel = () => { if (guardClose()) formVisible.value = false }
</script>

<style scoped>
.time-text { color: var(--color-text-2); font-size: 12px; }
</style>
