<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="users"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      creatable
      add-label="新增用户"
      @add="openForm()"
    >
      <template #name="{ record }">
        <div class="cell-user">
          <a-avatar v-if="record.avatar_url" :size="28" :image-url="record.avatar_url" />
          <a-avatar v-else :size="28" class="cell-avatar-fallback">{{ (record.name || '?').charAt(0) }}</a-avatar>
          <span class="cell-name">{{ record.name || '-' }}</span>
        </div>
      </template>
      <template #role="{ record }">
        <a-tag :color="roleTagType(record.role)" size="small">{{ record.roleLabel }}</a-tag>
      </template>
      <template #status="{ record }">
        <a-tag :color="record.status === 'active' ? 'green' : 'gray'" size="small">{{ record.status === 'active' ? '正常' : record.status }}</a-tag>
      </template>
      <template #password="{ record }">
        <a-tag v-if="record.has_password" color="green" size="small">已设置</a-tag>
        <a-tag v-else color="gray" size="small">未设置</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button v-if="isSuperAdmin && record.id !== 'admin'" type="text" size="small" @click="toggleRole(record)">
            {{ record.role === 'platform_admin' ? '取消管理员' : '设为管理员' }}
          </a-button>
          <span v-else-if="record.id === 'admin'" class="super-admin-tip">超级管理员</span>
          <a-button v-if="record.id !== 'admin'" type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无用户数据" />
      </template>
    </CrudList>

    <!-- 新增用户弹窗 -->
    <a-modal v-model:visible="formVisible" title="新增用户" :width="420" :mask-closable="false" :unmount-on-close="true" :on-before-cancel="beforeCancel">
      <a-form :model="form" layout="vertical">
        <a-form-item label="用户ID" required><a-input v-model="form.id" placeholder="输入唯一用户ID" style="width: 100%" /></a-form-item>
        <a-form-item label="角色">
          <a-select v-model="form.role" style="width: 100%">
            <a-option value="individual">个人用户</a-option>
            <a-option value="enterprise">企业用户</a-option>
            <a-option value="association_admin">协会管理员</a-option>
            <a-option value="platform_admin">平台管理员</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="密码">
          <a-input-password v-model="form.password" placeholder="可选：设置后可用于密码登录" style="width: 100%" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="beforeCancel">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">创建</a-button>
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
import { updateUserRole } from '@/api/admin/user'
import { useAuth } from '../composables/useAuth'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('users')
const defaultParams = { role: '' }

const { isSuperAdmin } = useAuth()

const roleTagType = (r) => ({ platform_admin: 'green', association_admin: 'orange', enterprise: 'arcoblue', individual: 'gray' }[r] || 'gray')

// 后端 listUsers 不读任何过滤参数，筛选全部移除（列表展示全部用户）
const searchFields = []

const columns = [
  { title: '用户名', dataIndex: 'name', slotName: 'name', minWidth: 160 },
  { title: '用户ID', dataIndex: 'id', minWidth: 220 },
  { title: '角色', dataIndex: 'role', slotName: 'role', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '密码', dataIndex: 'has_password', slotName: 'password', width: 110 },
  { title: '注册时间', dataIndex: 'created_at', width: 160 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const formVisible = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', role: 'individual', password: '' })

const openForm = () => {
  form.id = ''; form.role = 'individual'; form.password = ''
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

// 未保存守卫：Esc/点 X/点取消关闭前，若表单有改动则确认，避免输入全丢
let formSnapshot = ''
const beforeCancel = () => {
  if (JSON.stringify(form) === formSnapshot) { formVisible.value = false; return true }
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.id) { Message.warning('请输入用户ID'); return }
  formLoading.value = true
  try {
    await api.create({ id: form.id, role: form.role, password: form.password || undefined })
    Message.success('创建成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

// 角色切换：平台管理员 ↔ 个人用户（直接生效的高危操作，必须先确认）
const toggleRole = (user) => {
  const newRole = user.role === 'platform_admin' ? 'individual' : 'platform_admin'
  const promote = newRole === 'platform_admin'
  Modal.confirm({
    title: promote ? '设为平台管理员' : '取消平台管理员',
    content: promote
      ? `确定将用户「${user.name || user.id}」设为平台管理员吗？该用户将获得全部管理权限`
      : `确定取消用户「${user.name || user.id}」的平台管理员权限吗？`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await updateUserRole(user.id, newRole)
        user.role = newRole
        user.roleLabel = newRole === 'platform_admin' ? '平台管理员' : '个人用户'
        Message.success('权限已更新')
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确认删除用户「${row.id}」？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }

.super-admin-tip { color: #999; font-size: 12px; }

.cell-user { display: flex; align-items: center; gap: 8px; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.cell-avatar-fallback { background: #C9CDD4; color: #fff; font-size: 13px; }
</style>
