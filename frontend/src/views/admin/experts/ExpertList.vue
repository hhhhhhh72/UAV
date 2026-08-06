<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="searchForm" class="search-form">
        <a-space wrap>
          <a-form-item label="领域" class="form-item">
            <a-select v-model="searchForm.field" style="width: 160px" allow-clear @change="handleSearch">
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
          <a-button type="primary" @click="handleSearch"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="handleReset">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增专家</a-button>
        </a-space>
      </a-form>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="tableData"
        :loading="loading"
        row-key="id"
        :pagination="false"
        @page-change="loadData"
      >
        <template #tags="{ record }">
          <a-tag v-for="t in record.tags" :key="t" size="small" style="margin: 2px">{{ t }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusMap[record.status] || 'gray'" size="small">{{ statusLabel[record.status] || record.status }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无专家数据" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="pagination.total > 0">
        <a-pagination
          v-model:current="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="dialog.visible" :title="dialog.isEdit ? '编辑专家' : '新增专家'" :width="500" destroy-on-close>
      <a-form :model="form" layout="vertical">
        <a-form-item label="姓名" required><a-input v-model="form.name" /></a-form-item>
        <a-form-item label="职称"><a-input v-model="form.title" /></a-form-item>
        <a-form-item label="单位"><a-input v-model="form.org" /></a-form-item>
        <a-form-item label="领域"><a-input v-model="form.field" /></a-form-item>
        <a-form-item label="简介"><a-input v-model="form.bio" type="textarea" :autosize="{ minRows: 3 }" /></a-form-item>
        <a-form-item label="头像">
          <a-upload
            class="avatar-upload"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            accept="image/*"
            :before-upload="beforeUpload"
            @success="onUploadSuccess"
          >
            <a-avatar v-if="form.avatar_url" :src="form.avatar_url" :size="80" shape="square" />
            <a-button v-else type="outline">点击上传</a-button>
          </a-upload>
          <a-button v-if="form.avatar_url" size="small" style="margin-top: 8px" @click="form.avatar_url = ''">清除</a-button>
        </a-form-item>
        <a-form-item label="标签"><a-input v-model="tagsInput" placeholder="逗号分隔" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="pending">待审核</a-option>
            <a-option value="published">已发布</a-option>
            <a-option value="archived">已下架</a-option>
          </a-select>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="dialog.visible = false">取消</a-button>
        <a-button type="primary" :loading="dialog.loading" @click="handleSubmit">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'

const loading = ref(false)
const tableData = ref([])
const searchForm = ref({})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const statusMap = { pending: 'orangered', published: 'green', archived: 'gray' }
const statusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }
const tagsInput = ref('')
const dialog = reactive({ visible: false, isEdit: false, loading: false })
const form = reactive({ id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }

const columns = [
  { title: 'ID', dataIndex: 'id', width: 140 },
  { title: '姓名', dataIndex: 'name', width: 100 },
  { title: '职称', dataIndex: 'title', width: 120 },
  { title: '所属单位', dataIndex: 'org', minWidth: 180 },
  { title: '领域', dataIndex: 'field', width: 120 },
  { title: '标签', slotName: 'tags', width: 200 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const beforeUpload = (item) => {
  const file = item?.file || item
  const isImage = !!file.type && file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

const onUploadSuccess = (fileItem) => {
  const res = fileItem?.response || {}
  form.avatar_url = res?.data?.url || res?.url || ''
  Message.success('上传成功')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/experts', { params: { ...searchForm.value, page: pagination.page, page_size: pagination.pageSize } })
    const items = Array.isArray(res.data) ? res.data : (res.data?.data || [])
    tableData.value = items
    pagination.total = res.data?.total || items.length
  } catch (e) { Message.error('加载失败') } finally { loading.value = false }
}

const resetForm = () => Object.assign(form, { id: '', name: '', title: '', org: '', field: '', bio: '', avatar_url: '', status: 'pending', tags: [] })
const handleSearch = () => { pagination.page = 1; loadData() }
const handleReset = () => { searchForm.value = {}; handleSearch() }
const handleAdd = () => { resetForm(); tagsInput.value = ''; dialog.isEdit = false; dialog.visible = true }
const handleEdit = (row) => {
  Object.assign(form, { ...row, tags: row.tags || [] })
  tagsInput.value = (row.tags || []).join(',')
  dialog.isEdit = true; dialog.visible = true
}
const handleSubmit = async () => {
  if (!form.name) { Message.warning('请输入姓名'); return }
  form.tags = tagsInput.value.split(',').map(s => s.trim()).filter(Boolean)
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await axios.put(`/api/v1/admin/experts/${form.id}`, { ...form })
      Message.success('更新成功')
    } else {
      await axios.post('/api/v1/admin/experts', { name: form.name, title: form.title, org: form.org, field: form.field, bio: form.bio, avatar_url: form.avatar_url, tags: form.tags, status: form.status })
      Message.success('创建成功')
    }
    dialog.visible = false; loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { dialog.loading = false }
}
const handleDelete = (row) => {
  Modal.confirm({
    title: '删除专家',
    content: `确定删除专家"${row.name}"吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await axios.delete(`/api/v1/admin/experts/${row.id}`); Message.success('已删除'); loadData() }
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

.avatar-upload { display: inline-block; cursor: pointer; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
