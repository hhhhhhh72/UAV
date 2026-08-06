<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索院校名称..." allow-clear style="width: 220px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 160px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="active">合作中</a-option>
              <a-option value="pending">待合作</a-option>
              <a-option value="closed">已终止</a-option>
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
        <template #name="{ record }">
          <span class="cell-title">{{ record.name || '-' }}</span>
        </template>
        <template #coopType="{ record }">
          <span>{{ coopTypeLabel[record.coop_type] || coopTypeLabel.both }}</span>
        </template>
        <template #majors="{ record }">
          <span>{{ arrText(record.majors) }}</span>
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
          <a-empty description="暂无院校数据" />
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
    <a-modal v-model:visible="detailVisible" title="院校详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="院校名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="分域">{{ coopTypeLabel[currentItem.coop_type] || coopTypeLabel.both }}</a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.region || '-' }}</a-descriptions-item>
          <a-descriptions-item label="合作状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="特色专业" :span="2">{{ arrText(currentItem.majors) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="实训设施" :span="2">{{ arrText(currentItem.facilities) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="简介" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑院校' : '新增院校'" :width="560" @ok="submitForm" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="14">
            <a-form-item label="院校名称" required><a-input v-model="form.name" /></a-form-item>
          </a-col>
          <a-col :span="10">
            <a-form-item label="分域">
              <a-select v-model="form.coop_type" style="width: 100%">
                <a-option value="research">科研合作</a-option>
                <a-option value="talent">人才培养</a-option>
                <a-option value="both">综合</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="地区"><a-input v-model="form.region" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="合作状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="active">合作中</a-option>
                <a-option value="pending">洽谈中</a-option>
                <a-option value="closed">已结束</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="特色专业"><a-input v-model="form.majorsText" placeholder="多个专业用逗号分隔，如：无人机应用技术,测绘地理信息" /></a-form-item>
        <a-form-item label="实训设施"><a-input v-model="form.facilitiesText" placeholder="多个设施用逗号分隔，如：实训基地,联合实验室" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="2" /></a-form-item>
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

const api = useAdminApi('colleges')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ 'active': 'green', 'pending': 'orangered', 'closed': 'gray' }[s] || 'gray')
const statusLabel = { active:'合作中', pending:'待合作', closed:'已终止' }

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
  { title: '院校名称', dataIndex: 'name', slotName: 'name', minWidth: 200 },
  { title: '分域', dataIndex: 'coop_type', slotName: 'coopType', width: 100 },
  { title: '地区', dataIndex: 'region', width: 120 },
  { title: '特色专业', dataIndex: 'majors', slotName: 'majors', minWidth: 160 },
  { title: '合作状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false); const formEdit = ref(false); const formLoading = ref(false)
const coopTypeLabel = { research:'科研合作', talent:'人才培养', both:'综合' }
const arrText = (v) => (Array.isArray(v) ? v.join('、') : (v || ''))
const splitArr = (s) => String(s || '').split(/[,，、]/).map(x=>x.trim()).filter(Boolean)
const form = reactive({ id:'', name:'', coop_type:'both', region:'', majorsText:'', facilitiesText:'', status:'pending', description:'' })
const resetForm = () => Object.assign(form, { id:'', name:'', coop_type:'both', region:'', majorsText:'', facilitiesText:'', status:'pending', description:'' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => {
  Object.assign(form, {
    id: r.id, name: r.name || '', coop_type: r.coop_type || 'both', region: r.region || '',
    majorsText: arrText(r.majors), facilitiesText: arrText(r.facilities),
    status: r.status || 'pending', description: r.description || ''
  })
  formEdit.value = true
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.name) { Message.warning('请输入院校名称'); return }
  formLoading.value = true
  try {
    const p = {
      id: form.id, name: form.name, coop_type: form.coop_type, region: form.region,
      status: form.status, description: form.description,
      majors: splitArr(form.majorsText), facilities: splitArr(form.facilitiesText)
    }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除院校',
    content: '确定删除该院校吗？删除后不可恢复',
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
