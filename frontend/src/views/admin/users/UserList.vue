<template>
  <div class="user-list-page">
    <!-- 搜索过滤区 -->
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索用户姓名..."
          clearable style="width: 220px"
          @keyup.enter="onSearchSubmit" @clear="onSearchSubmit"
        >
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

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="name" label="姓名" min-width="140">
          <template #default="{ row }"><span class="cell-name">{{ row.name || '-' }}</span></template>
        </el-table-column>

        <el-table-column prop="phone" label="手机号" width="140" />

        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="注册时间" width="170" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="isSuperAdmin && row.phone !== SUPER_ADMIN_PHONE"
              link type="primary" size="small"
              @click="toggleRole(row)"
            >
              {{ row.role === 'platform_admin' ? '取消管理员' : '设为管理员' }}
            </el-button>
            <span v-else-if="row.phone === SUPER_ADMIN_PHONE" style="color: var(--el-text-color-placeholder); font-size: 12px;">超级管理员</span>
          </template>
        </el-table-column>

        <template #empty><el-empty description="暂无用户数据" /></template>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total" layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData" @current-change="loadData"
      />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { showSuccessToast, showFailToast } from 'vant'
import { useListRequest } from '@/hooks/useListRequest'
import { getUserList, updateUserRole } from '@/api/admin/user'
import { useAuth } from '../composables/useAuth'

const { isSuperAdmin, SUPER_ADMIN_PHONE } = useAuth()

const roleLabel = (r) => ({
  platform_admin: '平台管理员', association_admin: '协会管理员',
  enterprise: '企业用户', individual: '个人用户'
}[r] || r || '-')

const roleTagType = (r) => ({
  platform_admin: 'success', association_admin: 'warning',
  enterprise: '', individual: 'info'
}[r] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: getUserList,
  idKey: 'id',
  defaultParams: { role: '' }
})

const toggleRole = async (user) => {
  const newRole = user.role === 'platform_admin' ? 'individual' : 'platform_admin'
  try {
    await updateUserRole(user.id, newRole)
    user.role = newRole
    showSuccessToast('权限已更新')
  } catch (e) { showFailToast(e?.response?.data?.message || '权限更新失败') }
}

onMounted(loadData)
</script>

<style scoped>
.user-list-page { max-width: 1200px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-name { font-weight: 500; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
