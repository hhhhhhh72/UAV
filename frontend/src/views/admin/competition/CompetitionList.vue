<template>
  <div class="page">
    <!-- 搜索 + 新建 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 130px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部状态</a-option>
              <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button type="success" @click="showCreate">新建赛事</a-button>
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
        <template #startDate="{ record }">
          <span class="time-text">{{ formatDate(record.start_date) }}</span>
        </template>
        <template #regCount="{ record }">
          <span>{{ record.reg_count || 0 }} / {{ record.max_teams || 0 }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTagType(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" status="danger" size="small" @click="onDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无赛事" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="赛事详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="赛事名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类别">{{ currentItem.category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="报名/名额">{{ currentItem.reg_count || 0 }} / {{ currentItem.max_teams || 0 }}</a-descriptions-item>
          <a-descriptions-item label="主办方">{{ currentItem.sponsor || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.description" label="简介" :span="2">{{ currentItem.description }}</a-descriptions-item>
        </a-descriptions>

        <div class="review-actions">
          <a-divider />
          <span style="margin-right: 12px;">修改状态：</span>
          <a-select v-model="newStatus" style="width: 140px;">
            <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</a-option>
          </a-select>
          <a-button type="primary" @click="onUpdateStatus">更新</a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { getCompetitionList, updateCompetition, deleteCompetition } from '@/api/admin/competition'

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '报名中', value: 'open' },
  { label: '已结束', value: 'closed' }
]
const statusLabel = (s) => statusOptions.find(o => o.value === s)?.label || s || '-'
const statusTagType = (s) => ({ open: 'green', closed: 'gray', draft: 'orangered' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: getCompetitionList,
  idKey: 'id',
  defaultParams: { status: '' }
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 180 },
  { title: '赛事名称', dataIndex: 'title', minWidth: 160 },
  { title: '类别', dataIndex: 'category', width: 100 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始时间', dataIndex: 'start_date', slotName: 'startDate', width: 170 },
  { title: '报名/名额', dataIndex: 'reg_count', slotName: 'regCount', width: 110 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const newStatus = ref('draft')

const showDetail = (item) => {
  currentItem.value = { ...item }
  newStatus.value = item.status || 'draft'
  detailVisible.value = true
}

const showCreate = () => {
  Message.info('请使用小程序/直接调用 POST /api/v1/admin/competitions 创建赛事')
}

const onUpdateStatus = async () => {
  if (!currentItem.value) return
  try {
    await updateCompetition(currentItem.value.id, { status: newStatus.value })
    currentItem.value.status = newStatus.value
    Message.success('状态已更新')
    loadData()
  } catch (e) { Message.error('更新失败') }
}

const onDelete = (row) => {
  Modal.confirm({
    title: '删除赛事',
    content: `确定删除「${row.title}」吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteCompetition(row.id)
        Message.success('已删除')
        loadData()
      } catch (e) { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.time-text { color: #86909C; font-size: 12px; }

.review-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 16px;
  gap: 8px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
