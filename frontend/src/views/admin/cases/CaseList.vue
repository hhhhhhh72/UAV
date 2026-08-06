<template>
  <div class="page">
    <!-- 搜索 + 操作栏 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="分类" class="form-item">
            <a-select v-model="filterParams.categoryId" style="width: 160px" allow-clear @change="onSearchSubmit">
              <a-option label="全部分类" :value="null" />
              <a-option v-for="cat in caseCategories" :key="cat.id" :label="cat.name" :value="Number(cat.id)" />
            </a-select>
          </a-form-item>
          <a-form-item label="关键词" class="form-item">
            <a-input
              v-model="filterParams.keyword"
              placeholder="搜索案例标题..."
              allow-clear
              style="width: 200px"
              @press-enter="onSearchSubmit"
              @clear="onSearchSubmit"
            />
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" @click="createCase"><template #icon><icon-plus /></template>新增案例</a-button>
          <a-button @click="openCategoryManager"><template #icon><icon-settings /></template>管理分类</a-button>
          <a-button @click="refreshAll"><template #icon><icon-refresh /></template>刷新</a-button>
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
        @sorter-change="handleSorterChange"
      >
        <template #cover="{ record }">
          <div class="case-thumb">
            <img v-if="record.coverType !== 'video'" :src="record.cover" :alt="record.title" />
            <video v-else :src="record.cover" muted playsinline preload="metadata" />
          </div>
        </template>
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '未命名案例' }}</span>
          <a-tag v-if="record.subTag" color="orange" size="small" style="margin-left: 6px;">{{ record.subTag }}</a-tag>
        </template>
        <template #category="{ record }">
          <span>{{ getCategoryName(record.categoryId) }}</span>
        </template>
        <template #coverType="{ record }">
          <a-tag :color="record.coverType === 'video' ? 'green' : 'arcoblue'" size="small">
            {{ record.coverType === 'video' ? '视频' : '图片' }}
          </a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="caseStatusColor[record.status]" size="small">{{ caseStatusLabel[record.status] || record.status }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="editCase(record)">编辑</a-button>
            <a-divider direction="vertical" />
            <a-button type="text" status="danger" size="small" @click="onDeleteCase(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无案例数据" />
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

    <!-- 案例编辑弹窗 (保留原有逻辑) -->
    <a-modal
      v-model:visible="showCaseEditPopup"
      :title="currentCase?.id ? '编辑案例' : '新增案例'"
      :width="720"
      :footer="false"
    >
      <template v-if="currentCase">
        <a-form :model="currentCase" layout="horizontal" class="dialog-form">
          <a-divider orientation="left">基本信息</a-divider>
          <a-form-item label="所属分类" required>
            <a-radio-group v-model="currentCase.categoryId" @change="onCategoryChange">
              <a-radio v-for="cat in caseCategories" :key="cat.id" :value="Number(cat.id)">{{ cat.name }}</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item label="标题" required>
            <a-input v-model="currentCase.title" placeholder="请输入标题" allow-clear />
          </a-form-item>
          <a-form-item label="子标签">
            <a-input v-model="currentCase.subTag" placeholder="可选" allow-clear />
          </a-form-item>
          <a-form-item label="服务类型">
            <a-input v-model="currentCase.service" placeholder="如：无人机物流服务" allow-clear />
          </a-form-item>
          <a-form-item label="简介">
            <a-input v-model="currentCase.description" type="textarea" :auto-size="{ minRows: 2, maxRows: 6 }" placeholder="请输入简介" />
          </a-form-item>
          <a-form-item label="地点">
            <a-input v-model="currentCase.location" placeholder="请输入地点" allow-clear />
          </a-form-item>
          <a-form-item label="时间">
            <a-date-picker v-model="currentCase.date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
          </a-form-item>

          <a-divider orientation="left">封面设置</a-divider>
          <a-form-item label="封面类型">
            <a-radio-group v-model="currentCase.coverType">
              <a-radio value="image">图片</a-radio>
              <a-radio value="video">视频</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item label="封面地址">
            <template v-if="currentCase.coverType === 'image'">
              <a-upload
                class="cover-upload"
                :show-file-list="false"
                :custom-request="uploadCover"
                :before-upload="beforeUpload"
                accept="image/*"
              >
                <img v-if="currentCase.cover" :src="currentCase.cover" class="cover-preview" />
                <a-button v-else type="primary">点击上传</a-button>
              </a-upload>
              <a-button v-if="currentCase.cover" size="small" style="margin-top: 8px" @click="currentCase.cover = ''">清除</a-button>
            </template>
            <a-input v-else v-model="currentCase.cover" placeholder="输入视频URL" allow-clear />
          </a-form-item>

          <a-divider orientation="left">审核状态</a-divider>
          <a-form-item label="状态">
            <a-select v-model="currentCase.status" style="width: 200px">
              <a-option label="待审核" value="pending" />
              <a-option label="已发布" value="published" />
              <a-option label="已下架" value="archived" />
            </a-select>
          </a-form-item>

          <a-divider orientation="left">详细内容</a-divider>
          <a-form-item label="详细描述">
            <a-input v-model="currentCase.fullDescription" type="textarea" :auto-size="{ minRows: 4, maxRows: 10 }" placeholder="请输入详细描述" />
          </a-form-item>

          <!-- 亮点标签 -->
          <a-divider orientation="left">项目亮点</a-divider>
          <div v-for="(tag, idx) in currentCase.highlights" :key="idx" class="highlight-row">
            <a-input v-model="currentCase.highlights[idx]" :placeholder="'标签 ' + (idx + 1)" />
            <a-button type="text" status="danger" shape="circle" size="small" @click="currentCase.highlights.splice(idx, 1)">
              <template #icon><icon-delete /></template>
            </a-button>
          </div>
          <a-button type="outline" size="small" @click="currentCase.highlights.push('')"><template #icon><icon-plus /></template>添加标签</a-button>

          <!-- 媒体资源 -->
          <a-divider orientation="left">媒体资源</a-divider>
          <div v-for="(media, idx) in currentCase.media" :key="idx" class="media-card">
            <div class="media-card-head">
              <b>资源 #{{ idx + 1 }}</b>
              <a-button type="text" status="danger" size="small" @click="currentCase.media.splice(idx, 1)">
                <template #icon><icon-delete /></template>
              </a-button>
            </div>
            <a-radio-group v-model="media.type" style="margin-bottom: 8px;">
              <a-radio value="image">图片</a-radio>
              <a-radio value="video">视频</a-radio>
            </a-radio-group>
            <a-input v-model="media.url" placeholder="URL" />
          </div>
          <a-button type="outline" size="small" @click="currentCase.media.push({ type: 'image', url: '' })"><template #icon><icon-plus /></template>添加资源</a-button>
        </a-form>

        <div class="modal-footer">
          <a-space>
            <a-button @click="showCaseEditPopup = false">取消</a-button>
            <a-button v-if="currentCase?.id" status="danger" @click="onDeleteCase(currentCase)">删除案例</a-button>
            <a-button type="primary" @click="onSaveCase">保存</a-button>
          </a-space>
        </div>
      </template>
    </a-modal>

    <!-- 分类管理弹窗 (保留原有逻辑) -->
    <a-modal v-model:visible="showCategoryPopup" title="分类管理" :width="500" :footer="false">
      <div class="cat-head">
        <span>共 {{ caseCategories.length }} 项</span>
        <a-button type="primary" size="small" @click="openAddCategoryDialog"><template #icon><icon-plus /></template>新增分类</a-button>
      </div>

      <a-empty v-if="caseCategories.length === 0" description="暂无分类" />

      <div v-else>
        <div v-for="(cat, idx) in caseCategories" :key="cat.id" class="cat-item">
          <span class="cat-index">#{{ idx + 1 }}</span>
          <a-tag color="arcoblue" size="small">ID: {{ cat.id }}</a-tag>
          <span class="cat-name">{{ cat.name || '-' }}</span>
          <span class="cat-service">{{ cat.service || '-' }}</span>
          <a-button type="text" status="success" size="small" @click="startEditCategory(cat)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="deleteCategory(cat)">删除</a-button>
        </div>
      </div>
    </a-modal>

    <!-- 分类编辑子弹窗 -->
    <a-modal v-model:visible="showCategoryDialog" :title="editingCategory?.id ? '编辑分类' : '新增分类'" :width="400">
      <a-form v-if="editingCategory" :model="editingCategory" layout="horizontal" class="dialog-form">
        <a-form-item label="分类名称" required>
          <a-input v-model="editingCategory.name" placeholder="如：共享无人机" allow-clear />
        </a-form-item>
        <a-form-item label="默认服务">
          <a-input v-model="editingCategory.service" placeholder="如：共享无人机服务" allow-clear />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="showCategoryDialog = false">取消</a-button>
        <a-button type="primary" @click="onSaveCategory">确认</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import axios from '@/utils/http'
import { useListRequest } from '@/hooks/useListRequest'
import { normalizeMediaUrl } from '../composables/useMedia'

// --- 分类数据 ---
const caseCategories = ref([])

const fetchCaseCategories = async () => {
  try {
    const res = await axios.get('/api/case-categories')
    caseCategories.value = Array.isArray(res.data) ? res.data : []
  } catch (error) {
    console.error('获取分类失败', error)
    caseCategories.value = []
  }
}

const getCategoryName = (id) => {
  const cat = caseCategories.value.find(c => Number(c.id) === Number(id))
  return cat ? cat.name : (id ?? '-')
}

const onCategoryChange = (val) => {
  if (!currentCase.value) return
  const cat = caseCategories.value.find(c => Number(c.id) === Number(val))
  if (cat && !currentCase.value.service) {
    currentCase.value.service = cat.service || ''
  }
}

// --- 案例列表 ---
const fetchCases = async (params) => {
  try {
    const res = await axios.get('/api/cases', { params })
    const raw = res.data?.data || res.data || []
    return {
      data: Array.isArray(raw) ? raw.map(c => ({
        ...c,
        cover: normalizeMediaUrl(c.cover),
        media: Array.isArray(c.media) ? c.media.map(m => ({ ...m, url: normalizeMediaUrl(m?.url) })) : c.media
      })) : [],
      total: res.data?.total || (Array.isArray(raw) ? raw.length : 0)
    }
  } catch (error) {
    Message.error('获取案例数据失败')
    return { data: [], total: 0 }
  }
}

const { listData, loading, total, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: computed(() => fetchCases),
  idKey: 'id',
  defaultParams: { categoryId: '' }
})

// a-table 排序（Arco direction → useListRequest 的 order 语义）
const handleSorterChange = (dataIndex, direction) => {
  const order = direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  onSortChange({ prop: dataIndex, order })
}

const columns = [
  { title: '封面', dataIndex: 'cover', slotName: 'cover', width: 90 },
  { title: '标题', dataIndex: 'title', slotName: 'title', minWidth: 180, sortable: true },
  { title: '分类', dataIndex: 'categoryId', slotName: 'category', width: 110 },
  { title: '地点', dataIndex: 'location', width: 120 },
  { title: '时间', dataIndex: 'date', width: 110 },
  { title: '类型', dataIndex: 'coverType', slotName: 'coverType', width: 80 },
  { title: '浏览量', dataIndex: 'views', width: 90, align: 'right', sortable: true },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

const refreshAll = async () => {
  await fetchCaseCategories()
  loadData()
}

// --- 案例编辑 ---
const showCaseEditPopup = ref(false)
const currentCase = ref(null)
const caseStatusColor = { pending: 'orange', published: 'green', archived: 'gray' }
const caseStatusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// 封面图片上传（/api/v1/upload 返回相对 URL）
const uploadCover = async ({ file, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await axios.post('/api/v1/upload', fd)
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    if (currentCase.value) currentCase.value.cover = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error('上传失败')
  }
}

const createCase = async () => {
  if (caseCategories.value.length === 0) await fetchCaseCategories()
  const firstCat = caseCategories.value[0]
  currentCase.value = {
    title: '', description: '', location: '', date: '', fullDescription: '',
    coverType: 'image', cover: '', media: [], highlights: [], status: 'pending',
    categoryId: firstCat ? Number(firstCat.id) : null,
    service: firstCat ? firstCat.service : '', subTag: ''
  }
  showCaseEditPopup.value = true
}

const editCase = async (caseItem) => {
  if (caseCategories.value.length === 0) await fetchCaseCategories()
  currentCase.value = JSON.parse(JSON.stringify(caseItem))
  if (!currentCase.value.media) currentCase.value.media = []
  if (!currentCase.value.highlights) currentCase.value.highlights = []
  if (!currentCase.value.coverType) currentCase.value.coverType = 'image'
  if (currentCase.value.subTag == null) currentCase.value.subTag = ''
  showCaseEditPopup.value = true
}

const onSaveCase = async () => {
  if (!currentCase.value) return
  Message.loading('保存中...', 0)
  try {
    if (currentCase.value.id) {
      await axios.post('/api/cases/update', currentCase.value)
    } else {
      await axios.post('/api/cases/create', currentCase.value)
    }
    Message.clear()
    Message.success('保存成功')
    showCaseEditPopup.value = false
    loadData()
  } catch (error) {
    Message.clear()
    Message.error(error?.response?.data?.message || '保存失败')
  }
}

const onDeleteCase = (caseItem) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这个案例吗？删除后无法恢复。',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await axios.post('/api/cases/delete', { id: caseItem.id })
        Message.success('删除成功')
        if (currentCase.value?.id === caseItem.id) showCaseEditPopup.value = false
        loadData()
      } catch (error) {
        Message.error('删除失败')
      }
    }
  })
}

// --- 分类管理 ---
const showCategoryPopup = ref(false)
const showCategoryDialog = ref(false)
const editingCategory = ref(null)

const openCategoryManager = async () => {
  await fetchCaseCategories()
  showCategoryPopup.value = true
}

const openAddCategoryDialog = () => {
  editingCategory.value = { id: null, name: '', service: '' }
  showCategoryDialog.value = true
}

const startEditCategory = (cat) => {
  editingCategory.value = { id: cat.id, name: cat.name || '', service: cat.service || '' }
  showCategoryDialog.value = true
}

const onSaveCategory = async () => {
  const form = editingCategory.value
  if (!form || !form.name?.trim()) {
    Message.error('分类名称不能为空')
    return
  }
  Message.loading('保存中...', 0)
  try {
    if (form.id == null) {
      await axios.post('/api/case-categories/create', { name: form.name.trim(), service: (form.service || '').trim() })
    } else {
      await axios.post('/api/case-categories/update', { id: form.id, name: form.name.trim(), service: (form.service || '').trim() })
    }
    Message.clear()
    Message.success(form.id == null ? '分类已新增' : '分类已更新')
    showCategoryDialog.value = false
    await fetchCaseCategories()
  } catch (error) {
    Message.clear()
    Message.error(error?.response?.data?.message || '保存失败')
  }
}

const deleteCategory = (cat) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除分类「${cat.name}」吗？若有案例仍归属该分类将无法删除。`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      Message.loading('删除中...', 0)
      try {
        await axios.post('/api/case-categories/delete', { id: cat.id })
        Message.clear()
        Message.success('删除成功')
        await fetchCaseCategories()
      } catch (error) {
        Message.clear()
        Message.error(error?.response?.data?.message || '删除失败')
      }
    }
  })
}

onMounted(refreshAll)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.case-thumb { width: 56px; height: 56px; border-radius: 6px; overflow: hidden; background: #F7F8FA; }
.case-thumb img, .case-thumb video { width: 100%; height: 100%; object-fit: cover; display: block; }
.cell-title { font-weight: 500; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #EEF1F4;
}

.highlight-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.media-card {
  margin-bottom: 12px;
  padding: 10px;
  border: 1px dashed #E0E0E0;
  border-radius: 6px;
}

.media-card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.cat-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.cat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px;
  border: 1px solid #EBEEF5;
  border-radius: 6px;
  margin-bottom: 8px;
}

.cat-index { color: #909399; font-size: 12px; }
.cat-name { flex: 1; }
.cat-service { color: #909399; font-size: 12px; }

.cover-upload { display: inline-block; margin-right: 8px; }
.cover-preview { width: 160px; height: 100px; object-fit: cover; border: 1px solid #ddd; border-radius: 4px; cursor: pointer; }
</style>
