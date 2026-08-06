<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索企业或品牌名称" allow-clear style="width: 240px" @press-enter="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部状态</a-option>
              <a-option value="draft">草稿</a-option>
              <a-option value="published">已发布</a-option>
              <a-option value="rejected">已驳回</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增品牌</a-button>
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
        @page-change="loadData"
      >
        <template #name="{ record }">
          <span class="cell-title">{{ record.name || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button v-if="record.status === 'pending'" type="text" status="success" size="small" @click="handleApprove(record)">通过</a-button>
            <a-button v-if="record.status === 'pending'" type="text" status="danger" size="small" @click="handleReject(record)">驳回</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无品牌数据" />
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
    <a-modal v-model:visible="detailVisible" title="品牌详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="企业名称">{{ currentItem.company || '-' }}</a-descriptions-item>
          <a-descriptions-item label="品牌名称">{{ currentItem.brand_name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="行业">{{ currentItem.industry || '-' }}</a-descriptions-item>
          <a-descriptions-item label="展示类型">{{ currentItem.portfolio_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="审核状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Logo" :span="2">{{ currentItem.logo || '-' }}</a-descriptions-item>
          <a-descriptions-item label="封面图" :span="2">{{ currentItem.cover_image || '-' }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="荣誉" :span="2">{{ currentItem.honors || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑品牌' : '新增品牌'" :width="560" destroy-on-close>
      <a-form :model="form" layout="vertical">
        <a-form-item label="品牌名称" required><a-input v-model="form.name" /></a-form-item>
        <a-form-item label="Logo URL"><a-input v-model="form.logo_url" /></a-form-item>
        <a-form-item label="封面图 URL"><a-input v-model="form.cover_url" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 2 }" /></a-form-item>
        <a-form-item label="荣誉"><a-input v-model="form.honorsText" type="textarea" :autosize="{ minRows: 2 }" placeholder="多个荣誉用逗号分隔" /></a-form-item>
        <a-form-item label="审核状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="published">已发布</a-option>
            <a-option value="rejected">已驳回</a-option>
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

const api = useAdminApi('portfolios')

const statusLabel = (s) => ({ pending: '待审核', approved: '已通过', rejected: '已驳回' }[s] || s || '-')
const statusTag = (s) => ({ pending: 'orangered', approved: 'green', rejected: 'red' }[s] || 'gray')

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { status: '' }
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
  { title: '品牌名称', dataIndex: 'name', slotName: 'name', minWidth: 160, ellipsis: true, tooltip: true },
  { title: '审核状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 220, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (r) => { currentItem.value = r; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', logo_url: '', cover_url: '', description: '', honorsText: '', status: 'draft' })
const resetForm = () => Object.assign(form, { id: '', name: '', logo_url: '', cover_url: '', description: '', honorsText: '', status: 'draft' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, { id: r.id, name: r.name || '', logo_url: r.logo_url || '', cover_url: r.cover_url || '', description: r.description || '', honorsText: Array.isArray(r.honors) ? r.honors.join('、') : (r.honors || ''), status: r.status || 'draft' }); formEdit.value = true; formVisible.value = true }
const submitForm = async () => {
  if (!form.name) { Message.warning('请输入品牌名称'); return }
  formLoading.value = true
  try {
    const p = { id: form.id, name: form.name, logo_url: form.logo_url, cover_url: form.cover_url, description: form.description, status: form.status, honors: String(form.honorsText || '').split(/[,，、]/).map(x => x.trim()).filter(Boolean) }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false; loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleApprove = async (r) => {
  try { await api.update(r.id, { status: 'published' }); Message.success('已发布'); loadData() }
  catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
}
const handleReject = async (r) => {
  try { await api.update(r.id, { status: 'rejected' }); Message.success('已驳回'); loadData() }
  catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除品牌',
    content: '确定删除该品牌吗？',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(r.id); Message.success('已删除'); loadData() }
      catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
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
