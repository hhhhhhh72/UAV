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
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无报名记录" />
      </template>
    </CrudList>

    <!-- 报名详情（含证件资料） -->
    <a-modal v-model:visible="detailVisible" title="报名详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="姓名" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="电话">{{ currentItem.phone || '-' }}</a-descriptions-item>
          <a-descriptions-item label="身份证号">{{ currentItem.id_card || '-' }}</a-descriptions-item>
          <a-descriptions-item label="性别">{{ currentItem.gender || '-' }}</a-descriptions-item>
          <a-descriptions-item label="生日">{{ currentItem.birthday || '-' }}</a-descriptions-item>
          <a-descriptions-item label="邮箱">{{ currentItem.email || '-' }}</a-descriptions-item>
          <a-descriptions-item label="学历">{{ currentItem.education || '-' }}</a-descriptions-item>
          <a-descriptions-item label="从业经验" :span="2">{{ currentItem.experience || '-' }}</a-descriptions-item>
          <a-descriptions-item label="课程ID" :span="2">{{ currentItem.course_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证件照">
            <a-image v-if="currentItem.photo_url" :src="fullUrl(currentItem.photo_url)" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="身份证照片">
            <a-image v-if="currentItem.id_card_image" :src="fullUrl(currentItem.id_card_image)" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="无犯罪证明" :span="2">{{ currentItem.no_crime || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 编辑弹窗（基础信息 + 状态） -->
    <a-modal v-model:visible="formVisible" title="编辑报名记录" :width="560" @cancel="formVisible = false">
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
            <a-option value="pending">待审核</a-option>
            <a-option value="approved">已通过</a-option>
            <a-option value="paid">已缴费</a-option>
            <a-option value="enrolled">已入学</a-option>
            <a-option value="rejected">已拒绝</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import { useAdminApi } from '@/api/admin/common'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('enrollments')

// 图片路径同源相对引用：生产 nginx 已反代 /uploads/，开发 vite 已代理 /uploads；
// 早期版本回退拼接 localhost:8080 会在生产环境指向管理员本机导致图片加载失败
const fullUrl = (u) => (u && u.startsWith('http') ? u : u || '')

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
  pending: '待审核', paid: '已缴费'
}
const statusTag = (s) => ({ enrolled: 'green', approved: 'green', paid: 'arcoblue', pending: 'orangered', rejected: 'red' }[s] || 'gray')

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
  { title: '身份证号', dataIndex: 'id_card', width: 190 },
  { title: '课程ID', dataIndex: 'course_id', slotName: 'courseId', minWidth: 180 },
  { title: '性别', dataIndex: 'gender', width: 70 },
  { title: '学历', dataIndex: 'education', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '报名时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 80, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

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
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}
</script>

<style scoped>
.time-text { color: #86909C; font-size: 12px; }
</style>
