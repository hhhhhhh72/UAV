<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索活动名称..." allow-clear style="width: 220px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" @change="onSearchSubmit">
              <a-option value="">全部状态</a-option>
              <a-option value="upcoming">即将开始</a-option>
              <a-option value="ongoing">进行中</a-option>
              <a-option value="ended">已结束</a-option>
              <a-option value="cancelled">已取消</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增活动</a-button>
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
        :row-selection="rowSelection"
        @sorter-change="handleSortChange"
      >
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #type="{ record }">
          <a-tag :color="typeTag(record.event_type)" size="small">{{ record.event_type || '-' }}</a-tag>
        </template>
        <template #startTime="{ record }">
          <span class="time-text">{{ formatDate(record.start_time) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ record.status || '-' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无活动数据" />
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
    <a-modal v-model:visible="detailVisible" title="活动详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="活动名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">{{ currentItem.event_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="开始时间">{{ formatDate(currentItem.start_time) }}</a-descriptions-item>
          <a-descriptions-item label="结束时间">{{ formatDate(currentItem.end_time) }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.max_attendees || '-' }}</a-descriptions-item>
          <a-descriptions-item label="已报名">{{ currentItem.reg_count || 0 }}</a-descriptions-item>
          <a-descriptions-item label="已报名" :span="2">{{ currentItem.reg_count || 0 }} 人</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑活动' : '新增活动'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="horizontal">
        <a-row :gutter="16">
          <a-col :span="16">
            <a-form-item label="活动名称" required><a-input v-model="form.title" /></a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="类型">
              <a-select v-model="form.event_type" style="width: 100%">
                <a-option value="论坛">论坛</a-option>
                <a-option value="走访">走访</a-option>
                <a-option value="沙龙">沙龙</a-option>
                <a-option value="培训">培训</a-option>
                <a-option value="其他">其他</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="地点"><a-input v-model="form.location" /></a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始时间"><a-input v-model="form.start_time" placeholder="YYYY-MM-DD HH:mm" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束时间"><a-input v-model="form.end_time" placeholder="YYYY-MM-DD HH:mm" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="名额"><a-input-number v-model="form.max_attendees" :min="0" hide-button style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="published">已发布</a-option>
                <a-option value="ongoing">进行中</a-option>
                <a-option value="ended">已结束</a-option>
                <a-option value="cancelled">已取消</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="3" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('events')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const typeTag = (t) => ({ '论坛': 'green', '走访': 'arcoblue', '沙龙': 'orange', '培训': 'gray', '其他': 'arcoblue' }[t] || 'gray')
const statusTag = (s) => ({ published: 'orange', ongoing: 'green', ended: 'gray', cancelled: 'red' }[s] || 'gray')

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

// Arco sorter-change → useListRequest.onSortChange（el-table 的 { prop, order } 形态）
const handleSortChange = (dataIndex, direction) => {
  onSortChange({
    prop: dataIndex,
    order: direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  })
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '活动名称', dataIndex: 'title', slotName: 'title', minWidth: 180 },
  { title: '类型', dataIndex: 'event_type', slotName: 'type', width: 100 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '开始时间', dataIndex: 'start_time', slotName: 'startTime', width: 160, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '名额', dataIndex: 'max_attendees', width: 70, align: 'center' },
  { title: '已报名', dataIndex: 'reg_count', width: 80, align: 'center' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', event_type: '论坛', location: '', start_time: '', end_time: '', max_attendees: 0, status: 'published', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', event_type: '论坛', location: '', start_time: '', end_time: '', max_attendees: 0, status: 'published', description: '' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, { ...r, capacity: r.capacity || 0 }); formEdit.value = true; formVisible.value = true }

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入活动名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (r) => {
  Modal.confirm({
    title: '提示',
    content: '确定删除该活动?',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); loadData() } catch { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
