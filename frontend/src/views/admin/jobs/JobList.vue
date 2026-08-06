<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索职位标题..." allow-clear style="width: 220px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="published">招聘中</a-option>
              <a-option value="closed">已关闭</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="success" @click="handleAdd">新增</a-button>
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
      >
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #company="{ record }">
          <span>{{ record.company || '-' }}</span>
        </template>
        <template #salary="{ record }">
          <span>{{ record.salary_fen ? '¥' + (record.salary_fen / 100).toLocaleString('zh-CN') : '-' }}</span>
        </template>
        <template #jobType="{ record }">
          <a-tag :color="jobTypeTag(record.job_type)" size="small">{{ record.job_type || '-' }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无职位数据" />
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
    <a-modal v-model:visible="detailVisible" title="职位详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="职位名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="公司">{{ currentItem.company || '-' }}</a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="薪资">{{ currentItem.salary || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="jobTypeTag(currentItem.job_type)" size="small">{{ currentItem.job_type || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="发布时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="要求" :span="2">{{ currentItem.requirements || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑职位' : '新增职位'" :width="560" @ok="submitForm" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="14">
            <a-form-item label="职位名称" required><a-input v-model="form.title" /></a-form-item>
          </a-col>
          <a-col :span="10">
            <a-form-item label="类型">
              <a-select v-model="form.job_type" style="width: 100%">
                <a-option value="全职">全职</a-option>
                <a-option value="兼职">兼职</a-option>
                <a-option value="实习">实习</a-option>
                <a-option value="外包">外包</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="地区"><a-input v-model="form.location" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="published">招聘中</a-option>
                <a-option value="closed">已关闭</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="薪资(元)">
              <a-input-number v-model="form.salary" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="职位描述"><a-input v-model="form.description" type="textarea" :rows="2" /></a-form-item>
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

const api = useAdminApi('jobs')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const jobTypeTag = (t) => ({ '全职': 'arcoblue', '兼职': 'orangered', '实习': 'gray', 'contract': 'gray' }[t] || 'gray')

const statusTag = (s) => ({ 'active': 'green', 'closed': 'gray' }[s] || 'gray')
const statusLabel = { active:'招聘中', closed:'已关闭' }

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
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

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '职位名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '公司ID', dataIndex: 'enterprise_id', slotName: 'company', width: 160 },
  { title: '地区', dataIndex: 'location', width: 120 },
  { title: '薪资', dataIndex: 'salary_fen', slotName: 'salary', width: 130 },
  { title: '类型', dataIndex: 'job_type', slotName: 'jobType', width: 100 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false); const formEdit = ref(false); const formLoading = ref(false)
const form = reactive({ id:'', title:'', location:'', salary: 0, job_type:'全职', status:'published', description:'' })
const resetForm = () => Object.assign(form, { id:'', title:'', location:'', salary: 0, job_type:'全职', status:'published', description:'' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, r); formEdit.value = true; formVisible.value = true }
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入职位名称'); return }
  formLoading.value = true
  try {
    const p = {
      id: form.id, title: form.title, location: form.location, job_type: form.job_type,
      status: form.status, description: form.description,
      salary_fen: Math.round((Number(form.salary) || 0) * 100)
    }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除职位',
    content: '确定删除该职位吗？删除后不可恢复',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(r.id)
        Message.success('已删除')
        loadData()
      } catch { Message.error('删除失败') }
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

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
