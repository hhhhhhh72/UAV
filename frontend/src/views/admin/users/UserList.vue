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
      <template #phone="{ record }">
        <span v-if="record.phone_masked" class="cell-phone">{{ record.phone_masked }}</span>
        <span v-else class="cell-phone-empty">未绑定</span>
      </template>
      <template #role="{ record }">
        <a-tag :color="roleTagType(record.role)" size="small">{{ record.roleLabel }}</a-tag>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusMeta(record.status).color" size="small">{{ statusMeta(record.status).label }}</a-tag>
        <div v-if="record.purge_after" class="purge-tip">{{ record.purge_after }} 自动清除</div>
      </template>
      <template #password="{ record }">
        <a-tag v-if="record.has_password" color="green" size="small">已设置</a-tag>
        <a-tag v-else color="gray" size="small">未设置</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <!-- 超级管理员（后端按 SUPER_ADMIN_PHONE 标记）与「我自己」都不给操作入口：
               前者受保护，后者是防自锁（把自己降级/删除，平台就没人能管了）。 -->
          <span v-if="record.is_super_admin" class="super-admin-tip">超级管理员</span>
          <template v-else-if="record.id !== currentUserId">
            <a-button v-if="isPlatformAdmin" type="text" size="small" @click="toggleRole(record)">
              {{ record.role === 'platform_admin' ? '取消管理员' : '设为管理员' }}
            </a-button>
            <a-button v-if="isPlatformAdmin" type="text" size="small" @click="openResetPwd(record)">重置密码</a-button>
            <a-button v-if="isPlatformAdmin && record.status === 'active'" type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
            <span v-else-if="record.status === 'deleted'" class="purge-tip">缓冲期内不可恢复</span>
          </template>
          <span v-else class="self-tip">当前账号</span>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无用户数据" />
      </template>
    </CrudList>

    <!-- 重置密码（平台管理员；此前只有建号时能设一次密码，之后谁都改不了） -->
    <a-modal v-model:visible="pwdVisible" title="重置密码" :width="'min(420px, 94vw)'" :mask-closable="false" :unmount-on-close="true">
      <a-form :model="pwdForm" layout="vertical">
        <a-form-item label="账号">
          <a-input :model-value="pwdForm.id" disabled />
        </a-form-item>
        <a-form-item label="新密码" required>
          <a-input-password v-model="pwdForm.password" placeholder="至少 8 位" />
        </a-form-item>
        <div class="pwd-tip">重置后该账号此前的登录全部失效，请把新密码告知本人。</div>
      </a-form>
      <template #footer>
        <a-button @click="pwdVisible = false">取消</a-button>
        <a-button type="primary" :loading="pwdLoading" @click="submitResetPwd">重置</a-button>
      </template>
    </a-modal>

    <!-- 新增用户弹窗 -->
    <a-modal v-model:visible="formVisible" title="新增用户" :width="'min(420px, 94vw)'" :mask-closable="false" :unmount-on-close="true" :on-before-cancel="beforeCancel">
      <a-form :model="form" layout="vertical">
        <a-form-item label="手机号（登录名）" required>
          <a-input v-model="form.phone" :aria-required="true" placeholder="11 位手机号，也是登录账号" :max-length="11" style="width: 100%" />
        </a-form-item>
        <a-form-item label="昵称">
          <a-input v-model="form.name" placeholder="可选，留空按手机号后四位生成" style="width: 100%" />
        </a-form-item>
        <a-form-item label="角色">
          <a-select v-model="form.role" style="width: 100%">
            <a-option value="individual">个人用户</a-option>
            <a-option value="enterprise">企业用户</a-option>
            <a-option value="association_admin">协会管理员</a-option>
            <a-option v-if="isPlatformAdmin" value="platform_admin">平台管理员</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="初始密码" required>
          <a-input-password v-model="form.password" :aria-required="true" placeholder="至少 8 位，创建后告知本人" style="width: 100%" />
        </a-form-item>
        <div class="pwd-tip">账号 ID 由系统生成为 user-手机号；请把手机号与初始密码告知本人，并提醒其登录后自行修改。</div>
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
import { updateUserRole, deleteUser, resetUserPassword } from '@/api/admin/user'
import { useAuth } from '../composables/useAuth'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('users')
const defaultParams = { role: '' }

// isPlatformAdmin：当前登录人是不是平台管理员（决定能不能改别人的角色/密码）；
// currentUserId：当前登录人自己的账号 ID（列表里对自己不给操作入口，防自锁）。
const { isPlatformAdmin, currentUserId } = useAuth()

const roleTagType = (r) => ({ platform_admin: 'green', association_admin: 'orange', enterprise: 'arcoblue', individual: 'gray' }[r] || 'gray')

// 后端 listUsers 不读任何过滤参数，筛选全部移除（列表展示全部用户）
const searchFields = []

const columns = [
  { title: '用户名', dataIndex: 'name', slotName: 'name', minWidth: 160 },
  { title: '手机号', dataIndex: 'phone_masked', slotName: 'phone', width: 140 },
  { title: '角色', dataIndex: 'role', slotName: 'role', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '密码', dataIndex: 'has_password', slotName: 'password', width: 110 },
  { title: '注册时间', dataIndex: 'created_at', width: 160 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const formVisible = ref(false)
const formLoading = ref(false)
const form = reactive({ phone: '', name: '', role: 'individual', password: '' })
const phoneRe = /^1[3-9]\d{9}$/

const openForm = () => {
  form.phone = ''; form.name = ''; form.role = 'individual'; form.password = ''
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
  if (!phoneRe.test(form.phone)) { Message.warning('请输入正确的 11 位手机号'); return }
  if ((form.password || '').length < 8) { Message.warning('初始密码至少 8 位'); return }
  formLoading.value = true
  try {
    await api.create({ phone: form.phone, name: form.name || undefined, role: form.role, password: form.password })
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

// 重置密码：平台管理员给账号设新密码（改完该账号所有端需重新登录）
const pwdVisible = ref(false)
const pwdLoading = ref(false)
const pwdForm = reactive({ id: '', password: '' })
const openResetPwd = (row) => {
  pwdForm.id = row.id
  pwdForm.password = ''
  pwdVisible.value = true
}
const submitResetPwd = async () => {
  if ((pwdForm.password || '').length < 8) { Message.warning('新密码至少 8 位'); return }
  pwdLoading.value = true
  try {
    const res = await resetUserPassword(pwdForm.id, pwdForm.password)
    Message.success(res?.data?.note || '已重置密码')
    pwdVisible.value = false
    pwdForm.password = ''
    crudRef.value?.reload()
  } catch (e) { Message.error(errMsg(e)) }
  finally { pwdLoading.value = false }
}

const statusMeta = (s) => ({
  active: { label: '正常', color: 'green' },
  deleted: { label: '已删除', color: 'gray' },
  banned: { label: '已封禁', color: 'red' },
}[s] || { label: s || '-', color: 'gray' })

// 删除账号（唯一动作，不可恢复）：立即失效 + 7 天缓冲期后自动清除账号行。
// 不删其发布内容：46 张业务表以文本列记用户 ID 且无外键，删账号行不会连带删内容，
// 只会让内容失去作者（生产现存 12 条孤儿需求 + 3 条动态就是这么来的）。
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除账号',
    content: `确定删除「${row.name || row.id}」吗？① 该用户立即无法登录并从平台消失，账号行保留 7 天缓冲期后自动清除，期间不可恢复；② 其在架的发布（需求/职位/商品/课程/动态）会下架；③ 简历、投递、站内信、上传文件、收藏等个人信息会被删除，报名/预约/企业联系人里的联系方式会被清空；④ 工单、合同、资金等履约记录保留。`,
    okText: '删除',
    okButtonProps: { status: 'danger' },
    cancelText: '取消',
    onOk: async () => {
      try {
        const res = await deleteUser(row.id)
        Message.success(res?.data?.note || '已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }

.super-admin-tip { color: #0A66C2; font-size: 12px; font-weight: 500; }

.self-tip { color: var(--color-text-3); font-size: 12px; }

.purge-tip { color: var(--color-text-3); font-size: 11px; line-height: 1.4; margin-top: 2px; }

.pwd-tip { color: var(--color-text-3); font-size: 12px; }

.cell-user { display: flex; align-items: center; gap: 8px; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.cell-phone { font-variant-numeric: tabular-nums; color: var(--color-text-1); }

.cell-phone-empty { color: var(--color-text-3); font-size: 12px; }

.cell-avatar-fallback { background: #C9CDD4; color: #fff; font-size: 13px; }
</style>
