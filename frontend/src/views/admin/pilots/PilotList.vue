<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索姓名..." clearable style="width: 200px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.status" clearable placeholder="审核状态" style="width: 140px" @change="onSearchSubmit">
          <el-option label="待审核" value="pending" />
          <el-option label="已认证" value="approved" />
          <el-option label="已驳回" value="rejected" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="real_name" label="姓名" min-width="100" />
        <el-table-column prop="id_card" label="身份证号" width="200" />
        <el-table-column label="证书" width="80" align="center">
          <template #default="{ row }">{{ (row.cert_ids || []).length }} 项</template>
        </el-table-column>
        <el-table-column prop="flight_hours" label="时长(h)" width="80" align="center" />
        <el-table-column prop="bio" label="擅长领域" min-width="140">
          <template #default="{ row }">{{ row.bio || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel[row.status] || row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button link type="success" size="small" @click="handleApprove(row)">通过</el-button>
              <el-button link type="danger" size="small" @click="handleReject(row)">驳回</el-button>
            </template>
            <template v-else>
              <el-button v-if="row.status === 'approved'" link type="danger" size="small" @click="handleReject(row)">撤销</el-button>
              <el-button v-if="row.status === 'rejected'" link type="success" size="small" @click="handleApprove(row)">恢复通过</el-button>
            </template>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无飞手申请" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from '@/utils/http'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('certified-pilots')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusLabel = { pending: '待审核', approved: '已认证', rejected: '已驳回' }
const statusTag = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' },
})

// 审核：通过 / 驳回（专用端点）
const setStatus = async (row, action, tip) => {
  try {
    await axios.post(`/api/v1/admin/certified-pilots/${row.id}/${action}`)
    ElMessage.success(tip)
    loadData()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
}
const handleApprove = (row) => setStatus(row, 'approve', '已通过，飞手进入公开名录')
const handleReject = (row) => {
  ElMessageBox.confirm(`确定驳回 ${row.real_name} 的飞手认证申请？`, '提示', { type: 'warning' })
    .then(() => setStatus(row, 'reject', '已驳回'))
    .catch(() => {})
}
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
</style>
