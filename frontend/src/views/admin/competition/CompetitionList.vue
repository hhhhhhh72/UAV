<template>
  <div class="competition-page">
    <div class="search-bar">
      <div class="search-row">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始" end-placeholder="结束"
          value-format="YYYY-MM-DD"
          style="width: 240px"
          @change="handleSearch"
        />
        <el-input v-model="searchText" placeholder="搜索姓名、单位或手机号" clearable style="width: 220px" @input="onFilterChange">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="selectedRole" clearable style="width: 130px" @change="onFilterChange">
          <el-option label="全部角色" value="all" />
          <el-option label="运动员" value="athlete" />
          <el-option label="教练员" value="coach" />
          <el-option label="裁判员" value="referee" />
          <el-option label="俱乐部" value="club" />
        </el-select>
        <el-select v-model="selectedStatus" clearable style="width: 120px" @change="onFilterChange">
          <el-option label="全部状态" value="all" />
          <el-option label="待处理" value="待处理" />
          <el-option label="处理中" value="处理中" />
          <el-option label="已完成" value="已完成" />
        </el-select>

        <div style="margin-left: auto; display: flex; gap: 8px;">
          <el-button type="warning" :icon="Download" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</el-button>
          <el-button type="success" :icon="Download" @click="handleExport">导出全部</el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <MetricCard label="总报名" :value="competitionStats.total" value-color="var(--accent-color, #0071e3)" />
      <MetricCard label="运动员" :value="competitionStats.athlete" value-color="#ff3b30" />
      <MetricCard label="教练员" :value="competitionStats.coach" value-color="#ff9f0a" />
      <MetricCard label="裁判员" :value="competitionStats.referee" value-color="#34c759" />
      <MetricCard label="俱乐部" :value="competitionStats.club" value-color="#5856d6" />
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="filteredList" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="regNo" label="注册号" width="130" />
        <el-table-column label="姓名/单位" min-width="140">
          <template #default="{ row }">
            <span class="cell-name">{{ row.name || row.companyName || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="90">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.competitionRole)" size="small">{{ row.competitionRoleText || row.competitionRole || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="联系电话" width="140">
          <template #default="{ row }">{{ row.phone || row.managerPhone || row.contactPhone || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status || '待处理' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="报名时间" width="160" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无报名数据" /></template>
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
    <el-dialog v-model="detailVisible" title="报名详情" width="600px">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="注册号">{{ currentItem.regNo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ currentItem.status || '待处理' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="currentItem.competitionRole === 'club' ? '负责人' : '姓名'">
            {{ currentItem.name || currentItem.manager || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="角色">{{ currentItem.competitionRoleText || currentItem.competitionRole || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.companyName" label="单位">{{ currentItem.companyName }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ currentItem.phone || currentItem.managerPhone || currentItem.contactPhone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="报名时间">{{ formatDate(currentItem.createTime) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.email" label="邮箱">{{ currentItem.email }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.level" label="等级">{{ currentItem.level }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.competitionProject" label="参赛项目" :span="2">{{ currentItem.competitionProject }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.remark" label="备注" :span="2">{{ currentItem.remark }}</el-descriptions-item>
        </el-descriptions>

        <div class="review-actions">
          <el-divider />
          <span style="margin-right: 12px;">修改状态：</span>
          <el-select v-model="newStatus" style="width: 140px;">
            <el-option v-for="s in statusOpts" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
          <el-button type="primary" @click="onUpdateStatus">更新</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Download } from '@element-plus/icons-vue'
import { showToast, showSuccessToast } from 'vant'
import { useListRequest } from '@/hooks/useListRequest'
import { getApplicationList, updateApplicationStatus, exportApplications } from '@/api/admin/application'
import { useAuth } from '../composables/useAuth'
import MetricCard from '../components/MetricCard.vue'

const { userRole } = useAuth()

const statusOpts = [
  { label: '待处理', value: '待处理' },
  { label: '处理中', value: '处理中' },
  { label: '已完成', value: '已完成' }
]
const statusTagType = (s) => ({ '已完成': 'success', '处理中': '', '待处理': 'warning' }[s] || 'info')
const roleTagType = (r) => ({ athlete: '', coach: 'warning', referee: 'success', club: 'danger' }[r] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const dateRange = ref(null)
const searchText = ref('')
const selectedRole = ref('all')
const selectedStatus = ref('all')

const { listData: allList, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSelectChange } = useListRequest({
  apiFunction: getApplicationList,
  idKey: 'id',
  defaultParams: { role: String(userRole.value || 'admin') }
})

// 自定义日期+搜索
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

const onFilterChange = () => {} // 前端筛选，不请求

// 前端筛选
const filteredList = computed(() => {
  let list = (allList.value || []).filter(i => i.serviceId === '13' || i.competitionRole)
  if (selectedRole.value !== 'all') list = list.filter(i => i.competitionRole === selectedRole.value)
  if (selectedStatus.value !== 'all') list = list.filter(i => i.status === selectedStatus.value)
  const kw = searchText.value.toLowerCase().trim()
  if (kw) {
    list = list.filter(i =>
      (i.name && i.name.toLowerCase().includes(kw)) ||
      (i.companyName && i.companyName.toLowerCase().includes(kw)) ||
      (i.phone && i.phone.includes(kw)) ||
      (i.managerPhone && i.managerPhone.includes(kw)) ||
      (i.contactPhone && i.contactPhone.includes(kw)) ||
      (i.regNo && i.regNo.toLowerCase().includes(kw))
    )
  }
  return list
})

// 统计
const competitionStats = computed(() => {
  const stats = { total: 0, athlete: 0, coach: 0, referee: 0, club: 0 }
  ;(allList.value || []).filter(i => i.serviceId === '13' || i.competitionRole).forEach(i => {
    stats.total++
    if (i.competitionRole === 'athlete') stats.athlete++
    else if (i.competitionRole === 'coach') stats.coach++
    else if (i.competitionRole === 'referee') stats.referee++
    else if (i.competitionRole === 'club') stats.club++
  })
  return stats
})

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
  const params = { role: userRole.value || 'admin', serviceId: 13 }
  if (dateRange.value?.[0]) { params.startDate = dateRange.value[0]; params.endDate = dateRange.value[1] }
  if (selectedRole.value !== 'all') params.competitionRole = selectedRole.value
  if (selectedStatus.value !== 'all') params.status = selectedStatus.value
  window.open(exportApplications(params), '_blank')
}

const handleSelectiveExport = () => {
  if (selectedIds.value.length === 0) return
  window.open(exportApplications({ ids: selectedIds.value.join(','), role: userRole.value || 'admin' }), '_blank')
}

onMounted(loadData)
</script>

<style scoped>
.competition-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.stats-row { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; margin-bottom: 16px; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-name { font-weight: 500; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { display: flex; align-items: center; justify-content: center; padding-top: 16px; gap: 8px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .stats-row { grid-template-columns: repeat(3, 1fr); } .table-wrap { overflow-x: auto; } }
</style>
