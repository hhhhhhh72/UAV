<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索资源名称..." allow-clear style="width: 220px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="类型" class="form-item">
            <a-select v-model="filterParams.res_type" style="width: 140px" @change="onSearchSubmit">
              <a-option value="">全部类型</a-option>
              <a-option value="drone">无人机</a-option>
              <a-option value="comm">通信</a-option>
              <a-option value="light">照明</a-option>
              <a-option value="transport">运输</a-option>
              <a-option value="other">其他</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增资源</a-button>
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
        <template #name="{ record }">
          <span class="cell-title">{{ record.name || '-' }}</span>
        </template>
        <template #type="{ record }">
          <a-tag :color="typeTag(record.res_type)" size="small">{{ typeLabel(record.res_type) }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无应急资源" />
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
    <a-modal v-model:visible="detailVisible" title="应急资源详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="资源名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">
            <a-tag :color="typeTag(currentItem.res_type)" size="small">{{ typeLabel(currentItem.res_type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="规格">{{ currentItem.specs || '-' }}</a-descriptions-item>
          <a-descriptions-item label="数量">{{ currentItem.quantity || 0 }}</a-descriptions-item>
          <a-descriptions-item label="位置">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系人">{{ currentItem.contact_info || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="备注" :span="2">{{ currentItem.notes || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑应急资源' : '新增应急资源'" :width="560" :mask-closable="false" :unmount-on-close="true" @cancel="formVisible = false">
      <a-form :model="form" layout="horizontal">
        <a-row :gutter="16">
          <a-col :span="14">
            <a-form-item label="资源名称" required><a-input v-model="form.name" /></a-form-item>
          </a-col>
          <a-col :span="10">
            <a-form-item label="类型">
              <a-select v-model="form.res_type" style="width: 100%">
                <a-option value="drone">无人机</a-option>
                <a-option value="comm">通信</a-option>
                <a-option value="light">照明</a-option>
                <a-option value="transport">运输</a-option>
                <a-option value="other">其他</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="规格"><a-input v-model="form.specs" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="数量"><a-input-number v-model="form.quantity" :min="0" hide-button style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="位置"><a-input v-model="form.location" /></a-form-item>
        <a-form-item label="联系人信息"><a-input v-model="form.contact_info" placeholder="姓名 / 电话，如：张工 13800138000" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="available">可用</a-option>
            <a-option value="in_use">使用中</a-option>
            <a-option value="maintenance">维护中</a-option>
          </a-select>
        </a-form-item>
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

const api = useAdminApi('emergency-resources')

const typeLabel = (t) => ({ drone: '无人机', comm: '通信', light: '照明', transport: '运输', other: '其他' }[t] || t || '-')
const typeTag = (t) => ({ drone: 'green', comm: 'orange', light: 'gray', transport: 'arcoblue', other: 'arcoblue' }[t] || 'gray')
const statusLabel = (s) => ({ available: '可用', in_use: '使用中', maintenance: '维护中' }[s] || s || '-')
const statusTag = (s) => ({ available: 'green', in_use: 'orange', maintenance: 'red' }[s] || 'gray')

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { res_type: '' }
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
  { title: '资源名称', dataIndex: 'name', slotName: 'name', minWidth: 160 },
  { title: '类型', dataIndex: 'res_type', slotName: 'type', width: 100 },
  { title: '规格', dataIndex: 'specs', width: 140 },
  { title: '数量', dataIndex: 'quantity', width: 70, align: 'center' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', res_type: 'drone', specs: '', quantity: 0, location: '', contact_info: '', status: 'available' })
const resetForm = () => Object.assign(form, { id: '', name: '', res_type: 'drone', specs: '', quantity: 0, location: '', contact_info: '', status: 'available' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, { ...r, quantity: r.quantity || 0 }); formEdit.value = true; formVisible.value = true }

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入资源名称'); return }
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
    content: '确定删除?',
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

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
