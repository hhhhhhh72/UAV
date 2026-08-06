<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索用户ID..." allow-clear style="width: 220px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="角色" class="form-item">
            <a-select v-model="filterParams.role" style="width: 150px" @change="onSearchSubmit">
              <a-option value="">全部角色</a-option>
              <a-option value="platform_admin">平台管理员</a-option>
              <a-option value="association_admin">协会管理员</a-option>
              <a-option value="enterprise">企业用户</a-option>
              <a-option value="individual">个人用户</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增用户</a-button>
        </a-space>
      </a-form>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="listData"
        :loading="loading"
        row-key="id"
        :pagination="false"
      >
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
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50, 100]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 新增用户弹窗 -->
    <a-modal v-model:visible="dialog.visible" title="新增用户" :width="420" :mask-closable="false" :unmount-on-close="true" @cancel="dialog.visible = false">
      <a-form :model="form" layout="horizontal">
        <a-form-item label="用户ID" required><a-input v-model="form.id" placeholder="输入唯一用户ID" /></a-form-item>
        <a-form-item label="角色">
          <a-select v-model="form.role" style="width: 100%">
            <a-option value="individual">个人用户</a-option>
            <a-option value="enterprise">企业用户</a-option>
            <a-option value="association_admin">协会管理员</a-option>
            <a-option value="platform_admin">平台管理员</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="密码">
          <a-input-password v-model="form.password" placeholder="可选：设置后可用于密码登录" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="dialog.visible = false">取消</a-button>
        <a-button type="primary" :loading="dialog.loading" @click="handleSubmit">创建</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { getUserList, updateUserRole } from '@/api/admin/user'
import { useAuth } from '../composables/useAuth'
import axios from '@/utils/http'

const { isSuperAdmin } = useAuth()

const roleTagType = (r) => ({ platform_admin: 'green', association_admin: 'orange', enterprise: 'arcoblue', individual: 'gray' }[r] || 'gray')

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: getUserList,
  idKey: 'id',
  defaultParams: { role: '' }
})

const columns = [
  { title: '用户ID', dataIndex: 'id', minWidth: 220 },
  { title: '角色', dataIndex: 'role', slotName: 'role', width: 120 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '密码', dataIndex: 'has_password', slotName: 'password', width: 110 },
  { title: '注册时间', dataIndex: 'created_at', width: 160 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const dialog = reactive({ visible: false, loading: false })
const form = reactive({ id: '', role: 'individual', password: '' })

const handleAdd = () => { form.id = ''; form.role = 'individual'; form.password = ''; dialog.visible = true }

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const handleSubmit = async () => {
  if (!form.id) { Message.warning('请输入用户ID'); return }
  dialog.loading = true
  try {
    await axios.post('/api/v1/admin/users', { id: form.id, role: form.role, password: form.password || undefined })
    Message.success('创建成功')
    dialog.visible = false
    loadData()
  } catch (e) { Message.error(errMsg(e)) }
  finally { dialog.loading = false }
}

// 角色切换：平台管理员 ↔ 个人用户
const toggleRole = async (user) => {
  const newRole = user.role === 'platform_admin' ? 'individual' : 'platform_admin'
  try {
    await updateUserRole(user.id, newRole)
    user.role = newRole
    user.roleLabel = newRole === 'platform_admin' ? '平台管理员' : '个人用户'
    Message.success('权限已更新')
  } catch (e) { Message.error(errMsg(e)) }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确认删除用户「${row.id}」？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await axios.delete(`/api/v1/admin/users/${row.id}`)
        Message.success('已删除')
        loadData()
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.super-admin-tip { color: #999; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
