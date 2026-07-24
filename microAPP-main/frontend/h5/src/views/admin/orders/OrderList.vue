<template>
  <div class="order-list-page">
    <!-- Toolbar: date filter + actions -->
    <DataToolbar>
      <template #filters>
        <van-cell
          title="日期范围"
          :value="dateRangeText"
          is-link
          @click="showCalendar = true"
          style="flex: 1; padding: 4px 0; background: transparent;"
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

    <!-- Selection bar -->
    <div v-if="list.length > 0" class="selection-bar">
      <van-checkbox v-model="allSelected" @click="toggleSelectAll">全选 ({{ selectedIds.length }})</van-checkbox>
      <van-button type="default" size="mini" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</van-button>
    </div>

    <!-- List -->
    <van-empty v-if="list.length === 0" description="暂无数据" />
    <van-cell-group v-else inset style="border-radius: var(--card-radius);">
      <van-cell
        v-for="item in list"
        :key="item.id"
        :label="formatDate(item.createTime)"
        is-link
        @click="showDetail(item)"
      >
        <template #title>
          <div style="display: flex; align-items: center;">
            <van-checkbox v-model="item.selected" @click.stop style="margin-right: 8px;" />
            <span>{{ item.serviceName || '未知服务' }}</span>
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
          <template v-if="currentItem.serviceId === '6'">
            <van-cell title="姓名" :value="v(currentItem.traineeName)" />
            <van-cell title="联系电话" :value="v(currentItem.traineePhone)" />
          </template>
          <template v-else-if="currentItem.serviceId === '9'">
            <van-cell title="学校/机构" :value="v(currentItem.studyOrg)" />
            <van-cell v-if="currentItem.studyGrade" title="年级/年龄段" :value="v(currentItem.studyGrade)" />
            <van-cell title="参与人数" :value="v(currentItem.studyParticipants)" />
            <van-cell title="期望日期" :value="v(currentItem.studyDate)" />
            <van-cell title="场次" :value="v(currentItem.studySessionText || currentItem.studySession)" />
          </template>
          <template v-else-if="currentItem.serviceId === '13'">
            <van-cell v-if="currentItem.name || currentItem.manager" :title="currentItem.competitionRole === 'club' ? '负责人' : '姓名'" :value="v(currentItem.name || currentItem.manager)" />
            <van-cell v-if="currentItem.companyName" title="单位名称" :value="v(currentItem.companyName)" />
            <van-cell title="联系电话" :value="v(currentItem.phone || currentItem.managerPhone || currentItem.contactPhone)" />
          </template>
          <template v-else>
            <van-cell title="联系人" :value="v(currentItem.contactName)" />
            <van-cell title="联系电话" :value="v(currentItem.contactPhone)" />
          </template>
        </van-cell-group>

        <van-cell-group title="服务详情">
          <van-cell title="服务类型" :value="v(currentItem.serviceName)" />
          <!-- 无人机赛事 -->
          <template v-if="currentItem.serviceId === '13'">
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
          </template>
          <!-- 飞手培训 -->
          <template v-else-if="currentItem.serviceId === '6'">
            <van-cell title="性别" :value="currentItem.traineeGender === 'male' ? '男' : '女'" />
            <van-cell title="出生日期" :value="v(currentItem.traineeBirthday)" />
            <van-cell title="身份证号" :value="v(currentItem.traineeIdCard)" />
            <van-cell title="执照种类" :value="v(currentItem.examModel)" />
            <van-cell title="证照级别" :value="v(currentItem.licenseLevel)" />
            <van-cell title="有无基础" :value="currentItem.hasExperience === 'yes' ? '有' : '无'" />
          </template>
          <!-- 物流等其他 -->
          <template v-else>
            <van-cell title="客户类型" :value="currentItem.customerType === 'enterprise' ? '企业' : '个人'" />
            <van-cell v-if="currentItem.companyName" title="企业名称" :value="currentItem.companyName" />
            <van-cell v-if="currentItem.cargoType" title="货物类型" :value="currentItem.cargoType" />
            <van-cell v-if="currentItem.startAddress" title="起运地" :value="currentItem.startAddress" />
            <van-cell v-if="currentItem.endAddress" title="目的地" :value="currentItem.endAddress" />
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

const list = ref([])
const showCalendar = ref(false)
const startDate = ref('')
const endDate = ref('')
const showDetailPopup = ref(false)
const showStatusPicker = ref(false)
const currentItem = ref(null)

const v = val => val || '-'

const statusOptions = [
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
  return '全部时间'
})

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const statusClass = (status) => status === '已完成' ? 'text-success' : 'text-warning'

// Selection logic
const allSelected = ref(false)
const selectedIds = computed(() => list.value.filter(i => i.selected).map(i => i.id))
const toggleSelectAll = () => { list.value.forEach(i => { i.selected = allSelected.value }) }

const onConfirmDate = (values) => {
  const [start, end] = values
  showCalendar.value = false
  startDate.value = start
  endDate.value = end
  fetchData()
}

const fetchData = async () => {
  try {
    const params = { role: 'admin' }
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
  let url = `/api/export?role=admin&`
  if (startDate.value) url += `startDate=${startDate.value.toISOString()}&`
  if (endDate.value) url += `endDate=${endDate.value.toISOString()}&`
  window.open(url, '_blank')
}

const handleSelectiveExport = () => {
  if (selectedIds.value.length === 0) return
  window.open(`/api/export?ids=${selectedIds.value.join(',')}&role=admin`, '_blank')
}

onMounted(fetchData)
</script>

<style scoped>
.selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: var(--bg-primary, #fff);
  border-radius: var(--card-radius, 12px);
  margin-bottom: 10px;
  box-shadow: var(--card-shadow);
}
.detail-content { padding: 16px 0; }
.text-success { color: var(--success-color, #34c759); }
.text-warning { color: var(--warning-color, #ff9f0a); }
</style>
