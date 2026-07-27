<template>
  <div class="order-list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
          @change="onSearchSubmit"
        />
        <el-select v-model="filterParams.status" clearable style="width: 130px" @change="onSearchSubmit">
          <el-option label="全部状态" value="" />
          <el-option label="待处理" value="待处理" />
          <el-option label="处理中" value="处理中" />
          <el-option label="已完成" value="已完成" />
          <el-option label="已取消" value="已取消" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>

        <div style="margin-left: auto; display: flex; gap: 8px;">
          <el-button type="warning" :icon="Download" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</el-button>
          <el-button type="success" :icon="Download" @click="handleExport">导出全部</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="申请单号" width="160" sortable="custom" />
        <el-table-column prop="serviceName" label="服务名称" min-width="140" />
        <el-table-column prop="contactName" label="联系人" width="100" />
        <el-table-column prop="contactPhone" label="联系电话" width="130" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status || '待处理' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="申请时间" width="160" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total" layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData" @current-change="loadData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="申请单号">{{ currentItem.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ currentItem.status || '待处理' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="服务类型">{{ currentItem.serviceName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ currentItem.contactName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentItem.contactPhone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ formatDate(currentItem.createTime) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.remark" label="备注" :span="2">{{ currentItem.remark }}</el-descriptions-item>
        </el-descriptions>

        <div class="review-actions">
          <el-divider />
          <span style="margin-right: 12px; color: var(--el-text-color-regular);">修改状态：</span>
          <el-select v-model="newStatus" style="width: 140px;">
            <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
          <el-button type="primary" @click="onUpdateStatus">更新</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search, Download } from '@element-plus/icons-vue'
import { showToast, showSuccessToast } from 'vant'
import { useListRequest } from '@/hooks/useListRequest'
import { getApplicationList, updateApplicationStatus, exportApplications } from '@/api/admin/application'
import { useAuth } from '../composables/useAuth'

const { userRole } = useAuth()

const statusOptions = [
  { label: '待处理', value: '待处理' },
  { label: '处理中', value: '处理中' },
  { label: '已完成', value: '已完成' },
  { label: '已取消', value: '已取消' }
]

const statusTagType = (s) => ({ '已完成': 'success', '处理中': '', '待处理': 'warning', '已取消': 'info' }[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const dateRange = ref(null)

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({
  apiFunction: getApplicationList,
  idKey: 'id',
  defaultParams: { role: String(userRole.value || 'admin'), status: '' }
})

// 自定义日期范围逻辑
const origOnSearchSubmit = onSearchSubmit
const handleSearch = () => {
  if (dateRange.value && dateRange.value.length === 2) {
    filterParams.startDate = dateRange.value[0]
    filterParams.endDate = dateRange.value[1]
  } else {
    delete filterParams.startDate
    delete filterParams.endDate
  }
  origOnSearchSubmit()
}

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('待处理')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || '待处理'
  detailVisible.value = true
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    await updateApplicationStatus(currentItem.value.id, newStatus.value)
    currentItem.value.status = newStatus.value
    showSuccessToast('状态已更新')
    loadData()
  } catch (e) { showToast('更新失败') }
}

const handleExport = () => {
  const params = { role: userRole.value || 'admin' }
  if (dateRange.value?.[0]) { params.startDate = dateRange.value[0]; params.endDate = dateRange.value[1] }
  window.open(exportApplications(params), '_blank')
}

const handleSelectiveExport = () => {
  if (selectedIds.value.length === 0) return
  window.open(exportApplications({ ids: selectedIds.value.join(','), role: userRole.value || 'admin' }), '_blank')
}

onMounted(loadData)
</script>

<style scoped>
.order-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
