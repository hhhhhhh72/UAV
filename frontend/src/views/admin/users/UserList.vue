<template>
  <div class="user-list-page">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="handleAdd">新增用户</el-button>
    </div>

    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索用户ID..." clearable style="width: 220px"
          @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.role" clearable style="width: 150px" @change="onSearchSubmit">
          <el-option label="全部角色" value="" />
          <el-option label="平台管理员" value="platform_admin" />
          <el-option label="协会管理员" value="association_admin" />
          <el-option label="企业用户" value="enterprise" />
          <el-option label="个人用户" value="individual" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="id" label="用户ID" min-width="220" />
        <el-table-column prop="roleLabel" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">{{ row.roleLabel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status === 'active' ? '正常' : row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="密码" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.has_password" type="success" size="small">已设置</el-tag>
            <el-tag v-else type="info" size="small">未设置</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="isSuperAdmin && row.id !== 'admin'" link type="primary" size="small" @click="toggleRole(row)">
              {{ row.role === 'platform_admin' ? '取消管理员' : '设为管理员' }}
            </el-button>
            <span v-else-if="row.id === 'admin'" style="color:#999;font-size:12px">超级管理员</span>
            <el-button v-if="row.id !== 'admin'" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无用户数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50, 100]" :total="total" layout="total,sizes,prev,pager,next,jumper" background
        @size-change="loadData" @current-change="loadData" />
    </div>

    <el-dialog v-model="dialog.visible" title="新增用户" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户ID" required><el-input v-model="form.id" placeholder="输入唯一用户ID" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" style="width:100%">
            <el-option label="个人用户" value="individual" />
            <el-option label="企业用户" value="enterprise" />
            <el-option label="协会管理员" value="association_admin" />
            <el-option label="平台管理员" value="platform_admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="可选：设置后可用于密码登录" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible=false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="dialog.loading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { getUserList, updateUserRole } from '@/api/admin/user'
import { useAuth } from '../composables/useAuth'
import axios from '@/utils/http'

const { isSuperAdmin } = useAuth()

const roleTagType = (r) => ({ platform_admin:'success', association_admin:'warning', enterprise:'', individual:'info' }[r]||'info')

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: getUserList,
  idKey: 'id',
  defaultParams: { role: '' }
})

const dialog = reactive({ visible: false, loading: false })
const form = reactive({ id: '', role: 'individual', password: '' })

const handleAdd = () => { form.id = ''; form.role = 'individual'; form.password = ''; dialog.visible = true }

const handleSubmit = async () => {
  if (!form.id) { ElMessage.warning('请输入用户ID'); return }
  dialog.loading = true
  try {
    await axios.post('/api/v1/admin/users', { id: form.id, role: form.role, password: form.password || undefined })
    ElMessage.success('创建成功')
    dialog.visible = false
    loadData()
  } catch (e) { ElMessage.error(e?.response?.data?.message || '创建失败') }
  finally { dialog.loading = false }
}

const toggleRole = async (user) => {
  const newRole = user.role === 'platform_admin' ? 'individual' : 'platform_admin'
  try {
    await updateUserRole(user.id, newRole)
    user.role = newRole
    user.roleLabel = newRole === 'platform_admin' ? '平台管理员' : '个人用户'
    ElMessage.success('权限已更新')
  } catch (e) { ElMessage.error(e?.response?.data?.message || '更新失败') }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确认删除用户「${row.id}」？`, '提示', { type:'warning', confirmButtonText:'删除', cancelButtonText:'取消' })
    .then(async () => {
      await axios.delete(`/api/v1/admin/users/${row.id}`)
      ElMessage.success('已删除')
      loadData()
    })
    .catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.user-list-page { max-width:1200px; margin:0 auto; }
.page-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:16px; }
.page-header h2 { margin:0; font-size:20px; }
.search-bar { background:#fff; border-radius:8px; padding:16px 20px; margin-bottom:16px; box-shadow:0 1px 3px rgba(0,0,0,.06); }
.search-row { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.table-wrap { background:#fff; border-radius:8px; box-shadow:0 1px 3px rgba(0,0,0,.06); overflow:hidden; }
.pagination-wrap { display:flex; justify-content:flex-end; margin-top:16px; background:#fff; border-radius:8px; padding:16px 20px; box-shadow:0 1px 3px rgba(0,0,0,.06); }
@media (max-width:767px) { .search-row { flex-direction:column; align-items:stretch; } }
</style>
