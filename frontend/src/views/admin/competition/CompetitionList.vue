<template>
  <div class="competition-page">
    <!-- Toolbar: date + search + filters -->
    <DataToolbar>
      <template #filters>
        <van-cell
          title="日期"
          :value="dateRangeText"
          is-link
          @click="showCalendar = true"
          style="flex: 0 0 auto; padding: 4px 0; background: transparent;"
        />
      </template>
      <template #actions>
        <van-button type="primary" size="small" @click="fetchData">查询</van-button>
        <van-button type="default" size="small" @click="handleExport">导出Excel</van-button>
      </template>
    </DataToolbar>

    <van-calendar
      :show="showCalendar"
      @update:show="v => showCalendar = v"
      type="range"
      @confirm="onConfirmDate"
      :min-date="new Date(2024, 0, 1)"
    />

    <!-- Search + filters -->
    <div class="filter-card">
      <van-search v-model="searchText" placeholder="搜索姓名、单位或手机号" @search="onFilterChange" @clear="onFilterChange" />
      <div style="padding: 0 12px 10px;">
        <van-dropdown-menu>
          <van-dropdown-item v-model="selectedRole" :options="roleFilterOptions" @change="onFilterChange" />
          <van-dropdown-item v-model="selectedStatus" :options="statusFilterOptions" @change="onFilterChange" />
        </van-dropdown-menu>
      </div>
    </div>

    <!-- Stats cards -->
    <div class="stats-grid">
      <MetricCard label="总报名" :value="competitionStats.total" value-color="var(--accent-color)" />
      <MetricCard label="运动员" :value="competitionStats.athlete" value-color="#ff3b30" />
      <MetricCard label="教练员" :value="competitionStats.coach" value-color="#ff9f0a" />
      <MetricCard label="裁判员" :value="competitionStats.referee" value-color="#34c759" />
      <MetricCard label="俱乐部" :value="competitionStats.club" value-color="#5856d6" />
    </div>

    <!-- Selection bar -->
    <div v-if="competitionList.length > 0" class="selection-bar">
      <van-checkbox v-model="allSelected" @click="toggleSelectAll">全选 ({{ selectedIds.length }})</van-checkbox>
      <van-button type="default" size="mini" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</van-button>
    </div>

    <!-- List -->
    <van-cell-group inset style="border-radius: var(--card-radius);">
      <van-cell
        v-for="item in competitionList"
        :key="item.id"
        :label="formatDate(item.createTime)"
        is-link
        @click="showDetail(item)"
      >
        <template #title>
          <div style="display: flex; align-items: center;">
            <van-checkbox v-model="item.selected" @click.stop style="margin-right: 8px;" />
            <span>{{ (item.name || item.companyName || '未知姓名') + ' - ' + (item.competitionRoleText || '未知角色') }}</span>
          </div>
        </template>
        <template #value>
          <span :class="statusClass(item.status)">{{ item.status || '待处理' }}</span>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- Detail Popup -->
    <van-popup :show="showDetailPopup" @update:show="v => showDetailPopup = v" position="bottom" :style="{ height: '70%' }" round>
      <div class="detail-content" v-if="currentItem">
        <van-cell-group title="基本信息">
          <van-cell title="申请单号" :value="v(currentItem.id)" />
          <van-cell title="申请时间" :value="formatDate(currentItem.createTime)" />
          <van-cell title="当前状态" :value="currentItem.status" is-link @click="showStatusPicker = true" />
          <van-cell v-if="currentItem.name || currentItem.manager" :title="currentItem.competitionRole === 'club' ? '负责人' : '姓名'" :value="v(currentItem.name || currentItem.manager)" />
          <van-cell v-if="currentItem.companyName" title="单位名称" :value="v(currentItem.companyName)" />
          <van-cell title="联系电话" :value="v(currentItem.phone || currentItem.managerPhone || currentItem.contactPhone)" />
        </van-cell-group>

        <van-cell-group title="赛事详情">
          <van-cell title="服务类型" :value="v(currentItem.serviceName)" />
          <van-cell title="注册号" :value="v(currentItem.regNo)" />
          <van-cell title="报名角色" :value="v(currentItem.competitionRoleText)" />
          <template v-if="currentItem.competitionRole === 'club'">
            <van-cell title="单位简称" :value="v(currentItem.companyShortName)" />
            <van-cell title="所在地" :value="v(currentItem.location)" />
            <van-cell title="负责人" :value="v(currentItem.manager)" />
            <van-cell title="负责人电话" :value="v(currentItem.managerPhone)" />
            <van-cell title="主要对接人" :value="v(currentItem.contactPerson)" />
            <van-cell title="对接人电话" :value="v(currentItem.contactPhone)" />
          </template>
          <template v-else>
            <van-cell title="性别" :value="currentItem.gender === 'male' ? '男' : '女'" />
            <van-cell title="证件号" :value="v(currentItem.idCard)" />
            <van-cell title="组别" :value="v(currentItem.competitionGroup || currentItem.athleteGroup)" />
            <van-cell v-if="currentItem.competitionProject" title="参赛项目" :value="v(currentItem.competitionProject)" />
            <van-cell :title="currentItem.competitionRole === 'referee' ? '裁判员等级' : (currentItem.competitionRole === 'coach' ? '教练员等级' : '等级')" :value="v(currentItem.level)" />
            <van-cell v-if="currentItem.validDate" title="有效期" :value="v(currentItem.validDate)" />
            <van-cell title="电子邮箱" :value="v(currentItem.email)" />
          </template>
          <van-cell title="备注" :label="v(currentItem.remark)" />
        </van-cell-group>
      </div>
    </van-popup>

    <!-- Status Picker -->
    <van-popup :show="showStatusPicker" @update:show="v => showStatusPicker = v" position="bottom">
      <van-picker :columns="statusOptions" @confirm="onUpdateStatus" @cancel="showStatusPicker = false" title="修改订单状态" />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast } from 'vant'
import DataToolbar from '../components/DataToolbar.vue'
import MetricCard from '../components/MetricCard.vue'
import { useAuth } from '../composables/useAuth'

const { userRole } = useAuth()

const list = ref([])
const showCalendar = ref(false)
const startDate = ref('')
const endDate = ref('')
const showDetailPopup = ref(false)
const showStatusPicker = ref(false)
const currentItem = ref(null)
const searchText = ref('')
const selectedRole = ref('all')
const selectedStatus = ref('all')

const v = val => val || '-'

const statusOptions = [
  { text: '待处理', value: '待处理' },
  { text: '处理中', value: '处理中' },
  { text: '已完成', value: '已完成' },
  { text: '已取消', value: '已取消' }
]

const roleFilterOptions = [
  { text: '全部角色', value: 'all' },
  { text: '运动员', value: 'athlete' },
  { text: '教练员', value: 'coach' },
  { text: '裁判员', value: 'referee' },
  { text: '俱乐部', value: 'club' }
]

const statusFilterOptions = [
  { text: '全部状态', value: 'all' },
  { text: '待处理', value: '待处理' },
  { text: '处理中', value: '处理中' },
  { text: '已完成', value: '已完成' },
  { text: '已取消', value: '已取消' }
]

const dateRangeText = computed(() => {
  if (startDate.value && endDate.value) {
    const s = new Date(startDate.value)
    const e = new Date(endDate.value)
    return `${s.getMonth() + 1}/${s.getDate()} - ${e.getMonth() + 1}/${e.getDate()}`
  }
  return '全部'
})

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const statusClass = (status) => status === '已完成' ? 'text-success' : 'text-warning'

const onFilterChange = () => {}

const competitionList = computed(() => {
  const listData = Array.isArray(list.value) ? list.value : []
  return listData.filter(item => {
    const isCompetition = item.serviceId === '13'
    const matchesRole = selectedRole.value === 'all' || item.competitionRole === selectedRole.value
    const matchesStatus = selectedStatus.value === 'all' || item.status === selectedStatus.value
    const searchLower = searchText.value.toLowerCase().trim()
    const matchesSearch = !searchLower ||
      (item.name && item.name.toLowerCase().includes(searchLower)) ||
      (item.companyName && item.companyName.toLowerCase().includes(searchLower)) ||
      (item.phone && item.phone.includes(searchLower)) ||
      (item.managerPhone && item.managerPhone.includes(searchLower)) ||
      (item.contactPhone && item.contactPhone.includes(searchLower)) ||
      (item.regNo && item.regNo.toLowerCase().includes(searchLower))
    return isCompetition && matchesRole && matchesStatus && matchesSearch
  })
})

const competitionStats = computed(() => {
  const stats = { total: 0, athlete: 0, coach: 0, referee: 0, club: 0 }
  const listData = Array.isArray(list.value) ? list.value : []
  listData.filter(item => item.serviceId === '13').forEach(item => {
    stats.total++
    if (item.competitionRole === 'athlete') stats.athlete++
    else if (item.competitionRole === 'coach') stats.coach++
    else if (item.competitionRole === 'referee') stats.referee++
    else if (item.competitionRole === 'club') stats.club++
  })
  return stats
})

// Selection logic
const allSelected = ref(false)
const selectedIds = computed(() => competitionList.value.filter(i => i.selected).map(i => i.id))
const toggleSelectAll = () => { competitionList.value.forEach(i => { i.selected = allSelected.value }) }

const onConfirmDate = (values) => {
  const [start, end] = values
  showCalendar.value = false
  startDate.value = start
  endDate.value = end
  fetchData()
}

const fetchData = async () => {
  try {
    const params = { role: userRole.value }
    if (startDate.value) params.startDate = startDate.value
    if (endDate.value) params.endDate = endDate.value
    const res = await axios.get('/api/list', { params })
    let data = res.data
    if (typeof data === 'string') { try { data = JSON.parse(data) } catch (e) { data = [] } }
    if (!Array.isArray(data)) data = []
    list.value = data.map(item => ({ ...item, selected: false }))
  } catch (error) {
    showFailToast('获取数据失败')
    console.error(error)
  }
}

const showDetail = (item) => {
  currentItem.value = { ...item }
  showDetailPopup.value = true
}

const onUpdateStatus = async ({ selectedOptions }) => {
  const newStatus = selectedOptions[0].value
  if (!currentItem.value) return
  try {
    await axios.post('/api/update', { id: currentItem.value.id, status: newStatus })
    currentItem.value.status = newStatus
    const index = list.value.findIndex(i => i.id === currentItem.value.id)
    if (index !== -1) list.value[index].status = newStatus
    showSuccessToast('状态更新成功')
    showStatusPicker.value = false
  } catch (error) {
    showFailToast('更新状态失败')
    console.error(error)
  }
}

const handleExport = () => {
  let url = `/api/export?role=${userRole.value}&serviceId=13&`
  if (startDate.value) url += `startDate=${startDate.value.toISOString()}&`
  if (endDate.value) url += `endDate=${endDate.value.toISOString()}&`
  if (selectedRole.value !== 'all') url += `competitionRole=${selectedRole.value}&`
  if (selectedStatus.value !== 'all') url += `status=${selectedStatus.value}`
  window.open(url, '_blank')
}

const handleSelectiveExport = () => {
  if (selectedIds.value.length === 0) return
  window.open(`/api/export?ids=${selectedIds.value.join(',')}&role=${userRole.value}`, '_blank')
}

onMounted(fetchData)
</script>

<style scoped>
.filter-card {
  background: var(--bg-primary, #fff);
  border-radius: var(--card-radius, 12px);
  margin-bottom: 14px;
  box-shadow: var(--card-shadow);
  overflow: hidden;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
  margin-bottom: 14px;
}

.selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: var(--bg-primary, #fff);
  border-radius: var(--card-radius) var(--card-radius) 0 0;
  box-shadow: var(--card-shadow);
}

.detail-content { padding: 16px 0; }
.text-success { color: var(--success-color, #34c759); }
.text-warning { color: var(--warning-color, #ff9f0a); }

@media (max-width: 767px) {
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
