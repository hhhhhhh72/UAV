<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索成果名称" allow-clear style="width: 220px" @press-enter="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="领域" class="form-item">
            <a-select v-model="filterParams.field" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="无人机平台">无人机平台</a-option>
              <a-option value="飞控系统">飞控系统</a-option>
              <a-option value="导航与定位">导航与定位</a-option>
              <a-option value="通信链路">通信链路</a-option>
              <a-option value="载荷与传感器">载荷与传感器</a-option>
              <a-option value="能源动力">能源动力</a-option>
              <a-option value="人工智能">人工智能</a-option>
              <a-option value="新材料">新材料</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增</a-button>
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
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #field="{ record }">
          <a-tag size="small">{{ record.field || '-' }}</a-tag>
        </template>
        <template #stage="{ record }">
          <a-tag :color="stageTag(record.stage)" size="small">{{ stageLabel(record.stage) }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无成果数据" />
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
    <a-modal v-model:visible="detailVisible" title="成果详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="成果名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="领域">{{ currentItem.field || '-' }}</a-descriptions-item>
          <a-descriptions-item label="成果类型">{{ currentItem.achieve_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="所处阶段">
            <a-tag :color="stageTag(currentItem.stage)" size="small">{{ stageLabel(currentItem.stage) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="成果描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item label="附件" :span="2">{{ (currentItem.attachments || []).length }} 份</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑成果' : '新增成果'" :width="560" @close="resetForm">
      <a-form :model="form" layout="vertical">
        <a-form-item label="成果名称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" /></a-form-item>
        <a-form-item label="成果类型"><a-input v-model="form.achieve_type" placeholder="如：专利 / 样机 / 技术方案" /></a-form-item>
        <a-form-item label="阶段">
          <a-select v-model="form.stage" style="width: 100%">
            <a-option value="lab">实验室</a-option>
            <a-option value="pilot">中试</a-option>
            <a-option value="industrialization">产业化</a-option>
            <a-option value="launched">上市</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="附件资料">
          <div style="width: 100%">
            <div v-for="(at, i) in form.attachments" :key="i" style="display: flex; gap: 6px; margin-bottom: 6px">
              <a-input v-model="at.name" placeholder="附件名" style="width: 40%" />
              <a-input v-model="at.size" placeholder="大小" style="width: 20%" />
              <a-input v-model="at.url" placeholder="/uploads/xxx.pdf" style="flex: 1" />
              <a-button type="text" status="danger" size="small" @click="form.attachments.splice(i, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <a-button type="outline" size="small" @click="form.attachments.push({ name: '', size: '', url: '' })">+ 添加附件</a-button>
          </div>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('achievements')

const stageLabel = (s) => ({ lab: '实验室', pilot: '中试', industrialization: '产业化', launched: '上市' }[s] || s || '-')
const stageTag = (s) => ({ lab: 'gray', pilot: 'orangered', industrialization: 'green', launched: 'arcoblue' }[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { field: '' }
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
  { title: '成果名称', dataIndex: 'title', slotName: 'title', minWidth: 180 },
  { title: '领域', dataIndex: 'field', slotName: 'field', width: 120 },
  { title: '阶段', dataIndex: 'stage', slotName: 'stage', width: 100 },
  { title: '成果类型', dataIndex: 'achieve_type', width: 120 },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 170 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', field: '', stage: 'lab', achieve_type: '', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', field: '', stage: 'lab', achieve_type: '', description: '', attachments: [] })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, r); formEdit.value = true; formVisible.value = true }
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入成果名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false; loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除成果',
    content: `确定删除成果"${r.title}"吗？`,
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

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
